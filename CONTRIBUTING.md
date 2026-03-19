# Contributing to toggl-tui

Thanks for your interest in contributing! This guide covers the workflow and conventions for the project.

## Getting Started

```bash
git clone https://github.com/gdiab/toggl-tui.git
cd toggl-tui
make build      # Build
make test       # Run tests
make run        # Build and run
```

## Making Changes

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Run `make test` and `make vet` to verify
4. Open a pull request against `main`

### Branch naming

Use a descriptive prefix:

```
feat/edit-entry-duration
fix/timer-height-bug
docs/update-readme
ci/add-lint-step
```

## Pull Requests

PRs are squash-merged into `main`. **The PR title becomes the commit message**, so it must follow [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <description>
```

### Types

| Type       | When to use                              |
|------------|------------------------------------------|
| `feat`     | New feature                              |
| `fix`      | Bug fix                                  |
| `docs`     | Documentation only                       |
| `ci`       | CI/CD changes                            |
| `refactor` | Code change, no new feature or fix       |
| `test`     | Adding or fixing tests                   |
| `chore`    | Build, deps, tooling                     |

### Examples

```
feat(dashboard): add inline edit for entry project
fix(timer): pass terminal height on form creation
docs: add contributing guide
ci: add GoReleaser CD workflow
chore: fix module path to match repo URL
```

### Scope

Scope is optional but encouraged. Common scopes: `dashboard`, `timer`, `api`, `config`, `setup`.

### Requirements

- CI must pass (vet, test, build)
- 1 approving review required
- Keep PRs focused — one logical change per PR

## Code Style

- Run `go fmt` before committing
- Run `go vet` to catch issues
- Keep changes minimal — only touch what's necessary
- No need to add comments for self-evident code

## Releases

Releases are automated. When a version tag is pushed (`v*`), GoReleaser builds binaries and creates a GitHub Release.

```bash
git tag v0.2.0
git push origin v0.2.0
```

Version tags follow [semver](https://semver.org/):
- **Patch** (`v0.1.3`): bug fixes
- **Minor** (`v0.2.0`): new features, backwards compatible
- **Major** (`v1.0.0`): breaking changes
