# reflect: gRPC reflection CLI [![CI/CD](https://github.com/juliaogris/reflect/workflows/CI/CD/badge.svg?branch=master)](https://github.com/squareup/gurl/actions?query=workflow%3ACI%2FCD+branch%3Amaster)

reflect is a gRPC reflection client for the command line.

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

## Pre-requistes

* GNU Make 3.81
* go 1.16.3
* golangci-lint 1.37.0

	make
	make help  # for more options
