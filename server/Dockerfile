FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/arcana-server ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S arcana \
    && adduser -S -G arcana arcana
COPY --from=build /out/arcana-server /usr/local/bin/arcana-server
USER arcana
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O - "http://127.0.0.1:${PORT:-8080}/ready" >/dev/null || exit 1
ENTRYPOINT ["/usr/local/bin/arcana-server"]
