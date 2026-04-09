# Contributing

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/blueprint.git`
3. Create a feature branch: `git checkout -b feature/your-feature`

## Development

### Prerequisites
- Go 1.21+
- Node.js 18+
- Bun (optional, for frontend)

### Frontend Setup
```bash
cd frontend
bun install
bun run dev
```

### Backend Setup
```bash
cd backend
go mod download
go run cmd/server/main.go
```

## Code Style

- Run `go fmt ./...` before commits (backend)
- Run `bun run lint` before commits (frontend)
- Use meaningful commit messages

## Pull Requests

1. Ensure all tests pass
2. Update documentation if needed
3. Describe your changes in the PR description

## Issues

- Use clear, descriptive titles
- Include reproduction steps
- Tag appropriately (bug, feature, enhancement)