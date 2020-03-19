FROM golang:alpine as build-env
ARG arch=amd64

# Install git and certificates
RUN apk update && apk add --no-cache git make ca-certificates && update-ca-certificates

WORKDIR /app
## dependences
COPY go.mod .
COPY go.sum .
RUN go mod download

## build
COPY . .
RUN make binaries SYSTEM="GOOS=linux GOARCH=${arch}"

## create docker
FROM alpine

LABEL maintainer="Luis Guillén Civera <luisguillenc@gmail.com>"

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

# create user for service
RUN adduser -D -g 'luids' luevent && \
	mkdir -p /var/lib/luids/event && \
	chown luids:luevent /var/lib/luids/event

COPY --from=build-env /app/bin/* /bin/
COPY --from=build-env /app/configs/docker/services.json /etc/luids/
COPY --from=build-env /app/configs/docker/eventproc/* /etc/luids/event/

USER luevent

EXPOSE 5851
VOLUME [ "/etc/luids", "/var/lib/luids/event" ]
CMD [ "/bin/eventproc", "--config", "/etc/luids/event/eventproc.toml" ]
