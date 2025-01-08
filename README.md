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

## Configuration

To start using Qory, you need to set two values, your API key and the model you want to use.

You only need to do it once. Your configurations are stored under the user's home directory.

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
