# syntax = docker/dockerfile:1
FROM golang:1.20-alpine3.18 AS builder

ENV CGO_ENABLED=1
ENV APP_VERSION="dev"

RUN apk add --no-cache \
    build-base \
    libjpeg-turbo-dev \
    vips-dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build \
    -trimpath \
    -buildvcs=false \
    -ldflags="-linkmode 'external' -X 'main.Version=${APP_VERSION}'" \
    -mod=readonly \
    -o /app \
    ./cmd/app

CMD ["/app"]

FROM alpine:3.18

# needed as not a static binary
RUN apk add --no-cache \
    libjpeg-turbo \
    vips

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app /app

CMD ["/app"]
