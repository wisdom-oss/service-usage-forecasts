FROM golang:latest AS build-http-server
COPY . /tmp/src
WORKDIR /tmp/src
RUN mkdir -p /tmp/build && go mod download & go build -v -o /tmp/build/app

FROM python:3.10-slim AS python-prep
USER root
RUN pip install --no-cache-dir pandas numpy scikit-learn orjson prophet

FROM alpine:latest AS algorithm-converter
COPY algorithms /algorithms
RUN apk add --no-cache dos2unix
RUN find /algorithms -type f -print0 | xargs -0 dos2unix

FROM python:3.10-slim
COPY --from=algorithm-converter --chmod=777 /algorithms /algorithms
COPY --from=build-http-server /tmp/build/app /usage-forecasts
COPY --from=python-prep /usr/local/lib/python3.10/site-packages /usr/local/lib/python3.10/site-packages
EXPOSE 8000
ENTRYPOINT ["/usage-forecasts"]