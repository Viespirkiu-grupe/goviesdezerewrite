FROM golang:1.25.1 AS builder

WORKDIR /temp

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY internal internal

RUN CGO_ENABLED=0 go build -o app .

FROM alpine:latest

RUN apk add --no-cache p7zip

COPY --from=builder /temp/app /app

WORKDIR /work

ENTRYPOINT ["/app"]
