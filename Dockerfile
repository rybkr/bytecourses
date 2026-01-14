ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app ./cmd/server
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM debian:bookworm

RUN apt-get update\
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /run-app /usr/local/bin/run-app
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /usr/src/app/migrations ./migrations
COPY --from=builder /usr/src/app/web/static ./web/static
COPY --from=builder /usr/src/app/web/templates ./web/templates

ENTRYPOINT ["run-app"]
CMD []
