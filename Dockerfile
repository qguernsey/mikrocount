# build stage
FROM golang:1.18-alpine AS builder
WORKDIR /go/src/github.com/atajsic/mikrocount
COPY ./ ./
RUN go build -o mikrocount

# final stage
FROM alpine:latest
WORKDIR /app/
COPY --from=builder /go/src/github.com/atajsic/mikrocount/mikrocount /app/mikrocount
ENTRYPOINT ["./mikrocount"]
