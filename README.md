# reflect [![CI/CD](https://github.com/juliaogris/reflect/actions/workflows/cicd.yaml/badge.svg?branch=master)](https://github.com/juliaogris/reflect/actions/workflows/cicd.yaml?query=branch%3Amaster)
reflect is a gRPC reflection CLI.

Use as

	export GRPC_ADDRESS=localhost:9090
	export GRPC_PLAINTTEXT=true

	reflect filename routeguide.proto
	reflect symbol RouteGuide
	reflect extension google.protobuf.MethodOptions 72295728 # HttpRule

	reflect services  # List all services
	reflect extensions google.protobuf.MethodOptions # List all extension numbers

	reflect --help


## Install

Download binary from [releases](releases), untar and put on path, or
install from source with

	go install github.com/juliaogris/reflect@latest

## Development

### Pre-requistes

* GNU Make 3.81
* go 1.16.3
* golangci-lint 1.37.0

To build and test run

	make

for more options see

	make help  # for more options
