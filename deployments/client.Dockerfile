FROM golang:alpine as build

RUN apk add ca-certificates git gcc musl-dev mingw-w64-gcc

WORKDIR /opt

COPY go.mod go.sum ./
RUN  go mod download

COPY cmd/client/      cmd/client/
COPY config/          config/
COPY internal/core/   internal/core/
COPY internal/pow/    internal/pow/
COPY internal/client  internal/client

RUN go test -cover -race -v ./...

RUN cd /opt/cmd/client && \
    go build -o /srv/client


FROM alpine:latest

COPY --from=build /srv /srv
COPY config/ /srv/config/


WORKDIR /srv
CMD /srv/client

