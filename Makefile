TARGET=ariadna

all: clean build

clean:
	rm -rf $(TARGET)

depends:
	dep esure

build:
	go build -v -o  $(TARGET) *.go

fmt:
	go fmt ./...
