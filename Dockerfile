FROM --platform=$BUILDPLATFORM docker.io/golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags='-w -s' \
    -o inventory-transaction-service .

FROM --platform=$TARGETPLATFORM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/inventory-transaction-service /app/inventory-transaction-service

USER nonroot:nonroot

EXPOSE 14330

CMD ["/app/inventory-transaction-service"]
