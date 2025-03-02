# Wiki Project Guidelines

## Git Workflow
- **NEVER make changes directly on master**
- Always create a feature branch for new work: `git checkout -b feature-name`
- Branch naming convention: `feature-[descriptive-name]` or `fix-[issue-name]`
- Keep commits focused and atomic with descriptive messages
- Before starting work: 
  1. Check current branch: `git branch`
  2. If on master, create new branch before proceeding
  3. Pull latest changes: `git pull origin master`
  4. Create summary of planned changes as a new markdown file for future reference.  use the same naming convention as the branch name 
- When feature is complete:
  1. Run tests: `go test ./...`
  2. Push branch: `git push -u origin feature-name`
  3. Create PR or merge to master as appropriate

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
- **Testing**: Always include tests when adding features.  Seek to increase test coverage with every change. Table-driven tests preferred, use descriptive error messages
- **Naming**:
  - Functions: camelCase for unexported, TitleCase for exported
  - Variables: descriptive camelCase names
  - Interfaces: single method interfaces named by method + 'er' suffix
- **Documentation**: Add comments above exported functions and types
- **File Organization**: Related functionality grouped in same file
