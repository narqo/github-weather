FROM golang:1.22-alpine as builder
RUN apk add --update --no-cache build-base ca-certificates
WORKDIR /go/src/github-weather
COPY . /go/src/github-weather
RUN CGO_ENABLED=0 go build -a -ldflags '-s -w -extldflags "-static"' -o bin/github-weather .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github-weather/bin/github-weather /bin/github-weather
ENTRYPOINT ["/bin/github-weather"]
