# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS base
WORKDIR /money-module

FROM base AS deps
COPY go.mod go.sum* ./
RUN go mod download


FROM base AS dev
COPY . .
EXPOSE 6767


CMD ["go", "run", "main.go"]
