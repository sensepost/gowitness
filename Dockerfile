FROM golang:1-bookworm AS build

RUN apt-get update && \
	apt-get install -y npm

ADD . /src
WORKDIR /src

RUN cd web/ui && \
	rm -Rf node_modules && \
	npm i && \
	npm run build && \
	cd ../.. && \
	go build -ldflags="-s -w" -o gowitness

FROM ghcr.io/go-rod/rod

COPY --from=build /src/gowitness /usr/local/bin/gowitness

EXPOSE 7171

VOLUME ["/data"]
WORKDIR /data

ENTRYPOINT ["dumb-init", "--"]