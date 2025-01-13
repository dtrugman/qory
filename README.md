![qory](./img/qory.png "Logo")

# Qory

**A language model in your terminal**

Stop alt+tab-ing to your browser!

Stop copy-pasting from ChatGPT/Anthropic/Gemini/etc...

Make your life easier by running queries directly from your terminal. As simple as:
```
qory "Please create a basic OpenAPI yaml template" > openapi.yaml
```

Add context from existing files:
```
qory openapi.yaml main.py "Please implement the endpoint /ping in my python server" > ping.py
```

Add output from a shell command:
```
qory "This is my project dir" "$(ls)" "How should I improve it?"
```

## Install

Qory is compiled for all major operating systems and architectures.
If you are looking for an architecture that is not supported, please open a ticket.

Check out the release page and the relevant file for your platform:

- MacOS / Apple Silicon (M1, M2, ...): `qory_<ver>_darwin_arm64`
- MacOS / Intel CPU: `qory_<ver>_darwin_amd64`
- Linux / x64: `qory_<ver>_linux_amd64`
- Linux / ARM: `qory_<ver>_linux_arm64`
- Windows / x64: `qory_<ver>_windows_amd64`
- Windows / ARM: `qory_<ver>_windows_arm64`

There are three install options:

### 1. Manual

Go to the 'releases' tab, and download the right asset for your system.
On Unix systems, set it as an executable using `chmod +x <file>`, and run it.

See the [Configuration](#configuration) section below before using it for the first time.

### 2. Unix

Download and install into a directory of your choosing using a one-liner.

**NOTE:** Adjust the filename you download according to your platform

Download into your system's bin directory, i.e. `/usr/local/bin` (requires `sudo`):
```
curl -L -o ./qory https://github.com/dtrugman/qory/releases/download/v0.1/qory_0.1_darwin_arm64 && chmod +x ./qory && sudo mv ./qory /usr/local/bin/.
```
If you prefer `wget`:
```
wget -O ./qory https://github.com/dtrugman/qory/releases/download/v0.1/qory_0.1_darwin_arm64 && chmod +x ./qory && sudo mv ./qory /usr/local/bin/.
```

If you are an advanced user, feel free to install it into **any other dir in your PATH**:
```
curl -L -o ./qory https://github.com/dtrugman/qory/releases/download/v0.1/qory_0.1_darwin_arm64 && chmod +x ./qory && mv ./qory ~/.local/bin/.
```

See the [Configuration](#configuration) section below before using it for the first time.

And run your first qory: `qory hi`

## Check installation

Run `qory --version` and see if the command succeeds.

## Configuration

Before you use Qory for the first time, you need to set two values, your API key and the model you want to use.

### Choose your preferred model

Run: `qory --config model set`

Use any OpenAI supported model, including: `gpt-4o`, `gpt-4o-mini`, `gpt-o1`, ...

### Setup your API key

Run: `qory --config api-key set`

You only need to do it once. Your configuration is stored under the user's home directory: `~/.config/qory` on MacOS or Linux, `%APPDATA%` on Windows.

### API Key

There are two ways to set up your API key:
1. Using the `OPENAI_API_KEY` environment variable
1. Via the tools, using `qory --config api-key set`

### Model

Set your desired model using: `qory --config model set`.

You can find OpenAI's models on this [page](https://platform.openai.com/docs/models).
An example would be: `gpt-4o` or `gpt-4o-mini`.

### Base URL

By default, the tool will use the public OpenAI API.
If you would like to change it, you can:
1. Set the `OPENAI_BASE_URL` environment variable
1. Use `qory --config api-key set`

### Persistent prompt

If you want to set a custom system prompt to always include with your qory-s,
you can set it using `qory --config prompt set`.

I find it useful to tell the model
"do not explain, just provide the essence of the request"
