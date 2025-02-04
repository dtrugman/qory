# 🚀 Qory

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

Follow up on last query:

```bash
qory ^ "With that project structure, where should I put my integration tests?"
```

Use smart sessions:

```bash
qory -s gosort "Please implement a method to find all files in a dir and sort them"
```

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
  curl -L -o ./qory https://github.com/dtrugman/qory/releases/download/v0.2.2/qory_0.2.2_darwin_arm64 && chmod +x ./qory && sudo mv ./qory /usr/local/bin/.
  ```

- **wget**:

  ```bash
  wget -O ./qory https://github.com/dtrugman/qory/releases/download/v0.2.2/qory_0.2.2_darwin_arm64 && chmod +x ./qory && sudo mv ./qory /usr/local/bin/.
  ```

Advanced users can install in **any other dir in your PATH**:

```bash
curl -L -o ./qory https://github.com/dtrugman/qory/releases/download/v0.2.2/qory_0.2.2_darwin_arm64 && chmod +x ./qory && mv ./qory ~/.local/bin/.
```

⚙️ Remember to configure before first use by visiting the [Configuration](#configuration) section.

Run your first qory: `qory hi`

## 🔍 Check Installation

Run:

```bash
qory --version
```

Ensure it runs successfully.

## ⚙️ Configuration

Before using Qory, set up your API key and preferred model.

Save once. Configuration is stored in `~/.qory` on MacOS/Linux or `%APPDATA%` on Windows.

### 📌 Model Selection

Run:

```bash
qory --config model set
```

Use any OpenAI model, including: `gpt-4o`, `gpt-4o-mini`, `gpt-o1`, ...

Check available models [here](https://platform.openai.com/docs/models).

### 🔑 API Key Setup

Run:

```bash
qory --config api-key set
```

#### Alternatives to Set API Key

### 🔄 Base URL

Defaults to OpenAI API. Change it by:

1. Setting `OPENAI_BASE_URL`.
2. Using `qory --config base-url set`.

### 📌 Persistent Prompt

Create a custom system prompt to always accompany your qory commands:

```bash
qory --config prompt set
```

Example prompt: "do not explain, just provide the essence of the request" 💡

🔗 For further assistance and updates, visit [Qory on GitHub](https://github.com/dtrugman/qory).
