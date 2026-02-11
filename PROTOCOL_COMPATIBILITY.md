# MCP Protocol Version Compatibility

## Automatic Protocol Detection

**As of v3.0.2**, this server implements automatic MCP protocol version detection for universal compatibility with any MCP client.

### How It Works

When a client sends the `initialize` request, the server:
1. Reads the `protocolVersion` from the client's params
2. Responds with the **same version** the client requested
3. Falls back to `2024-11-05` if no version is specified

### Example

**Client sends:**
```json
{
  "method": "initialize",
  "params": {
    "protocolVersion": "2025-06-18",
    "capabilities": {...}
  }
}
```

**Server responds:**
```json
{
  "result": {
    "protocolVersion": "2025-06-18",
    "capabilities": {...}
  }
}
```

### Supported Protocol Versions

This server is compatible with **all MCP protocol versions** because it mirrors the client's requested version:
- ✅ `2024-11-05` (MCP spec v1.0)
- ✅ `2025-06-18` (Claude Desktop current)
- ✅ `2025-11-25` (Latest MCP spec)
- ✅ Any future versions

### Fallback Behavior

If the client doesn't specify a protocol version (edge case), the server defaults to:
- **Default**: `2024-11-05` (stable, widely supported)

### Migration from Previous Versions

**v2.2.1 and earlier**: Used fixed protocol version `2024-11-05`
**v3.0.0 - v3.0.1**: Used fixed protocol version `2025-11-25` (caused compatibility issues)
**v3.0.2+**: Automatic detection (universal compatibility) ✅

### For Developers

The protocol version detection is implemented in `internal/server/server.go`:

```go
case "initialize":
    // Extract client's requested protocol version and respond with the same
    clientProtocolVersion := "2024-11-05" // Default fallback
    if version, ok := req.Params["protocolVersion"].(string); ok && version != "" {
        clientProtocolVersion = version
    }

    response.Result = map[string]interface{}{
        "protocolVersion": clientProtocolVersion,
        // ...
    }
```

### Testing Different Protocol Versions

You can test the server with different protocol versions using any MCP client. The server will automatically adapt.

### No Recompilation Needed

Users who clone this repository and compile will get a server that works with **any MCP client version** automatically.
