FROM golang:1.17.2-alpine3.13 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o kv-svc

FROM alpine
COPY --from=builder /app/kv-svc /kv-svc
EXPOSE 10000
ENTRYPOINT ["/kv-svc"]
