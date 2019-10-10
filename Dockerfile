FROM golang:alpine as build

LABEL maintainer="Leon Jacobs <leonja511@gmail.com>"

COPY . /src

WORKDIR /src
RUN go build -o gowitness

# final image
FROM zenika/alpine-chrome:latest

COPY --from=build /src/gowitness /

VOLUME ["/screenshots"]
WORKDIR /screenshots

# https://github.com/Zenika/alpine-chrome#-with---no-sandbox
ENTRYPOINT ["/gowitness", "--chrome-arg=\"-no-sandbox\""]
