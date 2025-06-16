# Expenses Tracker

A full-stack expense tracking application built with Go backend and React frontend.

## Architecture

- **Backend**: Go with gRPC/Connect for API services
- **Frontend**: React + TypeScript with Vite, TailwindCSS, and TanStack Router
- **Infrastructure**: Terraform for multi-cloud deployment (Railway, Fly.io, Vercel)
- **Protocol Buffers**: Type-safe API definitions with code generation

## Development Setup

### Prerequisites

- Go 1.23+
- Node.js 18+
- Buf CLI (for protocol buffer generation)
- Terraform (for infrastructure)

### Quick Start

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd expenses
   ```

2. **Backend Setup**

   ```bash
   cd backend
   go mod download
   make dev
   ```

3. **Frontend Setup**

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. **Generate Protocol Buffers** (if modified)

   ```bash
   # Backend
   cd backend && buf generate

   # Frontend
   cd frontend && npm run gen
   ```

## Git Workflow

This project follows **Git Flow** with feature branches, pull requests, and semantic releases.

### Branch Structure

- `main` - Production-ready code
- `develop` - Integration branch for development
- `feature/*` - Feature development branches
- `bugfix/*` - Bug fix branches
- `hotfix/*` - Critical production fixes
- `release/*` - Release preparation branches

### Development Process

1. **Create Feature Branch**

   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/your-feature-name
   ```

2. **Development**

   - Make your changes
   - Write tests
   - Ensure all checks pass:

     ```bash
     # Backend
     cd backend && go test ./...

     # Frontend
     cd frontend && npm run test && npm run lint && npm run check
     ```

3. **Commit Guidelines**
   Follow [Conventional Commits](https://www.conventionalcommits.org/):

   ```
   feat: add expense categorization
   fix: resolve calculation rounding issue
   docs: update API documentation
   test: add unit tests for expense service
   ```

4. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

### Pull Request Process

#### Before Creating a PR

- [ ] All tests pass locally
- [ ] Code follows project conventions
- [ ] Documentation updated if needed
- [ ] No merge conflicts with target branch

#### PR Requirements

- **Title**: Clear, descriptive summary
- **Description**: What changes were made and why
- **Testing**: How the changes were tested
- **Screenshots**: For UI changes
- **Breaking Changes**: List any breaking changes

#### PR Template

```markdown
## Summary

Brief description of changes

## Changes Made

- List of specific changes

## Testing

- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist

- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

#### Review Process

1. **Automated Checks**: All CI/CD checks must pass
2. **Code Review**: At least one approval required
3. **Testing**: Reviewer should test the changes
4. **Merge**: Use "Squash and Merge" for feature branches

### Release Process

#### Semantic Versioning

- `MAJOR.MINOR.PATCH` (e.g., 1.2.3)
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

#### Release Steps

1. **Create Release Branch**

   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b release/v1.2.0
   ```

2. **Prepare Release**

   - Update version numbers
   - Update CHANGELOG.md
   - Final testing

3. **Create Release PR**

   - Target: `main`
   - Title: `Release v1.2.0`

4. **After Merge**

   ```bash
   git checkout main
   git pull origin main
   git tag v1.2.0
   git push origin v1.2.0

   # Merge back to develop
   git checkout develop
   git merge main
   git push origin develop
   ```

5. **Deploy**
   - Automatic deployment from `main` branch
   - Monitor deployment status

### Hotfix Process

For critical production issues:

1. **Create Hotfix Branch**

   ```bash
   git checkout main
   git checkout -b hotfix/critical-fix
   ```

2. **Fix and Test**

   - Make minimal changes
   - Test thoroughly

3. **Create PR to Main**

   - Fast-track review
   - Deploy immediately after merge

4. **Backport to Develop**
   ```bash
   git checkout develop
   git merge main
   git push origin develop
   ```

## Scripts and Commands

### Backend

```bash
# Development
make dev          # Start development server
make build        # Build binary
make test         # Run tests
make gen          # Generate protocol buffers

# Database
make db-up        # Start database
make db-down      # Stop database
make db-migrate   # Run migrations
```

### Frontend

```bash
npm run dev       # Development server
npm run build     # Production build
npm run test      # Run tests
npm run lint      # Lint code
npm run check     # Type check
npm run gen       # Generate types from protobuf
```

### Infrastructure

```bash
cd deploy/terraform
terraform plan    # Plan infrastructure changes
terraform apply   # Apply infrastructure changes
terraform destroy # Destroy infrastructure
```

## Deployment

### Environments

- **Development**: Automatic deployment from `develop` branch
- **Staging**: Automatic deployment from release branches
- **Production**: Automatic deployment from `main` branch

### Deployment Targets

- **Backend**: Railway/Fly.io
- **Frontend**: Vercel
- **Database**: Managed database services

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

- Create an issue for bug reports
- Use discussions for questions
- Check existing issues before creating new ones

## License

[Add your license here]

