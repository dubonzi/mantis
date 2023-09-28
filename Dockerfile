FROM alpine

RUN mkdir /app

WORKDIR /app

COPY ./dist/mantis .

CMD [ "./mantis" ]