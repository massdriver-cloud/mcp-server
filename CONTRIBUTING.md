# Contributing

## Development

```bash
make test     # Run tests
make lint     # Run linters (go vet + golangci-lint)
make build    # Build binary (outputs ./bin/mcp-server)
make tidy     # go mod tidy
```

## Architecture

```
main.go              # Entrypoint — initializes SDK client, starts stdio server
mcp/
  server.go          # MCP server setup, tool registration, SDK wiring
  tools/
    services.go      # Service interfaces + Client struct (DI)
    helpers.go       # textResult, jsonResult, mutationErr helpers
    metadata.go      # Behavioral annotations, display titles, enum constraints
    projects.go      # One file per service domain
    ...
```

Tool handlers depend on service interfaces defined in `services.go`. In production, `server.go` wires real SDK services. In tests, stub structs with func fields are injected directly.

## Adding a Tool

1. Add the method to the appropriate interface in `mcp/tools/services.go`
2. Add the tool definition and handler in the corresponding `mcp/tools/<service>.go`
3. Register the tool in `mcp/server.go`
4. Add the tool to the matching annotation group in `mcp/tools/metadata.go` (read-only / additive / update / destructive)
5. Add stub method and tests in the `*_test.go` file
6. Add the tool name to the expected list in `main_test.go`
7. Add the input type to `allToolInputs` (and bump `wantTools`) in `mcp/tools/schema_required_test.go`
8. Add the tool to `registeredTools()` (and bump its count) in `mcp/tools/metadata_test.go`
9. Update the tool lists and counts in `README.md` and `MCP_README.md`

> **Optional fields:** any input field that the handler does not enforce as
> required must carry `,omitempty` in its json tag. The go-sdk infers a property
> as *required* whenever `,omitempty` is absent, so a missing tag silently
> advertises an optional field as required and MCP clients will reject valid
> calls. `TestOptionalFieldsAreNotRequired` guards this invariant.

## Releasing

Pushing a semver tag kicks off the full release:

```bash
git tag v1.0.0
git push --tags
```

This triggers three workflows:

1. **GitHub Releases** ([release.yaml](.github/workflows/release.yaml)) — GoReleaser builds binaries for linux/darwin/windows (amd64 + arm64) per [.goreleaser.yaml](.goreleaser.yaml) and creates the GitHub release with checksums and a changelog.
2. **Docker Hub** ([dockerhub.yaml](.github/workflows/dockerhub.yaml)) — multi-arch image (linux/amd64 + arm64) pushed to `massdrivercloud/mcp-server`, tagged with the release version and `latest`.
3. **MCP Registry** (same workflow, after the image push) — publishes [server.json](server.json) to the [official MCP Registry](https://registry.modelcontextprotocol.io) as `cloud.massdriver/mcp-server`, with the version and image tag stamped from the git tag.

The server version reported to MCP clients comes from `mcp.Version`, stamped via ldflags in all three build paths (Makefile, GoReleaser, Dockerfile).

### MCP Registry authentication

Registry publishing uses DNS-based namespace verification for `cloud.massdriver/*`:

- A TXT record on `massdriver.cloud` holds the public key (`v=MCPv1; k=ed25519; p=...`), managed in Terraform alongside the other DNS records.
- The matching Ed25519 private key is stored in the `MCP_REGISTRY_PRIVATE_KEY` GitHub Actions secret (hex-encoded seed). If it ever needs to be rotated, generate a new keypair, update the TXT record, and replace the secret:

  ```bash
  openssl genpkey -algorithm Ed25519 -out key.pem
  # TXT record value:
  echo "v=MCPv1; k=ed25519; p=$(openssl pkey -in key.pem -pubout -outform DER | tail -c 32 | base64)"
  # Secret value:
  openssl pkey -in key.pem -noout -text | grep -A3 'priv:' | tail -n +2 | tr -d ' :\n'
  ```

Verify a published release:

```bash
curl "https://registry.modelcontextprotocol.io/v0.1/servers?search=cloud.massdriver/mcp-server"
```
