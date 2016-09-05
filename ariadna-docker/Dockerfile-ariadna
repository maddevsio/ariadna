FROM golang
RUN go get github.com/maddevsio/ariadna
WORKDIR /go/src/github.com/maddevsio/ariadna
RUN go get -u && go build && go install

EXPOSE 8080
CMD /go/bin/ariadna http
