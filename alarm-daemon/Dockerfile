# syntax=docker/dockerfile:1

FROM golang:1.20-alpine3.18 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/alarm-daemon

# ---

FROM alpine:3.18 as certs
RUN apk add --no-cache ca-certificates

# ---

FROM scratch

WORKDIR /

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /build/alarm-daemon /alarm-daemon

ENTRYPOINT ["/alarm-daemon"]
