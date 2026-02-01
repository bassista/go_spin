# Copilot Instructions for go_spin

## Developer Workflows

### Build & Run
```bash
go build -o .build/main ./cmd/server/main.go
./.build/main
```

Always execute go vet ./... before building, testing, or committing code. ensure code passes go fmt ./... for consistent formatting. ensure code passes golangci-lint run for linting. ensure all dependencies are properly managed with go mod tidy. ensure all unit tests pass with go test ./... before pushing changes. ensure commit messages follow conventional commit standards. ensure code is documented with comments where necessary for clarity. ensure pull requests include a clear description of changes made. ensure code reviews are conducted for all pull requests before merging. ensure branch protection rules are in place to enforce checks before merging. ensure continuous integration is set up to run tests and checks on each push. ensure code coverage is monitored and maintained at an acceptable level. ensure security vulnerabilities are regularly scanned and addressed. ensure dependencies are kept up to date with regular reviews. ensure coding standards and best practices are followed throughout the codebase. ensure proper error handling is implemented consistently. ensure logging is used effectively for debugging and monitoring. ensure performance considerations are taken into account during development. ensure scalability is considered in the architecture and design decisions. ensure user input is validated and sanitized to prevent security issues. ensure sensitive information is not hardcoded and is managed securely. ensure configuration management follows best practices for different environments. ensure documentation is kept up to date with code changes. 
```

### Test
```bash
go test ./...
```

