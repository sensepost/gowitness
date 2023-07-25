FROM golang:1-bullseye as build

LABEL maintainer="Leon Jacobs <leonja511@gmail.com>"

COPY . /src

WORKDIR /src
RUN make docker

# final image
# https://github.com/chromedp/docker-headless-shell#using-as-a-base-image
FROM chromedp/headless-shell:latest

RUN export DEBIAN_FRONTEND=noninteractive \
  && apt-get update \
  && apt-get install -y --no-install-recommends \
  dumb-init \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

COPY --from=build /src/gowitness /usr/local/bin

EXPOSE 7171

VOLUME ["/data"]
WORKDIR /data

ENTRYPOINT ["dumb-init", "--"]
