# ovai - ollama-vertex-ai

HTTP proxy for accessing [Vertex AI] with the REST API interface of [ollama]. Optionally forwarding requests for other models to `ollama`. Written in [Go].

## Synopsis

Get embeddings for a text:

```
❯ curl localhost:22434/api/embed -d '{
  "model": "gemini-embedding-001",
  "input": "Half-orc is the best race for a barbarian."
}'

{ "embeddings": [[0.05424513295292854, -0.023687424138188362, ...]] }
```

## Setup

1. Download an archive with the executable for your hardware and operating system from [GitHub Releases].
2. Download a JSON file with your Google account key from Google Project Console and save it to the current directory under the name `google-account.json`.
3. Optionally create a file `model-defaults.json` in the current directory to change the [default model parameters].
4. Run the server:

```
❯ ovai

Listening on http://localhost:22434 ...
```

Files `model-defaults.json` and `google-account.json` have to be in either JSON or JSONC formats.

### Configuring

The following properties from `google-account.json` are used:

```jsonc
{
  "project_id": "...",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "...",
  "scope": "https://www.googleapis.com/auth/cloud-platform", // optional, can be missing
  "auth_uri": "https://www.googleapis.com/oauth2/v4/token"   // optional, can be missing
}
```

The content of the built-in `model-defaults.json`, which your local `model-defaults.json` will be merged to, if provided. (Properties in `generationConfig` are not set, you can set them in your local `model-defaults.json`.):

```jsonc
{
  "apiLocation": "us-central1",
  "apiEndpoint": "us-central1-aiplatform.googleapis.com",
  // An alternative for the global Google servers:
  // "apiLocation": "global",
  // "apiEndpoint": "aiplatform.googleapis.com",
  "geminiDefaults": {
    "generationConfig": {
      // The following values are examples, they are not set by default,
      // leaving the default behaviour on the Goole implementation:
      // "maxOutputTokens": 8192,
      // "temperature": 1,
      // "topP": 0.95,
      // "topK": 40,
      // "scope": "https://www.googleapis.com/auth/cloud-platform",
      // "thinkingConfig: {
      //   includeThoughts: true,
      //   thinkingBudget: 1024
      // }
    },
    "safetySettings": [
      {
        "category": "HARM_CATEGORY_HATE_SPEECH",
        "threshold": "BLOCK_ONLY_HIGH"
      },
      {
        "category": "HARM_CATEGORY_DANGEROUS_CONTENT",
        "threshold": "BLOCK_ONLY_HIGH"
      },
      {
        "category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
        "threshold": "BLOCK_ONLY_HIGH"
      },
      {
        "category": "HARM_CATEGORY_HARASSMENT",
        "threshold": "BLOCK_ONLY_HIGH"
      }
    ]
  }
}
```

Set the environment variable `PORT` to override the default port 22434.

Set the environment variable `DEBUG` to one or more strings separated by commas to customise logging on `stderr`. The default value is `ovai` when run on the command line and `ovai:srv` inside the Docker container.

| `DEBUG` value | What will be logged                                              |
|:--------------|:-----------------------------------------------------------------|
| `ovai`        | important information about the bodies of requests and responses |
| `ovai:srv`    | methods and URLs of requests and status codes of responses       |
| `ovai:net`    | requests forwarded to Vertex AI and received responses           |
| `ovai,ovai:*` | all information above                                            |

Set the environment variable `OLLAMA_ORIGIN` to the origin of the `ollama` service to enable forwarding to `ollama`. If the requested model doesn't start with `gemini`, `multimodalembedding`, `textembedding`  or `text-embedding`, the request will be forwarded to the `ollama` service. This can be used for using `ovai` as the single service with the `ollama` interface, which recognises both `Vertex AI` and `ollama` models.

Set the environment variable `NETWORK` to enforce IPV4 or IPV6. The default behaviour is to depend on the [Happy Eyeballs] implementation in Go and in the underlying OS. valid values:

| `NETWORK` value | What will be used                            |
|:----------------|:---------------------------------------------|
| `IPV4`          | enforce the network connection via IPV4 only |
| `IPV6`          | enforce the network connection via IPV6 only |

### Docker

For example, run a container for testing purposes with verbose logging, deleted on exit, exposing the port 22434:

    docker run --rm -it -p 22434:22434 -e DEBUG=ovai,ovai:* \
      -v ${PWD}/google-account.json:/usr/src/app/google-account.json \
      ghcr.io/prantlf/ovai

For example, run a container named `ovai` in the background with custom defaults, forwarding to `ollama`, exposing the port 22434:

    docker run --rm -dt -p 22434:22434 --name ovai \
      --add-host host.docker.internal:host-gateway \
      -e OLLAMA_ORIGIN=http://host.docker.internal:11434 \
      -v ${PWD}/google-account.json:/usr/src/app/google-account.json \
      -v ${PWD}/model-defaults.json:/usr/src/app/model-defaults.json \
      prantlf/ovai

And the same task as above, only using Docker Compose (place [docker-compose.yml] or [docker-compose-ollama.yml], if you want to use ollama too, to the current directory) to make it easier. Just a note - [docker-compose-ollama.yml] builds a local image using CPU-only Ollama. If you want to use GPU, customize `Dockerfile.ollama`:

    docker-compose up -d --wait
    docker-compose -f docker-compose-ollama.yml up -d --wait

The image is available as both `ghcr.io/prantlf/ovai` (GitHub) or `prantlf/ovai` (DockerHub).

### Building

Make sure that you have installed [Go] 1.22.3 or newer.

    git clone https://github.com/prantlf/ovai.git
    cd ovai
    make

Executing `./ovai`, `make docker-start` or `make docker-up` will require the `google-account.json` file in the current directory, if you don't just proxy the calls to ollama (which needs the `OLLAMA_ORIGIN` environment variable).

## API

See the original [REST API documentation] for details about the interface. See also the [lifecycle of the Vertex AI models].

### Embeddings

Creates a vectors from the specified input. See the available [embedding models].

```
❯ curl localhost:22434/api/embed -d '{
  "model": "gemini-embedding-001",
  "input": ["Half-orc is the best race for a barbarian."],
  "dimensionality": 768
}'

{ "embeddings": [[0.05424513295292854, -0.023687424138188362, ...]] }
```

The returned vector of floats has 768 dimensions.

Previous request remains supported for compatibility:

```
❯ curl localhost:22434/api/embeddings -d '{
  "model": "gemini-embedding-001",
  "prompt": "Half-orc is the best race for a barbarian."
}'

{ "embedding": [0.05424513295292854, -0.023687424138188362, ...] }
```

The property `dimensionality` is supported by the model `gemini-embedding-001`. Possible values are 768, 1536 and 3072. The default is 3072.

### Text

Generates a text using the specified prompt. See the available [gemini text and chat models].

```
❯ curl localhost:22434/api/generate -d '{
  "model": "gemini-2.5-flash-lite",
  "prompt": "Describe guilds from Dungeons and Dragons.",
  "images": [],
  "think": false,
  "stream": false
}'

{
  "model": "gemini-2.5-flash-lite",
  "created_at": "2024-05-10T14:10:54.885Z",
  "response": "Guilds serve as organizations that bring together individuals with ...",
  "done": true,
  "total_duration": 13884049373,
  "load_duration": 0,
  "prompt_eval_count": 7,
  "prompt_eval_duration: 3471012343,
  "eval_count: 557,
  "eval_duration: 10413037030
}
```

The property `stream` defaults to `true`. The property `think` defaults to `false` and except for boolean values, it accepts strings `high`, `medium`, `low` and `default`. See also [Gemini Thinking]. The property `options` is optional, letting the model provide its defaults. It can be set to the following values, for example:

```
"options": {
  "num_predict": 8192,
  "temperature": 1,
  "top_p": 0.95,
  "top_k": 40,
  // available only for gemini-2.5-flash-lite: 512-24576, 0 or -1 (default:  0)
  //             or for gemini-2.5-flash:        0-24576    or -1 (default: -1)
  //             or for gemini-2.5-pro:        128-32768    or -1 (default: -1)
  "thinking_budget": 0
}
```

### Chat

Replies to a chat with the specified message history. See the available [gemini text and chat models].

```
❯ curl localhost:22434/api/chat -d '{
  "model": "gemini-2.5-flash-lite",
  "messages": [
    {
      "role": "system",
      "content": "You are an expert on Dungeons and Dragons."
    },
    {
      "role": "user",
      "content": "What race is the best for a barbarian?",
      "images": []
    }
  ],
  "think": false,
  "stream": false
}'

{
  "model": "gemini-2.5-flash",
  "created_at": "2024-05-06T23:32:05.219Z",
  "message": {
    "role": "assistant",
    "content": "Half-Orcs are a strong and resilient race, making them ideal for barbarians. ..."
  },
  "done": true,
  "total_duration": 2325524053,
  "load_duration": 0,
  "prompt_eval_count": 9,
  "prompt_eval_duration: 581381013,
  "eval_count: 292,
  "eval_duration: 1744143040
}
```

The property `stream` defaults to `true`. The property `think` defaults to `false` and except for boolean values, it accepts strings `high`, `medium`, `low` and `default`. See also [Gemini Thinking]. The property `options` is optional, letting the model provide its defaults. It can be set to the following values, for example:

```
"options": {
  "num_predict": 8192,
  "temperature": 1,
  "top_p": 0.95,
  "top_k": 40,
  // available only for gemini-2.5-flash-lite: 512-24576, 0 or -1 (default:  0)
  //             or for gemini-2.5-flash:        0-24576    or -1 (default: -1)
  //             or for gemini-2.5-pro:        128-32768    or -1 (default: -1)
  "thinking_budget": 0
}
```

### Tools

An extension to chat that requests information from a local function, which can be used for formulating an answer.

```
❯ curl localhost:22434/api/chat -d '{
  "model": "gemini-2.5-flash-lite",
  "messages": [
    {
      "role": "system",
      "content": "You are an expert on Dungeons and Dragons."
    },
    {
      "role": "user",
      "content": "Is the the initial HP of barbarian higher than 50?"
    }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_hp_by_class",
        "description": "This tool returns the initial health points (HP) for the specified character class.",
        "parameters": {
          "type": "object",
          "properties": {
            "class": {
              "type": "string",
              "description": "Generate one character class from the user request for the health points (HP)"
            }
          },
          "required": ["class"]
        }
      }
    }
  ],
  "think": false,
  "stream": false
}'

{
  "model": "gemini-2.5-flash-lite",
  "created_at": "2024-05-06T23:32:05.219Z",
  "message": {
    "role": "assistant",
    "content": "Let's call a tool for computing the initial HP of barbarian."
    "tool_calls": [
      {
        "function": {
          "name": "get_hp_by_class",
          "arguments": {
            "keyword": "barbarian"
          }
        }
      }
    ],
  },
  "done": true,
  "total_duration": 2325524053,
  "load_duration": 0,
  "prompt_eval_count": 9,
  "prompt_eval_duration: 581381013,
  "eval_count: 292,
  "eval_duration: 1744143040
}

❯ curl localhost:22434/api/chat -d '{
  "model": "gemini-2.5-flash-lite",
  "messages": [
    {
      "role": "system",
      "content": "You are an expert on Dungeons and Dragons."
    },
    {
      "role": "user",
      "content": "What are the initial HP of barbarian?"
    },
    {
      "role": "assistant",
      "content": "Let's call a tool for computing the initial HP of barbarian."
      "tool_calls": [
        {
          "function": {
            "name": "get_hp_by_class",
            "arguments": {
              "keyword": "barbarian"
            }
          }
        }
      ],
    },
    {
      "role": "tool",
      "content": "The initial HP of barbarian is 80."
    }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_hp_by_class",
        "description": "This tool returns the initial health points (HP) for the specified character class.",
        "parameters": {
          "type": "object",
          "properties": {
            "class": {
              "type": "string",
              "description": "Generate one character class from the user request for the health points (HP)"
            }
          },
          "required": ["class"]
        }
      }
    }
  ],
  "think": false,
  "stream": false
}'

{
  "model": "gemini-2.5-flash-lite",
  "created_at": "2024-05-06T23:32:05.219Z",
  "message": {
    "role": "assistant",
    "content": "Yes, the initial HP of barbarian is higher than 50. It's 80."
  },
  "done": true,
  "total_duration": 2325524053,
  "load_duration": 0,
  "prompt_eval_count": 9,
  "prompt_eval_duration: 581381013,
  "eval_count: 292,
  "eval_duration: 1744143040
}
```

The property `stream` defaults to `true`. The property `think` defaults to `false` and except for boolean values, it accepts strings `high`, `medium`, `low` and `default`. See also [Gemini Thinking]. The property `options` is optional, letting the model provide its defaults. It can be set to the following values, for example:

```
"options": {
  "num_predict": 8192,
  "temperature": 1,
  "top_p": 0.95,
  "top_k": 40,
  // available only for gemini-2.5-flash-lite: 512-24576, 0 or -1 (default:  0)
  //             or for gemini-2.5-flash:        0-24576    or -1 (default: -1)
  //             or for gemini-2.5-pro:        128-32768    or -1 (default: -1)
  "thinking_budget": 0
}
```

### Tags

Lists available models.

```
❯ curl localhost:22434/api/tags

{
  "models": [
    {
      "name": "moondream:latest",
      "model": "moondream:latest",
      "modified_at": "2024-06-02T16:39:32.532400236+02:00",
      "size": 1738451197,
      "digest": "55fc3abd386771e5b5d1bbcc732f3c3f4df6e9f9f08f1131f9cc27ba2d1eec5b",
      "details": {
        "parent_model": "",
        "format": "gguf",
        "family": "phi2",
        "families": [
          "phi2",
          "clip"
        ],
        "parameter_size": "1B",
        "quantization_level": "Q4_0"
      },
      "expires_at": "0001-01-01T00:00:00Z"
    }
  ]
}
```

### Show

Show information about a model.

```
❯ curl localhost:22434/api/show -d '{"name":"moondream"}'

{
  "license": "....",
  "modelfile": "...",
  "parameters": "temperature 0\nstop \"\u003c|endoftext|\u003e\"\nstop \"Question:\"",
  "template": "{{ if .Prompt }} Question: {{ .Prompt }}\n\n{{ end }} Answer: {{ .Response }}\n\n",
  "details": {
    "parent_model": "",
    "format": "gguf",
    "family": "phi2",
    "families": [
      "phi2",
      "clip"
    ],
    "parameter_size": "1B",
    "quantization_level": "Q4_0"
  }
}
```

### Ping

Checks that the server is running.

```
❯ curl -f localhost:22434/api/ping -X HEAD
```

### Shutdown

Gracefully shuts down the HTTP server and exits the process.

```
❯ curl localhost:22434/api/shutdown -X POST
```

## Models

### Gemini Thinking

The parameter `think` defaults to `false`. It accepts the following boolean and string values:

| Value       | Effect                                                                                  |
|:------------|:----------------------------------------------------------------------------------------|
| `false`     | disables thinking, if the module supports it, otherwise sets the lowest thinking budget |
| `true`      | enables thinking with the default thinking budget, if the module supports it            |
| `"default"` | enables the default behaviour of the model                                              |
| `"low"`     | enables thinking with the lowest thinking budget, if the module supports it             |
| `"medium"`  | enables thinking with the medium thinking budget, if the module supports it             |
| `"height"`  | enables thinking with the highest thinking budget, if the module supports it            |

The option `thinking_budget` can refine the thinking budget decided by the parameter `think`. The value `-1` enables the default budget according to the selected model. Other thinking budget values:

| Model                   | Range       | Low | Medium | High  | Zero disables | Default |
|:------------------------|------------:|----:|-------:|------:|:-------------:|:-------:|
| `gemini-2.5-flash-lite` | 512 - 24576 | 512 |  12544 | 24576 |      Yes      |    0    |
| `gemini-2.5-flash`      |   0 - 24576 | 128 |  12288 | 24576 |      Yes      |   -1    |
| `gemini-2.5-pro`        | 128 - 32768 | 128 |  16448 | 32768 |      N/A      |   -1    |

### Vertex AI

Recognised models for embeddings: gemini-embedding-001, textembedding-gecko@001, textembedding-gecko@002, textembedding-gecko@003, textembedding-gecko-multilingual@001, text-multilingual-embedding-002, text-embedding-004, text-embedding-005, multimodalembedding@001.

Recognised models for content generation and chat: gemini-2.5-flash-lite, gemini-2.5-flash-lite-preview-06-17, gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-pro-preview-06-05, gemini-2.5-flash-preview-05-20, gemini-2.5-pro-exp-03-25, gemini-2.0-flash-thinking-exp-01-21, gemini-2.0-flash, gemini-2.0-flash-lite, gemini-2.0-flash-exp, gemini-1.5-flash-001, gemini-1.5-flash-002, gemini-1.5-flash-8b-001, gemini-1.5-pro-001, gemini-1.5-pro-002, gemini-1.0-pro-vision-001, gemini-1.0-pro-001, gemini-1.0-pro-002.

### Ollama

Small models usable on machines with less memory and no AI accelerator:

| Name             | Size   |
|:-----------------|-------:|
| deepseek-r1:1.5b | 1.1 GB |
| gemma3:270m      | 291 MB |
| gemma3:1b        | 815 MB |
| gemma3:4b        | 3.3 GB |
| granite3.1-dense:2b    | 1.5 GB |
| granite3.1-moe:1b      | 2.0 GB |
| granite3.1-moe:3b      | 1.4 GB |
| granite-embedding:30m  |  63 MB |
| granite-embedding:278m | 563 MB |
| internlm2:1.8b   | 1.1 GB |
| llama3.2:1b      | 1.3 GB |
| llama3.2:3b      | 2.0 GB |
| llava-phi3       | 2.9 GB |
| moondream        | 1.7 GB |
| nomic-embed-text | 274 MB |
| orca-mini        | 2.0 GB |
| phi              | 1.6 GB |
| phi3             | 2.2 GB |
| qwen3:0.6b       | 397 MB |
| qwen3:1.7b       | 986 MB |
| smollm           | 990 MB |
| smollm:135m      | 91 MB  |
| smollm:360m      | 229 MB |
| snowflake-arctic-embed2 | 1.2 GB |
| stablelm-zephyr  | 1.6 GB |
| stablelm2        | 982 MB |
| tinyllama        | 637 MB |

#### deepseek-r1
DeepSeek's first-generation of reasoning models with comparable performance to OpenAI-o1, including six dense models distilled from DeepSeek-R1 based on Llama and Qwen. 

#### gemma3
Gemma is a lightweight, family of models from Google built on Gemini technology. The Gemma 3 models are multimodal—processing text and images—and feature a 128K context window with support for over 140 languages. Available in 1B, 4B, 12B, and 27B parameter sizes, they excel in tasks like question answering, summarization, and reasoning, while their compact design allows deployment on resource-limited devices.

#### granite3.1-dense
The IBM Granite 2B and 8B models are text-only dense LLMs trained on over 12 trillion tokens of data, demonstrated significant improvements over their predecessors in performance and speed in IBM’s initial testing. 

#### granite3.1-moe
The IBM Granite 1B and 3B models are long-context mixture of experts (MoE) Granite models from IBM designed for low latency usage. 

#### granite-embedding
The IBM Granite Embedding 30M and 278M models models are text-only dense biencoder embedding models, with 30M available in English only and 278M serving multilingual use cases.

#### internlm2
InternLM2.5 is a 7B parameter model tailored for practical scenarios with outstanding reasoning capability.

#### llama3.2
Meta's Llama 3.2 goes small with 1B and 3B models.

#### llava-phi3
A new small LLaVA model fine-tuned from Phi 3 Mini.

#### moondream
moondream2 is a small vision language model designed to run efficiently on edge devices.

#### nemotron-mini
A commercial-friendly small language model by NVIDIA optimized for roleplay, RAG QA, and function calling.

#### nomic-embed-text
A high-performing open embedding model with a large token context window.

#### nuextract
A 3.8B model fine-tuned on a private high-quality synthetic dataset for information extraction, based on Phi-3.

#### orca-mini
A general-purpose model ranging from 3 billion parameters to 70 billion, suitable for entry-level hardware.

#### phi
Phi-2: a 2.7B language model by Microsoft Research that demonstrates outstanding reasoning and language understanding capabilities.

#### phi3
Phi-3 is a family of lightweight 3B (Mini) and 14B (Medium) state-of-the-art open models by Microsoft.

#### qwen3
Qwen3 is the latest generation of large language models in Qwen series, offering a comprehensive suite of dense and mixture-of-experts (MoE) models. 

#### smollm
🪐 A family of small models with 135M, 360M, and 1.7B parameters, trained on a new high-quality dataset.

#### snowflake-arctic-embed2
Snowflake's frontier embedding model. Arctic Embed 2.0 adds multilingual support without sacrificing English performance or scalability. 

#### stablelm-zephyr
A lightweight chat model allowing accurate, and responsive output without requiring high-end hardware.

#### stablelm2
Stable LM 2 is a state-of-the-art 1.6B and 12B parameter language model trained on multilingual data in English, Spanish, German, Italian, French, Portuguese, and Dutch.

#### tinydolphin
An experimental 1.1B parameter model trained on the new Dolphin 2.8 dataset by Eric Hartford and based on TinyLlama.

#### tinyllama
The TinyLlama project is an open endeavor to train a compact 1.1B Llama model on 3 trillion tokens.

## Contributing

In lieu of a formal styleguide, take care to maintain the existing coding style. Lint and test your code.

## License

Copyright (C) 2024-2025 Ferdinand Prantl

Licensed under the [MIT License].

[MIT License]: http://en.wikipedia.org/wiki/MIT_License
[Vertex AI]: https://cloud.google.com/vertex-ai
[ollama]: https://ollama.com
[GitHub Releases]: https://github.com/prantlf/ovai/releases/
[Go]: https://go.dev
[default model parameters]: ./model-defaults.json
[Happy Eyeballs]: https://en.wikipedia.org/wiki/Happy_Eyeballs
[docker-compose.yml]: ./docker-compose.yml
[docker-compose-ollama.yml]: ./docker-compose-ollama.yml
[REST API documentation]: https://github.com/ollama/ollama/blob/main/docs/api.md
[lifecycle of the Vertex AI models]: https://cloud.google.com/vertex-ai/generative-ai/docs/learn/model-versioning
[embedding models]: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/text-embeddings#model_versions
[gemini text and chat models]: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/gemini#model_versions
[Gemini Thinking]: #gemini-thinking
