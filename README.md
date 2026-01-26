# Copilot OpenAI-Compatible Server

An HTTP server that exposes OpenAI-compatible API endpoints (`/v1/chat/completions`, `/v1/models`), powered by the [GitHub Copilot SDK](https://github.com/github/copilot-sdk).

This allows you to use GitHub Copilot's agentic capabilities with any client that supports the OpenAI API (e.g., Open WebUI, custom scripts, etc.).

## Prerequisites

1.  **GitHub Copilot CLI** must be installed and authenticated.
    *   Install via GitHub CLI: `gh extension install github/gh-copilot`
    *   Or ensure the `copilot` binary is in your PATH.
    *   Verify installation: `copilot --version`
2.  **Go 1.21+** installed.

## Quick Start

### 1. Build the Server

```bash
git clone https://github.com/yourusername/copilot-openai-server.git
cd copilot-openai-server
go build -o copilot-server .
```

### 2. Run the Server

```bash
# Default port is 8080
./copilot-server

# Specify a custom port
./copilot-server -port 9000
```

## API Endpoints

### List Models (`GET /v1/models`)

Lists available models from the Copilot SDK.

```bash
curl http://localhost:8080/v1/models
```

### Chat Completions (`POST /v1/chat/completions`)

Supports standard OpenAI chat completion parameters, including streaming and tool calling.

**Basic Request:**

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Hello! How does this server work?"}
    ]
  }'
```

**Streaming Request:**

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Write a short poem about coding."}],
    "stream": true
  }'
```

**Tool Calling:**

The server supports defining tools in the OpenAI format. These are passed to the Copilot SDK, which allows the model to "call" them. Note that the **client** is responsible for executing the tool and sending the result back if needed (standard OpenAI flow).

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "What'\''s the weather in London?"}],
    "tools": [{
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get current weather",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {"type": "string"}
          },
          "required": ["location"]
        }
      }
    }]
  }'
```

## Open WebUI Integration

You can easily use this with [Open WebUI](https://docs.openwebui.com/):

1.  Run this server (`./copilot-server`).
2.  In Open WebUI, go to **Settings > Connections > OpenAI**.
3.  Add a new connection:
    *   **API Base URL:** `http://localhost:8080`
    *   **API Key:** `any-string` (not validated by this server, but authentication with Copilot happens via the CLI on the host machine).
4.  Save and select a Copilot model to start chatting.

## Development

To add this as a dependency in your own Go project:

```bash
go get github.com/github/copilot-sdk/go
```

## License

MIT
