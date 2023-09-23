#TODO: Add build make command with ldflags for version
FROM alpine

RUN mkdir /app

COPY ./dist/mantis /app

CMD [ "/app/mantis" ]