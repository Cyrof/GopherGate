FROM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" \
    -o /out/gophergate-wg-agent ./cmd/gophergate-wg-agent

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /

COPY --from=builder /out/gophergate-wg-agent /usr/local/bin/gophergate-wg-agent

ENTRYPOINT ["/usr/local/bin/gophergate-wg-agent"]
CMD ["serve"]
