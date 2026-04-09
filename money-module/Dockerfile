# syntax=docker/dockerfile:1
# language which we use
FROM golang:1.22-alpine AS base
WORKDIR /money-module

FROM base AS deps
COPY money-module/go.mod money-module/go.sum ./
RUN go mod download


FROM base AS dev
COPY money-module/ .
EXPOSE 6767


CMD ["go", "run", "main.go"]
