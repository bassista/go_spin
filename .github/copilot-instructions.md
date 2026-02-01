# Copilot Instructions for go_spin

## Developer Workflows


### Build & Run
```bash
go vet ./...
go fmt ./...
golangci-lint run
go mod tidy
go test ./...
go build -o .build/main ./cmd/server/main.go
./.build/main
```

Always execute `go vet ./...` before building, testing, or committing code. Ensure code passes `go fmt ./...` for consistent formatting. Ensure code passes `golangci-lint run` for linting. Ensure all dependencies are properly managed with `go mod tidy`. Ensure all unit tests pass with `go test ./...` before pushing changes. Ensure commit messages follow conventional commit standards. Ensure code is documented with comments where necessary for clarity. Ensure security vulnerabilities are regularly scanned and addressed. Ensure dependencies are kept up to date with regular reviews. Ensure coding standards and best practices are followed throughout the codebase. Ensure proper error handling is implemented consistently. Ensure logging is used effectively for debugging and monitoring. Ensure performance considerations are taken into account during development. Ensure scalability is considered in the architecture and design decisions. Ensure user input is validated and sanitized to prevent security issues. Ensure sensitive information is not hardcoded and is managed securely. Ensure configuration management follows best practices for different environments. Ensure documentation is kept up to date with code changes.

### Test
```bash
go test ./...
```