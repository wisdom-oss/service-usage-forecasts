FROM golang:latest AS build-http-server
COPY . /tmp/src
WORKDIR /tmp/src
RUN mkdir -p /tmp/build && go mod download & go build -v -o /tmp/build/app

FROM python:3.10-bookworm
COPY algorithms /algorithms
COPY resources/* /
USER root
RUN chmod +x /algorithms -R && \
    find /algorithms -type f -print0 | xargs -0 dos2unix && \
    echo "conversion done" && \
    pip install pandas numpy scikit-learn orjson prophet
COPY --from=build-http-server /tmp/build/app /usage-forecasts
EXPOSE 8000
ENTRYPOINT ["/usage-forecasts"]