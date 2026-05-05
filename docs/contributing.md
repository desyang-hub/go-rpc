# Contributing

Contributions are welcome! Here are guidelines to help.

## Development Setup

```bash
git clone https://github.com/desyang/go-rpc.git
cd go-rpc
make setup
```

## Project Structure

```
/
├── cmd/            # Main packages
├── pkg/            # Public packages
├── internal/       # Private packages
├── api/            # Protobuf definitions
├── generators/     # Code generation templates
├── examples/       # Example projects
├── docs/           # Documentation
└── scripts/        # Build and CI scripts
```

## Guidelines

### Git Commits

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new feature
fix: fix bug
docs: update documentation
refactor: restructure code
test: add tests
```

### Pull Requests

1. Fork the repository and create a new branch
2. Make your changes
3. Run all tests: `make test`
4. Check linting: `make lint`
5. Submit a pull request

### Code Review

All pull requests require at least one approval from maintainers.

## Running Tests

```bash
make test
make test-coverage
```

## Building Documentation

```bash
pip install mkdocs-material
mkdocs serve
```
