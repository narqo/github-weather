FROM alpine as build

RUN apk add --update --no-cache ca-certificates

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY github-weather /bin/github-weather
ENTRYPOINT ["/bin/github-weather"]
