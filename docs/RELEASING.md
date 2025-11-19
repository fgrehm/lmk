# Release Instructions

This document contains instructions for releasing new versions of lmk.

## Pre-release Checklist

- [ ] All tests passing (`make test`)
- [ ] Documentation updated (README.md)
- [ ] CHANGELOG.md updated with new version
- [ ] Version number set in code (`lmk.go`)
- [ ] README examples tested
- [ ] Cross-compilation verified (`make xbuild`)
- [ ] Travis CI removed (one-time cleanup for v2.0.0)

## Release Steps

1. **Commit all changes:**
   ```bash
   git add .
   git commit -m "feat: lmk vX.Y.Z - description of changes"
   ```

2. **Create and push tag:**
   ```bash
   git tag vX.Y.Z
   git push origin master
   git push origin vX.Y.Z
   ```

3. **GitHub Actions will automatically:**
   - Run tests on Linux, macOS, and Windows
   - Build binaries for 6 platforms (Linux/macOS/Windows × amd64/arm64)
   - Create GitHub Release
   - Upload binaries and checksums

4. **Post-release (if needed):**
   - Update Homebrew tap formula with new version and checksums
   - Announce release on relevant channels

## GitHub Actions Workflows

### Release Workflow
- **Trigger:** Push tag matching `v*`
- **File:** `.github/workflows/release.yml`
- **Actions:**
  - Checkout code
  - Setup Go 1.23
  - Run GoReleaser
  - Create GitHub Release with binaries

### Test Workflow
- **Trigger:** Push to master, pull requests
- **File:** `.github/workflows/test.yml`
- **Actions:**
  - Matrix test across OS (Linux/macOS/Windows) and Go versions (1.23, 1.24)
  - Run tests with race detection
  - Upload coverage to Codecov

## GoReleaser Configuration

Configuration file: `.goreleaser.yml`

**Builds:**
- 6 platform combinations (Linux/macOS/Windows × amd64/arm64)
- Stripped binaries (`-ldflags="-s -w"`)
- Version injected via ldflags

**Archives:**
- tar.gz for Linux/macOS
- zip for Windows
- Includes README.md and LICENSE.txt

## Manual Build Testing

Before tagging a release, test cross-compilation locally:

```bash
# Build for all platforms
make xbuild

# Verify binaries were created
ls -lh build/

# Test local binary
./lmk -version
./lmk -t 3s -m "Test"
```

## Troubleshooting

**Release failed:**
- Check GitHub Actions logs for errors
- Verify GoReleaser configuration is valid: `goreleaser check`
- Ensure GITHUB_TOKEN has correct permissions

**Binary size issues:**
- Binaries should be ~2MB each
- Use `go tool nm` to inspect symbols
- Verify `-ldflags="-s -w"` is removing debug info

**Missing platforms:**
- Check `.goreleaser.yml` includes all desired GOOS/GOARCH combinations
- Verify Go supports the target platform: `go tool dist list`
