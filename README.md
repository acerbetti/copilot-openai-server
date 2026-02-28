# Copilot OpenAI-Compatible Server

An HTTP server that exposes OpenAI-compatible API endpoints (`/v1/chat/completions`, `/v1/models`), powered by the [GitHub Copilot SDK](https://github.com/github/copilot-sdk).

This allows you to use GitHub Copilot's agentic capabilities with any client that supports the OpenAI API (e.g., Open WebUI, custom scripts, etc.).

## Prerequisites

1.  **GitHub Copilot CLI** must be installed and authenticated.
    - Install via GitHub CLI: `gh extension install github/gh-copilot`
    - Or ensure the `copilot` binary is in your PATH.
    - Verify installation: `copilot --version`
2.  **Go 1.21+** installed.

## Docker

You can run the server using Docker, which includes the GitHub Copilot CLI.

### Build the Docker Image

```bash
docker build -t copilot-openai-server .
```

### Run the Container

Set your GitHub personal access token with Copilot permissions. Historically this was the only way to authenticate, but starting in this version the token may also be supplied on each request (see **API Key** section below).

```bash
export GH_TOKEN="your_personal_access_token"          # default/fallback token
docker run -e GH_TOKEN=$GH_TOKEN -p 8080:8080 copilot-openai-server
```

**GitHub personal access token (PAT)** must be fine-grained with the `Copilot requests` permission enabled; see https://github.com/github/copilot-cli/issues/91 for details.

The server will be available at `http://localhost:8080`.

### Versioning note

The Dockerfile pulls the latest Copilot SDK and Copilot CLI at build time instead of pinning specific releases ([Dockerfile](Dockerfile)). If a future SDK or CLI update introduces an incompatible change, the image may break until you pin or update both components together.

## Quick Start

### 1. Build the Server

```bash
git clone https://github.com/RajatGarga/copilot-openai-server.git
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

You can pass an API key via the body or header. The header wins if both are supplied.

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $GH_TOKEN" \
  -d '{
    "model": "gpt-4o",
    "api_key": "'$GH_TOKEN'",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Hello! How does this server work?"}
    ]
  }'
```

When no key is supplied, the server will fall back to the value of `GH_TOKEN` set when the process started. If no key is available at all the request will be rejected with HTTP 401.

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
    - **API Base URL:** `http://localhost:8080`
    - **API Key:** `any-string` (not validated by this server, but authentication with Copilot happens via the CLI on the host machine). Alternatively you can provide a GitHub token on each request as described below.
4.  Save and select a Copilot model to start chatting.

## Development

To add this as a dependency in your own Go project:

```bash
go get github.com/github/copilot-sdk/go
```

## License

GPL-3.0 License
