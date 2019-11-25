FROM alpine
COPY ./bin/app config.json //
ENTRYPOINT [ "/app" ]
