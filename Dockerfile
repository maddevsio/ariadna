FROM golang
RUN go get github.com/gen1us2k/ariadna
WORKDIR /go/src/github.com/gen1us2k/ariadna
RUN go get
RUN go build
RUN go install

EXPOSE 8080
CMD /go/bin/ariadna http
