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


.PHONY: wkr-test wkr-build wkr-run
wkr-test:
	@echo "Running wrk benchmark"
	make wkr-build
	make wkr-run

wkr-build:
	cd benchmark && docker build -t wrk-benchmark .

wkr-run:
	cd benchmark && docker run --rm --add-host=host.docker.internal:host-gateway wrk-benchmark http://host.docker.internal:8080/