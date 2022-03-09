# syntax=docker/dockerfile:1

## BUILD
FROM golang:1.16-buster AS build

WORKDIR /app
COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o /FileMan

## DEPLOY
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /FileMan /FileMan

EXPOSE 8080

ENV GIN_MODE release

ENTRYPOINT [ "/FileMan" ]
