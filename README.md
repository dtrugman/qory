# 🚀 Qory

![version](https://img.shields.io/github/v/release/dtrugman/qory?color=green)

**A Language Model in Your Terminal**

💻 Skip the alt+tab to your browser!
📋 Stop copy-pasting from other language models!
🔧 Streamline your workflow with terminal queries.

```bash
qory "Please create a basic OpenAPI yaml template" > openapi.yaml
```

Add context from existing files:

```bash
qory openapi.yaml main.py "Please add a /ping endpoint to python server" > ping.py
```

Integrate shell command output:

```bash
qory "This is my project dir" "$(ls)" "How should I improve it?"
```

## 💥 **NEW**: Support for sessions

Keep refining and chatting with the model to improve results.

By default, each `qory` invocation starts a **new session**. You can change this behaviour with the `--last`, `--new`, and `--session` flags, or configure a persistent default with `qory config mode`.

### Continue the last session

```bash
qory --last "With that project structure, where should I put my integration tests?"
```

### Resume a specific session by ID

Use `qory history` to find session IDs, then pass one to `--session`:

```bash
qory history
qory --session <id> "Can you revise that last function?"
```

### Force a new session

```bash
qory --new "Start from scratch"
```

### Set a default mode

Instead of passing `--last` or `--new` every time, configure a default:

```bash
qory config mode set   # choose "new" or "last"
```

- `new` — start a fresh session each time (default)
- `last` — automatically continue the most recent session

Individual `--new` and `--last` flags always override the configured mode.

## 🌟 Install

Qory is compiled for major operating systems and architectures. If your architecture isn't supported, drop a ticket!

Check out the releases page for your platform:

- 🖥️ MacOS / Apple Silicon: `qory_<ver>_darwin_arm64`
- 🖥️ MacOS / Intel: `qory_<ver>_darwin_amd64`
- 🐧 Linux / x64: `qory_<ver>_linux_amd64`
- 🐧 Linux / ARM: `qory_<ver>_linux_arm64`
- 🖥️ Windows / x64: `qory_<ver>_windows_amd64`
- 🖥️ Windows / ARM: `qory_<ver>_windows_arm64`

### 📥 Installation Options

#### Manual

1. Go to the 'releases' tab.
2. Download the appropriate asset.
3. On Unix, set as executable: `chmod +x <file>`, and run it.

⚙️ See the [Configuration](#configuration) section before using it first.

#### Unix

For a quick download and install, choose your preferred method:

- **curl**:

  ```bash
  curl -L -o ./qory https://github.com/dtrugman/qory/releases/download/v0.3.1/qory_0.3.1_darwin_arm64 && chmod +x ./qory && sudo mv ./qory /usr/local/bin/.
  ```

- **wget**:

  ```bash
  wget -O ./qory https://github.com/dtrugman/qory/releases/download/v0.3.1/qory_0.3.1_darwin_arm64 && chmod +x ./qory && sudo mv ./qory /usr/local/bin/.
  ```

Advanced users can install in **any other dir in your PATH**:

```bash
curl -L -o ./qory https://github.com/dtrugman/qory/releases/download/v0.3.1/qory_0.3.1_darwin_arm64 && chmod +x ./qory && mv ./qory ~/.local/bin/.
```

⚙️ Remember to configure before first use by visiting the [Configuration](#configuration) section.

Run your first qory: `qory hi`

## 🔍 Check Installation

Run:

```bash
qory version
```

Ensure it runs successfully.

## ⚙️ Configuration

Before using Qory for the first time, you should set up your provider's Base URL and API key.
Qory uses OpenAI's Chat Completions SDK, and as such OpenAI models work out of the box.
If you want access to all the models out there, we strongly recommend using [Requesty](https://requesty.ai).

Note: Configuration is stored in `~/.config/qory` on MacOS/Linux or `%APPDATA%` on Windows.

### 🔑 API Key Setup

Run:

```bash
qory config api-key set
```

### 🔄 Base URL

Defaults to OpenAI API. Change it by:

1. Setting `OPENAI_BASE_URL`.
2. Using `qory config base-url set`.

### 📌 Model Selection

You can explicitly specify any specific model you want:

```bash
qory config model set gpt-5.4
```

Alternatively, running without a value will fetch all models from the `/v1/models/` endpoint,
and suggest an interactive menu to chosoe from.

### 📌 Persistent Prompt

Configure a custom system prompt to use with your Qory sessions:

```bash
qory config prompt set
```

Example prompt: "do not explain, just provide a concise response"

### ✏️ Editor

When qory is run without any input, it opens an editor so you can type your query:

```bash
qory
```

The editor is resolved in this order:
1. `qory config editor set` (stored config value)
2. `$VISUAL` environment variable
3. `$EDITOR` environment variable
4. `vi` (built-in default)

To change the default editor:

```bash
qory config editor set
```

🔗 For further assistance and updates, visit [Qory on GitHub](https://github.com/dtrugman/qory).
