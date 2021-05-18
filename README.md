# reflect [![CI/CD](https://github.com/juliaogris/reflect/actions/workflows/cicd.yaml/badge.svg?branch=master)](https://github.com/juliaogris/reflect/actions/workflows/cicd.yaml?query=branch%3Amaster)
reflect is a gRPC reflection CLI. See `reflect --help` for details.

## Install

Download binary from [releases](releases), untar and put on path, or
install from source with

	go install github.com/juliaogris/reflect@latest

## Development prerequisites

* GNU Make 3.81
* go 1.16.3
* golangci-lint 1.37.0

To build and test run

	make

for more options see

	make help  # for more options

## Test drive

Setup

	make install
	make run # starts test server on localhost:9090
	export GRPC_ADDRESS=localhost:9090
	export GRPC_PLAINTTEXT=true

and call reflect

	reflect services
	reflect filename echo3.Echo
	reflect filename echo3/echo3.proto
	reflect extensions google.protobuf.MethodOptions
	reflect extension google.protobuf.MethodOptions 72295728
