FROM alpine:3.8
COPY bin/server /

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

ENTRYPOINT ["/server"]
CMD ["--config", "config.json", "--authenticator", "google.com"]