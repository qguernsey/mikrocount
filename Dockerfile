# build stage
FROM golang AS builder
RUN apk add --update ca-certificates
RUN apk add --no-cache git && go get -u github.com/golang/dep/...
WORKDIR /go/src/github.com/atajsic/mikrocount
ADD . ./
RUN dep ensure
RUN go build -o mikrocount

# final stage
FROM alpine:latest
WORKDIR /app/
COPY --from=builder /go/src/github.com/atajsic/mikrocount/mikrocount /app/mikrocount
ENTRYPOINT ["./mikrocount"]
