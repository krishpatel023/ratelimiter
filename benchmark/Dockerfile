FROM williamyeh/wrk
COPY random-header.lua /random-header.lua
COPY run-benchmark.sh /run-benchmark.sh
RUN chmod +x /run-benchmark.sh
ENTRYPOINT ["/run-benchmark.sh"]