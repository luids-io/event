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
RUN adduser -D -g '' luevent && \
	mkdir -p /var/lib/luevent && \
	chown luevent /var/lib/luevent

COPY --from=build-env /app/bin/* /bin/
COPY --from=build-env /app/configs/docker/ /etc/luevent/

USER luevent

EXPOSE 5851
VOLUME [ "/etc/luevent", "/var/lib/luevent" ]
CMD [ "/bin/eventproc" ]
