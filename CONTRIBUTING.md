# Contributing to Mutating Registry Webhook

Thank you for your interest in contributing to the Mutating Registry Webhook project! This document provides guidelines and workflows for contributing.

## Development Setup

1. Fork and clone the repository
2. Install prerequisites:
   - Go 1.19 or higher
   - Docker
   - Kubernetes cluster (v1.16+)
   - Helm 3.0+

## Development Workflow

1. Create a new branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following our coding standards

3. Run tests locally:
   ```bash
   make test
   ```

4. Build and verify locally:
   ```bash
   make build
   make docker-build
   make test-deploy
   ```

## Coding Standards

- Follow Go best practices and conventions
- Use meaningful variable and function names
- Add comments for complex logic
- Include unit tests for new features
- Keep functions focused and concise

## Pull Request Process

1. Update documentation for any new features
2. Ensure all tests pass locally
3. Create a Pull Request with:
   - Clear title and description
   - Reference to related issues
   - Description of changes and impact

### PR Checks

Each PR triggers automated checks:
- Unit tests
- Build verification
- Helm chart validation

All checks must pass before merging.

## Release Process

1. Releases are created from the `main` branch
2. Tag format: `vX.Y.Z` following semantic versioning
3. Release workflow automatically:
   - Builds and tests code
   - Creates Docker image
   - Packages Helm chart
   - Creates GitHub release

## Getting Help

- Create an issue for bugs or feature requests
- Use PR comments for code-related discussions
- Tag maintainers for urgent matters

## Code Review

- All changes require at least one review
- Address review comments promptly
- Maintainers will merge approved PRs

Thank you for contributing!