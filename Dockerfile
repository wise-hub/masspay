FROM alpine:latest

WORKDIR /app

COPY bin/masspay /app
COPY web /app/web

EXPOSE 9000

CMD ["/app/masspay"]
