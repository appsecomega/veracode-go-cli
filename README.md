# veracode-go-cli

Alternative CLI to query application information from the Veracode platform (e.g. Analysis Center URL).

## Requirements

- Go 1.22+
- Veracode credentials in the environment:

```bash
export VERACODE_API_KEY_ID="..."
export VERACODE_API_KEY_SECRET="..."
```

## Installation

### Via `go install` (recommended)

```bash
go install github.com/appsecomega/veracode-go-cli/cmd/veracode-go-cli@latest
```

For a specific version:

```bash
go install github.com/appsecomega/veracode-go-cli/cmd/veracode-go-cli@v0.1.0
```

### Via GitHub Releases

Download the binary for your platform from [Releases](https://github.com/appsecomega/veracode-go-cli/releases) (linux/darwin/windows, amd64/arm64).

### From source

```bash
git clone https://github.com/appsecomega/veracode-go-cli.git
cd veracode-go-cli
make build
./bin/veracode-go-cli --help
```

## Usage

```bash
# Fetch an application (default test case)
veracode-go-cli app get --name "WebGoat-Legacy-master"

# JSON output (useful for scripting)
veracode-go-cli app get --name "WebGoat-Legacy-master" --output json

# Binary version
veracode-go-cli version
```

Text output:

```
Application: WebGoat-Legacy-master
App ID: 123456
GUID: ...
Analysis Center: https://analysiscenter.veracode.com/auth/index.jsp#...
```

## Versioned releases

1. Update the code and open a PR against `main` (CI runs `vet` + `test` + `build`).
2. Create and push a SemVer tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

3. The `Release` workflow runs GoReleaser and publishes the binaries with the version injected via `-ldflags` (`veracode-go-cli version` shows `Version/Commit/Date`).

To test packaging locally without publishing:

```bash
make snapshot
```

## Layout

```
cmd/veracode-go-cli/   thin entrypoint (just calls internal/cli)
internal/cli/          Cobra commands (root, app get, version) — add subcommands here
internal/config/       credentials via env
internal/veracode/     HTTP client + HMAC auth + application lookup
internal/version/      version injected via ldflags (Makefile/GoReleaser)
.github/workflows/    CI (vet/test/build) + Release (GoReleaser)
```


## Contributing notes
To scale: create a new file under `internal/cli/` (e.g. `app_list.go`) and register it in `newAppCmd()`.
