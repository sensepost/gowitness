FROM node:24-bookworm-slim AS frontend

WORKDIR /src/web/ui

COPY web/ui/package*.json ./
RUN npm ci

COPY web/ui/ ./
RUN npm run build

FROM golang:1-bookworm AS build

COPY . /src
WORKDIR /src

COPY --from=frontend /src/web/ui/dist ./web/ui/dist

RUN go install github.com/swaggo/swag/cmd/swag@latest && \
	swag i --exclude ./web/ui --output web/docs && \
	go build -trimpath -ldflags="-s -w \
	-X=github.com/sensepost/gowitness/internal/version.GitHash=$(git rev-parse --short HEAD) \
	-X=github.com/sensepost/gowitness/internal/version.GoBuildEnv=$(go version | cut -d' ' -f 3,4 | sed 's/ /_/g') \
	-X=github.com/sensepost/gowitness/internal/version.GoBuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
	-o gowitness

FROM docker.io/chromedp/headless-shell:stable

COPY --from=build /src/gowitness /usr/local/bin/gowitness

RUN apt-get update && \
	apt-get install -y --no-install-recommends tini ca-certificates && \
	rm -rf /var/lib/apt/lists/*

RUN ln -sf /headless-shell/headless-shell /usr/bin/google-chrome && \
	ln -sf /headless-shell/headless-shell /usr/bin/chromium && \
	ln -sf /headless-shell/headless-shell /usr/bin/chromium-browser

EXPOSE 7171

VOLUME ["/data"]
WORKDIR /data

ENTRYPOINT ["tini", "--", "gowitness"]
