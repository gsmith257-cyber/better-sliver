#!/bin/bash
go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build && \
go-fuzz-build && \
go-fuzz
