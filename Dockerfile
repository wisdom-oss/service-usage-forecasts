FROM golang:latest AS build-http-server
COPY . /tmp/src
WORKDIR /tmp/src
RUN mkdir -p /tmp/build && go mod download & go build -v -o /tmp/build/app

FROM rocker/tidyverse
COPY --from=build-http-server /tmp/build/app /usage-forecasts
COPY algorithms /algorithms
COPY resources/* /
USER root
RUN apt-get update && apt-get install -y python3.10-full dos2unix python-is-python3 && chmod +x /algorithms -R && find /algorithms -type f -print0 | xargs -0 dos2unix && echo "conversion done"
EXPOSE 8000
ENTRYPOINT ["/usage-forecasts"]