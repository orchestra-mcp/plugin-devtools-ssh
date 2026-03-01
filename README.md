# Orchestra Plugin: devtools-ssh

A tools plugin for the [Orchestra MCP](https://github.com/orchestra-mcp/framework) framework.

## Install

```bash
go install github.com/orchestra-mcp/plugin-devtools-ssh/cmd@latest
```

## Usage

Add to your `plugins.yaml`:

```yaml
- id: tools.devtools-ssh
  binary: ./bin/devtools-ssh
  enabled: true
```

## Tools

| Tool | Description |
|------|-------------|
| `hello` | Say hello to someone |

## Related Packages

- [sdk-go](https://github.com/orchestra-mcp/sdk-go) — Plugin SDK
- [gen-go](https://github.com/orchestra-mcp/gen-go) — Generated Protobuf types
