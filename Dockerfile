FROM alpine

RUN apk add --no-cache ca-certificates

COPY github-weather /bin/github-weather

ENTRYPOINT ["/bin/github-weather"]