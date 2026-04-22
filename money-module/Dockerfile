# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS dingo-builder
RUN go install github.com/MadAppGang/dingo/cmd/dingo@latest

FROM golang:1.24-alpine AS base
WORKDIR /money-module
RUN apk add --no-cache git 

FROM base AS deps
COPY go.mod go.sum* ./
RUN go mod download
COPY --from=dingo-builder /go/bin/dingo /usr/local/bin/dingo

FROM deps AS dev
COPY . .
EXPOSE 6767
CMD ["dingo", "run", "main.go"]

FROM deps AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux dingo build -o /money-module .

FROM alpine:latest AS prod
WORKDIR /root/
COPY --from=builder /money-module  .
EXPOSE 6767
CMD ["./money-module "]

