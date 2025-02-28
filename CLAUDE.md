# Wiki Project Guidelines

## Build & Test Commands
- Build: `go build`
- Run: `go run wiki.go`
- Run all tests: `go test ./...`
- Run single test: `go test -run TestName` (e.g., `go test -run TestViewHandler`)
- Run tests with verbose output: `go test -v ./...`

## Code Style Guidelines
- **Formatting**: Follow Go standard formatting (use `gofmt` or `go fmt ./...`)
- **Imports**: Group standard library imports first, then third-party packages with blank line between
- **Error Handling**: Check and handle errors explicitly, use `log.Printf` for context
- **Testing**: Table-driven tests preferred, use descriptive error messages
- **Naming**:
  - Functions: camelCase for unexported, TitleCase for exported
  - Variables: descriptive camelCase names
  - Interfaces: single method interfaces named by method + 'er' suffix
- **Documentation**: Add comments above exported functions and types
- **File Organization**: Related functionality grouped in same file