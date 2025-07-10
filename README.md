# Purpose

dobby-service is service which is a developer helper.

# Build

This service uses the `make` tool to manage build tasks. Some key tasks include the following. For the complete reference check [Makefile](Makefile)

```
make compile            ## Compiles the service
make test               ## Runs the unit tests
make build              ## Builds the binary in the output directiory "out"
make all                ## Builds the binary & runs tests
make build-run-server   ## Runs the server
```

## Dependency Injection

Uses [Wire](https://github.com/google/wire/) framework to manage dependencies. All the dependencies are managed by [di.go](./cmd/dobby-service/di.go) and the wire generated file is at `cmd/dobby-service/wire_gen.go`.

If you are introducing new dependencies, use the `make gen-wire-deps` target to generate the above file drom `di.go`

Ensure you commit both the original file, and the wire_gen files to the repository.

# Frameworks & Libraries used

| Framework / Tool | Purpose |
|---|---|
| [Gin](https://github.com/gin-gonic) | Web framework |
| [Wire](https://github.com/google/wire) | Dependency Injection |
| [Rapido-logger](https://github.com/roppenlabs/rapido-logger-go) | Logging |
| [Viper](https://github.com/spf13/viper) | Configuration Management |
| [Testify](https://github.com/stretchr/testify) | Unit testing and mocking library |
