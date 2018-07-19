FROM debian:latest
EXPOSE 8080
COPY gwitask /
ENTRYPOINT ["/gwitask"]
