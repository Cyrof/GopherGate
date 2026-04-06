FROM node:24-bookworm-slim AS css-builder
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build:css

FROM golang:1.25 AS go-builder
WORKDIR /src
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates git && rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=css-builder /src/web/static/css/styles.css /src/web/static/css/styles.css
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ui ./cmd/ui

FROM alpine:3.21
RUN apk add --no-cache ca-certificates \
    && adduser -D -H -u 65532 nonroot \
    && mkdir -p /home/nonroot \
    && chown -R 65532:65532 /home/nonroot

ENV HOME=/home/nonroot
COPY --from=go-builder /out/ui /ui
USER 65532:65532
ENTRYPOINT ["/ui"]
