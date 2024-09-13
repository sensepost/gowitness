FROM golang:1-bookworm AS build

RUN apt-get update && \
	apt-get install -y npm

ADD . /src
WORKDIR /src

RUN cd web/ui && \
	rm -Rf node_modules && \
	npm i && \
	npm run build && \
	cd ../..
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
	swag i --exclude ./web/ui --output web/docs && \
	go build -trimpath -ldflags="-s -w \
	-X=github.com/sensepost/gowitness/internal/version.GitHash=$(git rev-parse --short HEAD) \
	-X=github.com/sensepost/gowitness/internal/version.GoBuildEnv=$(go version | cut -d' ' -f 3,4 | sed 's/ /_/g') \
	-X=github.com/sensepost/gowitness/internal/version.GoBuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
	-o gowitness

FROM ghcr.io/go-rod/rod

COPY --from=build /src/gowitness /usr/local/bin/gowitness

EXPOSE 7171

VOLUME ["/data"]
WORKDIR /data

ENTRYPOINT ["dumb-init", "--"]