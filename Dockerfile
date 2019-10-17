FROM golang:alpine as build

LABEL maintainer="Leon Jacobs <leonja511@gmail.com>"

RUN apk --no-cache add make git

COPY . /src

WORKDIR /src
RUN make docker

# final image
FROM zenika/alpine-chrome:latest

COPY --from=build /src/gowitness /

VOLUME ["/screenshots"]
WORKDIR /screenshots

# https://github.com/Zenika/alpine-chrome#-with---no-sandbox
ENTRYPOINT ["/gowitness", "--chrome-arg=\"-no-sandbox\""]
