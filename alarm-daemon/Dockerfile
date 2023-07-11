# syntax=docker/dockerfile:1

FROM golang:1.20-alpine3.18 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/alarm-daemon

# ---

FROM gcr.io/distroless/static-debian11:nonroot

WORKDIR /

COPY --from=builder /build/alarm-daemon /alarm-daemon

ENTRYPOINT ["/alarm-daemon"]
