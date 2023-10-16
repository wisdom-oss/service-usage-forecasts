<div align="center">
<img height="150px" src="https://raw.githubusercontent.com/wisdom-oss/brand/main/svg/standalone_color.svg">
<h1>Usage Forecasts</h1>
<h3>usage-forecasts</h3>
<p>⚙️ A microservice for water usage forecasts with support for on-demand and preconfigured forecasting algorithms</p>

<img src="https://img.shields.io/github/go-mod/go-version/wisdom-oss/service-usage-forecasts?style=for-the-badge">
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
>[!NOTE]
> The storage and usage of custom forecasts is currently under active
> development and may not be usable or show unexpected behaviour.
 
### Supported Languages
| Language | Version |
| -------- | ------- |
| Python | v3.10 |
| R | v4.3.1 |

### On-demand
>[!IMPORTANT]
> To support on-demand forecasts, the microservice needs access to the Docker 
> Host.
> Read more [here](docs/on-demand-forecasts.md)

To allow a dynamic creation and adaptation of forecasting algorithms, the 
service allows forecast algorithms written in R or Python.
When receiving an on-demand custom forecast request, the script sent in the
request body will be parsed and put into a container. This container will then
be built with the needed requirements and will then be executed as a one-off
service.

Read more [here](docs/on-demand-forecasts.md)

### Preloaded
To allow the usage of preconfigured forecasts in addition to the already
pre-built algorithms, users may use the upload endpoint in this microservice
as it is documented.