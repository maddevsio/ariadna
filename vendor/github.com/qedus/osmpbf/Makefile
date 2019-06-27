export GORACE := halt_on_error=1

all: cover

init:
	go get -u golang.org/x/perf/cmd/benchstat

race:
	go install -v -race
	go test -v -race

cover:
	go install -v
	go test -v -coverprofile=profile.cov

bench:
	env OSMPBF_BENCHMARK_BUFFER=1048576  go test -v -run=NONE -bench=. -benchmem -benchtime=10s -count=5 | tee 01.txt
	env OSMPBF_BENCHMARK_BUFFER=33554432 go test -v -run=NONE -bench=. -benchmem -benchtime=10s -count=5 | tee 32.txt
	benchstat 01.txt 32.txt
