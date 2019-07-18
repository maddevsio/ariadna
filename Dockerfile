# build stage
FROM golang:alpine AS build-env

RUN echo "@edge http://nl.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories  && apk --no-cache add ca-certificates dumb-init@edge openssl curl git
ADD . /src
RUN cd /src && go build -o ariadna

FROM alpine
RUN echo "@edge http://nl.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories  && apk --no-cache add ca-certificates dumb-init@edge openssl
COPY --from=build-env /src/ariadna /ariadna
COPY ariadna.yml /ariadna.yml
ENTRYPOINT /ariadna
