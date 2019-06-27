FROM golang:1.12.5-stretch

RUN mkdir -p $GOPATH/src/github.com/maddevsio/ariadna
WORKDIR $GOPATH/src/github.com/maddevsio/ariadna

COPY . .
RUN set -x \
    && go install -v \
    && cp index.json.example index.json \
    && cp .env.example .env

EXPOSE 8080
CMD ["ariadna", "http"]
