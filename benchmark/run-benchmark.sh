#!/bin/sh
wrk -t4 -c100 -d60s -s /random-header.lua "$@"