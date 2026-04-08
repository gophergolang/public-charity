FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o auth ./cmd/auth

FROM alpine:3.21
RUN apk add --no-cache ca-certificates

# Install Litestream for continuous SQLite → S3 replication.
ARG LITESTREAM_VERSION=v0.3.13
ADD https://github.com/benbjohnson/litestream/releases/download/${LITESTREAM_VERSION}/litestream-${LITESTREAM_VERSION}-linux-amd64.tar.gz /tmp/litestream.tar.gz
RUN tar -xzf /tmp/litestream.tar.gz -C /usr/local/bin && rm /tmp/litestream.tar.gz

COPY --from=builder /app/auth /usr/local/bin/auth
COPY litestream.yml /etc/litestream.yml
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
