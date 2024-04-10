#!/usr/bin/python3
# -*- coding: utf-8 -*-

import argparse
import datetime
import json

import pandas
import prophet
import sklearn.metrics

parameters = {
    "size": 30,
    "interval_width": 0.80,
    "changepoint_prior_scale": 0.05,
    "changepoint_range": 0.8,
    "daily_seasonality": False,
    "weekly_seasonality": False,
    "yearly_seasonality": False,
}


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("data_file", default="")
    parser.add_argument("output_file", default="")
    parser.add_argument("parameter_file", default="")

    args = parser.parse_args()

    # load the parameters
    try:
        print(args.parameter_file)
        with open(args.parameter_file) as f:
            ext_parameters: dict = json.load(f)
            parameters = parameters | ext_parameters
    except Exception as e:
        print(e)
        print("using default parameters")

    # create a dataframe from the input file
    df = pandas.read_json(args.data_file,
                          dtype={'municipal': str, 'usageType': str, 'date': datetime.datetime, 'amount': float})
    df.rename(columns={'date': 'ds', 'amount': 'y'}, inplace=True)
    # Remove the localization of the timestamps
    df['ds'] = df['ds'].dt.tz_localize(None)

    # Group the usages by municipals and by the usage types
    municipals = df.groupby(df.municipal)

    return_objects = []

    meta = {
        "r-scores": {},
        "real-data-until": {}
    }

    for municipal, df in municipals:
        m = prophet.Prophet(weekly_seasonality=False, daily_seasonality=False,
                            yearly_seasonality=True)
        yearly_usages: pandas.Series = df.groupby(df.ds.dt.year, as_index=True, group_keys=False)['y'].sum()
        x_axis = []
        y_axis = []
        df = pandas.DataFrame(yearly_usages.items())
        df.rename(columns={0: "ds", 1: "y"}, inplace=True)
        for year, usage in yearly_usages.items():
            x_axis.append(year)
            y_axis.append(usage)
            return_objects.append({
                "label": municipal,
                "x": year,
                "y": usage,
                "uncertainty": [0, 0]
            })
        ds = pandas.to_datetime(x_axis, format="%Y")
        df['ds'] = ds
        m.fit(df)
        meta["real-data-until"][str(municipal)] = x_axis[-1]
        future = m.make_future_dataframe(periods=parameters["size"], freq="Y")
        forecast = m.predict(future)
        forecast = forecast[['ds', 'yhat', 'yhat_lower', 'yhat_upper']].copy()
        predicted_references = []

        for idx, row in forecast.iterrows():
            if len(predicted_references) < len(y_axis):
                predicted_references.append(row['yhat'])
            else:
                if idx == len(y_axis):
                    continue
                return_objects.append({
                    "label": municipal,
                    "x": row['ds'].year,
                    "y": row['yhat'],
                    "uncertainty": [row['yhat_lower'], row['yhat_upper']]
                })

        r_square = sklearn.metrics.r2_score(y_axis, predicted_references)
        meta["r-scores"][int(municipal)] = r_square

    with open(args.output_file, 'wt') as f:
        output_object = {
            "meta": meta,
            "data": return_objects
        }
        json.dump(output_object, f, indent=4)
