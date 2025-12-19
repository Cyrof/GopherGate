FROM golang:1.24-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/gophergate-wg-agent ./cmd/gophergate-wg-agent

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /

COPY --from=builder /out/gophergate-wg-agent /usr/local/bin/gophergate-wg-agent

RUN addgroup -S gophergate && adduser -S gophergate -G gophergate
USER gophergate

RUN ln -sf /prox/1/fd/1 \
    /home/gophergate/.local/state/gophergate/gophergate-wg-agent/logs/gophergate-wg-agent.log

ENTRYPOINT ["/usr/local/bin/gophergate-wg-agent"]
CMD ["serve"]
