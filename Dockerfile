#TODO: Add build make command with ldflags for version
FROM golang:alpine AS build

WORKDIR /app

COPY . .

RUN go build -o mantis cmd/*.go


FROM alpine AS app

WORKDIR /app

COPY --from=build /app/mantis .

CMD [ "/app/mantis" ]