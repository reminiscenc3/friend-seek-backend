ARG GO_VERSION=1.27.1
ARG ALPINE_VERSION=3.22

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# modernc.org/sqlite is a pure Go driver, so the binaries stay static.
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate


FROM alpine:${ALPINE_VERSION}

RUN adduser -D -u 10001 app && \
    mkdir -p /data && chown app:app /data

COPY --from=builder /out/api /usr/local/bin/api
COPY --from=builder /out/migrate /usr/local/bin/migrate

USER app
WORKDIR /data

ENV HTTP_ADDRESS=:8081 \
    DB_PATH=/data/friend_seek.db

EXPOSE 8081

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://127.0.0.1:8081/health || exit 1

CMD ["api"]
