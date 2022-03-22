#TODO: Add build make command with ldflags for version
FROM golang:alpine AS build

WORKDIR /app

COPY . .

RUN go build -o wirego cmd/*.go


FROM alpine AS app

WORKDIR /app

COPY --from=build /app/wirego .

CMD [ "/app/wirego" ]