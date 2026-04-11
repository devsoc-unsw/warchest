# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder
WORKDIR /backend

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

FROM alpine:latest

# (Optional) Add ca-certificates if your app makes outbound HTTPS calls
RUN apk --no-cache add ca-certificates

WORKDIR /

COPY --from=builder /server /server

EXPOSE 6767

CMD ["/server"]
