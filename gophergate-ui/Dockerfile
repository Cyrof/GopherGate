FROM golang:1.24-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY ./go.mod ./go.sum ./gophergate-ui/
WORKDIR /src/gophergate-ui

RUN go mod download -x

WORKDIR /src
COPY . .

WORKDIR /src/gophergate-ui
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/ui ./cmd/ui

FROM gcr.io/distroless/static-debian12

COPY --from=builder /out/ui /ui

EXPOSE 3000
ENTRYPOINT ["/ui"]