FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git nodejs npm

COPY ./go.mod ./go.sum ./
RUN go mod download -x

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

RUN npm run build:css

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/ui ./cmd/ui

# runtime 
FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
    && adduser -D -H -u 65532 nonroot \
    && mkdir -p /home/nonroot \
    && chown -R 65532:65532 /home/nonroot

ENV HOME=/home/nonroot

COPY --from=builder /out/ui /ui

USER 65532:65532
ENTRYPOINT ["/ui"]
