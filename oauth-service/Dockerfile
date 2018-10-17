FROM scratch
COPY bin/server /

ENTRYPOINT ["/server"]
CMD ["--config", "config.json", "--authenticator", "google.com"]