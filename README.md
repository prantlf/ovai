# ovai - ollama-vertex-ai

REST API proxy to [Vertex AI] with the interface of [ollama]. HTTP server for accessing `Vertex AI` via the REST API interface of `ollama`. Optionally forwarding requests with other models to `ollama`. Written in [Go].

## Synopsis

Get embeddings for a text:

```
❯ curl localhost:22434/api/embeddings -d '{
  "model": "textembedding-gecko@003",
  "prompt": "Half-orc is the best race for a barbarian."
}'

{ "embedding": [0.05424513295292854, -0.023687424138188362, ...] }
```

## Setup

Download Make sure that you have installed [Go] 1.22 or newer.

1. Download an archive with the executable for your hardware and operating system from [GitHub Releases].
2. Download a JSON file with your Google account key from Google Project Console and save it to the current directory under the name `google-account.json`.
3. Optionally create a file `model-defaults.json` in the current directory to change the [default model parameters].
4. Run the server:

```
❯ ovai

Listening on http://localhost:22434 ...
```

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

Set the environment variable `PORT` to override the default port 22434.

Set the environment variable `DEBUG` to one or more strings separated by commas to customise logging on `stderr`. The default value is `ovai` when run on the command line and `ovai:srv` inside the Docker container.

| `DEBUG` value | What will be logged                                              |
|:--------------|:-----------------------------------------------------------------|
| `ovai`        | important information about the bodies of requests and responses |
| `ovai:srv`    | methods and URLs of requests and status codes of responses       |
| `ovai:net`    | requests forwarded to Vertex AI and received responses           |
| `ovai,ovai:*` | all information above                                            |

Set the environment variable `OLLAMA_ORIGIN` to the origin of the `ollama` service to enable forwarding to `ollama`. If the requested model doesn't start with `gemini`, `chat-bison`, `text-bison` or `textembedding-gecko`, the request will be forwarded to the `ollama` service. This can be used for using `ovai` as the single service with the `ollama` interface, which recognises both `Vertex AI` and `ollama` models.

Set the environment variable `NETWORK` to enforce IPV4 or IPV6. The default behaviour is to depend on tHe [Happy Eyeballs] implementation in Go and in the underlying OS. valid values:

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
      ghcr.io/prantlf/ovai

And the same task as above, only using Docker Compose (place [docker-compose.yml] to the current directory) to make it easier:

    docker-compose up -d

The image is available as both `ghcr.io/prantlf/ovai` (GitHub) or `prantlf/ovai` (DockerHub).

### Building

Make sure that you have installed [Go] 1.22 or newer.

    git clone https://github.com/prantlf/ovai.git
    cd ovai
    make

Executing `./ovai`, `make docker-start` or `make docker-up` will require the `google-account.json` file in the current directory.

## API

See the original [REST API documentation] for details about the interface.

### Embeddings

Creates a vector from the specified prompt. See the available [embedding models].

```
❯ curl localhost:22434/api/embeddings -d '{
  "model": "textembedding-gecko@003",
  "prompt": "Half-orc is the best race for a barbarian."
}'

{ "embedding": [0.05424513295292854, -0.023687424138188362, ...] }
```

The returned vector of floats has 768 dimensions.

### Text

Generates a text using the specified prompt. See the available [bison text models] and [gemini chat models].

```
❯ curl localhost:22434/api/generate -d '{
  "model": "gemini-1.5-pro-preview-0409",
  "prompt": "Describe guilds from Dungeons and Dragons.",
  "stream": false
}'

{
  "model": "gemini-1.5-pro-preview-0409",
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

The property `stream` has to be always set to `false`, because the streaming mode isn't supported. The property `options` is optional with the following defaults:

```
"options": {
  "num_predict": 8192,
  "temperature": 1,
  "top_p": 0.95,
  "top_k": 40
}
```

### Chat

Replies to a chat with the specified message history. See the available [bison chat models] and [gemini chat models].

```
❯ curl localhost:22434/api/chat -d '{
  "model": "gemini-1.0-pro",
  "messages": [
    {
      "role": "system",
      "content": "You are an expert on Dungeons and Dragons."
    },
    {
      "role": "user",
      "content": "What race is the best for a barbarian?"
    }
  ],
  "stream": false
}'

{
  "model": "gemini-1.0-pro",
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

The property `stream` has to be always set to `false`, because the streaming mode isn't supported. The property `options` is optional with the following defaults:

```
"options": {
  "num_predict": 8192,
  "temperature": 1,
  "top_p": 0.95,
  "top_k": 40
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

## Contributing

In lieu of a formal styleguide, take care to maintain the existing coding style. Lint and test your code.

## License

Copyright (C) 2024 Ferdinand Prantl

Licensed under the [MIT License].

[MIT License]: http://en.wikipedia.org/wiki/MIT_License
[Vertex AI]: https://cloud.google.com/vertex-ai
[ollama]: https://ollama.com
[GitHub Releases]: https://github.com/prantlf/ovai/releases/
[Go]: https://go.dev
[default model parameters]: ./model-defaults.json
[Happy Eyeballs] :https://en.wikipedia.org/wiki/Happy_Eyeballs
[docker-compose.yml]: ./docker-compose.yml
[REST API documentation]: https://github.com/ollama/ollama/blob/main/docs/api.md
[embedding models]: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/text-embeddings#model_versions
[bison text models]: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/text#model_versions
[bison chat models]: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/text-chat#model_versions
[gemini chat models]: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/gemini#model_versions
