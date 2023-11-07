#!/usr/bin/python3
# -*- coding: utf-8 -*-
import argparse
import datetime
import json

import pandas
import numpy
import sklearn.metrics

description = """
This is an example on how to handle the input and output for algorithms and the
data pulled from the databases 
"""

parameters = {
    "size": 30
}

class print_description(argparse.Action):
    def __init__(self, option_strings, dest, **kwargs):
        return super().__init__(option_strings, dest, nargs=0, default=argparse.SUPPRESS, **kwargs)

    def __call__(self, *args, **kwargs):
        print(description.strip())
        parser.exit()
        exit(0)


class return_arguments(argparse.Action):
    def __init__(self, option_strings, dest, **kwargs):
        return super().__init__(option_strings, dest, nargs=0, default=argparse.SUPPRESS, **kwargs)

    def __call__(self, *args, **kwargs):
        print(json.dumps(parameters))
        parser.exit()
        exit(0)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--parameters", help="return parameters of forecast algorithm", action=return_arguments)
    parser.add_argument("--description", help="return parameters of forecast algorithm", action=print_description)
    parser.add_argument("data_file", default="")
    parser.add_argument("output_file", default="")
    parser.add_argument("parameter_file", default="")

    args = parser.parse_args()

    # load the parameters
    try:
        with open(args.parameter_file) as f:
            ext_parameters = json.load(f)
            parameters = ext_parameters
            f.close()
    except:
        print("using default parameters")

    # create a dataframe from the input file
    df = pandas.read_json(args.data_file, dtype={'municipal': str, 'usageType': str, 'date': datetime.datetime, 'amount': float})
    municipals = df.groupby(df.municipal)
    usage_types = df.groupby(df.usageType)

    return_objects = []

    meta = {
        "curves": {},
        "r-scores": {},
        "real-data-until": {}
    }

    for municipal, df in municipals:
        yearly_usages: pandas.Series = df.groupby(df.date.dt.year)['amount'].sum()
        x_axis = []
        y_axis = []
        for year, usage in yearly_usages.items():
            x_axis.append(year)
            y_axis.append(usage)
            return_objects.append({
                "label": f"{int(municipal)}",
                "x": year,
                "y": usage
            })

        prediction_x_axis = numpy.linspace(start=x_axis[0], stop=x_axis[-1] + parameters["size"], num=len(y_axis) + parameters["size"], dtype=int)
        curve = numpy.polynomial.Polynomial.fit(x_axis, y_axis, deg=3)
        prediction_y_axis = curve(prediction_x_axis).tolist()
        reference_values = prediction_y_axis[:len(x_axis)]
        forecasted_values = prediction_y_axis[len(x_axis):]

        r_square = sklearn.metrics.r2_score(y_axis, reference_values)
        meta["curves"][int(municipal)] = str(curve)
        meta["r-scores"][int(municipal)] = r_square
        meta["real-data-until"][int(municipal)] = str(x_axis[-1])
        for year in prediction_x_axis:
            idx = int(year) - int(prediction_x_axis[0])
            return_objects.append({
                "label": f"{int(municipal)}",
                "x": f"{int(year)}",
                "y": float(prediction_y_axis[idx])
            })

    with open(args.output_file, 'wt') as f:
        output_object = {
            "meta": meta,
            "data": return_objects
        }
        json.dump(output_object, f, indent=4)

