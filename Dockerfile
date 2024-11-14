FROM golang:latest AS build-http-server
WORKDIR /src
COPY go.* .
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build go build -o /out/service

FROM python:3.10-slim AS python-prep
USER root
COPY --link requirements.txt .
RUN --mount=type=cache,target=/root/.cache \
     pip install -r requirements.txt

FROM alpine:latest AS algorithm-converter
COPY --link algorithms /algorithms
RUN apk add --no-cache dos2unix
RUN find /algorithms -type f -print0 | xargs -0 dos2unix

FROM python:3.10-slim
COPY --link ./resources /resources
COPY --link --from=algorithm-converter --chmod=777 /algorithms /algorithms
COPY --link --from=build-http-server /out/service /usage-forecasts
COPY --from=python-prep /usr/local/lib/python3.10/site-packages /usr/local/lib/python3.10/site-packages
EXPOSE 8000
ENTRYPOINT ["/usage-forecasts"]