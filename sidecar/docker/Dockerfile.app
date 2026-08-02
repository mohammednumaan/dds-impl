ARG GO_VERSION=1.26.2

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

COPY go.mod ./
COPY cmd/app_server ./cmd/app_server
RUN go build -o /out/app-server ./cmd/app_server

FROM alpine:3.20
WORKDIR /app

COPY --from=build /out/app-server /usr/local/bin/app-server

ENTRYPOINT ["app-server"]
