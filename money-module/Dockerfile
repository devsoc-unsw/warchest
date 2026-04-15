# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS base
WORKDIR /money-module
RUN apk add --no-cache git 

FROM base AS deps
COPY go.mod go.sum* ./
RUN go mod download
RUN git clone https://github.com/MadAppGang/dingo.git /tmp/dingo && \
    cd /tmp/dingo && \
    go build -o /usr/local/bin/dingo ./cmd/dingo

FROM deps AS dev
COPY . .
EXPOSE 6767
CMD ["dingo", "run", "main.go"]

FROM deps as builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux dingo build -o /money-module .

FROM alpine:latest AS prod
WORKDIR /root/
COPY --from=builder /money-module  .
EXPOSE 6767
CMD ["./money-module "]

