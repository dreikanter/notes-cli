# notescli

## Build & Install

```sh
make install     # builds and installs to ~/go/bin/notes
make build       # builds local ./notes binary
make test        # run tests
make lint        # run golangci-lint
```

## Versioning

Version is set at build time via git tags and `-ldflags`. The `Version` var in
`internal/cli/root.go` defaults to `"dev"` and is overridden by `make install`
/ `make build` using `git describe --tags`.

Version format: `v0.{PR_number}.0` (e.g. PR #5 -> `v0.5.0`).

Tags are created automatically by GitHub Actions on PR merge (`.github/workflows/tag.yml`).

After merging a PR, reinstall locally:

```sh
git checkout main && git pull --tags
make install
```

Manual fallback: `make tag V=0.X.0` (X = PR number).
