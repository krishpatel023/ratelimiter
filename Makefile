.PHONY: build run stop clean benchmark

build:
	@echo "Building rate-limiter"
	go build
	@echo "Rate-limiter built successfully"

run: build
	@echo "Running rate-limiter"
	./ratelimiter
	@echo "Rate-limiter stopped"

benchmark:
	@echo "Running benchmark"
	go test -v -count=1 -run=TestBenchmark ./benchmark
	@echo "Benchmark completed"
