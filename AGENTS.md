1. Project type: Go CLI tool that generates Codex skills from a single docs URL.
2. Target Go version: `1.25.x`.
3. Binary name: `cli-skill` (update if different).
4. Source root: `.` with `cmd/cli-skill/` as the main package.
5. Initialize module (once): `go mod init <module-path>`.
6. Core deps (pin to versions from stack research):
   - `go get github.com/spf13/cobra@v1.10.2`
   - `go get charm.land/huh/v2@v2.0.3`
   - `go get github.com/openai/openai-go/v3@v3.26.0`
   - `go get github.com/spf13/viper@v1.21.0`
7. Supporting deps (as needed):
   - `go get github.com/PuerkitoBio/goquery@v1.11.0`
   - `go get github.com/hashicorp/go-retryablehttp@v0.7.8`
   - `go get github.com/santhosh-tekuri/jsonschema/v6@v6.0.2`
   - `go get github.com/go-playground/validator/v10@v10.30.1`
   - `go get github.com/BurntSushi/toml@v1.6.0`
   - `go get github.com/sahilm/fuzzy@v0.1.1`
   - `go get dario.cat/mergo@v1.0.2`
   - `go get github.com/spf13/afero@v1.15.0`
8. Tidy deps: `go mod tidy`.
9. Format: `go fmt ./...`.
10. Vet: `go vet ./...`.
11. Lint (if installed): `golangci-lint run ./...`.
12. Tests: `go test ./...`.
13. Run locally: `go run ./cmd/cli-skill --help`.
14. Build: `go build -o bin/cli-skill ./cmd/cli-skill`.
15. Install to `GOBIN`: `go install ./cmd/cli-skill`.
16. Show version (if supported): `./bin/cli-skill --version`.
17. Clean build artifacts: `rm -rf bin/`.
