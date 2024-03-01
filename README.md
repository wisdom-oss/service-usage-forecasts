<div align="center">
<img height="150px" src="https://raw.githubusercontent.com/wisdom-oss/brand/main/svg/standalone_color.svg">
<h1>Usage Forecasts</h1>
<h3>usage-forecasts</h3>
<p>⚙️ A microservice for water usage forecasts with support for on-demand and preconfigured forecasting algorithms</p>

<img src="https://img.shields.io/github/go-mod/go-version/wisdom-oss/service-usage-forecasts?style=for-the-badge">
<img alt="GitHub Workflow Status (with event)" src="https://img.shields.io/github/actions/workflow/status/wisdom-oss/service-usage-forecasts/docker.yaml?style=for-the-badge&label=Docker%20Image%20Build">

</div>

> [!NOTE]
> This microservice replaces the two-part microservice to minimize the used
> technologies in the WISdoM platform. All features present in the two-part
> microservice will be ported.
>
> The only feature that will not be ported is the communication using AMQP since
> this technology will be discontinued in the WISdoM platform.

## About

This microservice allows users to use one of the following pre-built algorithms
to generate water usage forecasts:

- Linear Regression
- Polynomial Regression (up to the 5th degree)
- Logarithmic Regression

## Custom Forecasts

> [!NOTE]
> The storage and usage of custom forecasts is currently under active
> development and may not be usable or show unexpected behaviour.

### Supported Languages

> [!NOTE]
> The addition of more languages is planned but for the time being only python
> is supported by default.

| Language | Version |
|----------|---------|
| Python   | v3.10   |

### On-demand
> [!IMPORTANT]
> The on-demand forecasts are currently a WIP since there are still some issues
> with script isolation.
> Therefore, on-demand forecasts will not be available at this time.

The service accepts custom on-demand written forecasting models and allows a
fast testing and validation of new forecasting models.
These on-demand forecasts are executed in a new docker container.

### Preloaded

To allow the usage of preconfigured forecasts in addition to the already
pre-built algorithms, users may use the upload endpoint in this microservice
as it is documented.