# Changes

## [0.12.0](https://github.com/prantlf/ovai/compare/v0.11.0...v0.12.0) (2025-01-01)

### Features

* Support the new /api/embed ([bf08701](https://github.com/prantlf/ovai/commit/bf087017b02000e7dc63fe775d20a1492c81f6fc))
* Recognise gemini-2.0-flash-exp ([439ffa3](https://github.com/prantlf/ovai/commit/439ffa3337816b0f4175d1b3ffb09fbba437aeb1))

## [0.11.0](https://github.com/prantlf/ovai/compare/v0.10.0...v0.11.0) (2024-11-06)

### Features

* Support method HEAD for /api/ping too ([5fd367b](https://github.com/prantlf/ovai/commit/5fd367b38a0369a0b7e31a1ee6137d50f546fdbd))
* Add docker compose example for ovai & ollama ([04677c1](https://github.com/prantlf/ovai/commit/04677c1e3d409e5157b64449fdb126f1bdcf92ff))

## [0.10.0](https://github.com/prantlf/ovai/compare/v0.9.0...v0.10.0) (2024-11-02)

### Features

* Remove support for legacy PaLM 2 models ([3325c54](https://github.com/prantlf/ovai/commit/3325c54a64ac4dbb65439f12f728a8a00ac9c286))

### BREAKING CHANGES

Models text-bison\* chat-bison\* and text-unicorn\*
have been removed. They can't be used in new projects. And they'll be
disabled on 9th April, 2025.

## [0.9.0](https://github.com/prantlf/ovai/compare/v0.8.0...v0.9.0) (2024-11-02)

### Features

* Recognise multimodalembedding model ([242709f](https://github.com/prantlf/ovai/commit/242709f58ddd8776eaef5476a861408bc93b11ce))
* Implement streaming mode ([e3989da](https://github.com/prantlf/ovai/commit/e3989dac05863436e21b3551cae8ba40ff5bee95))

## [0.8.0](https://github.com/prantlf/ovai/compare/v0.7.1...v0.8.0) (2024-11-02)

### Features

* Add new models and model versions released by Google this year ([0aeeff1](https://github.com/prantlf/ovai/commit/0aeeff19d14ea405d7544711a4a4d872a1603605))

### BREAKING CHANGES

Models chat-bison@001 and text-bison@001 were removed.
Google stopped supporting them.

## [0.7.1](https://github.com/prantlf/ovai/compare/v0.7.0...v0.7.1) (2024-06-09)

### Bug Fixes

* Base the image on busybox ([706c2fa](https://github.com/prantlf/ovai/commit/706c2fabb981c99d4ce3a667643e50752f87f591))

## [0.7.0](https://github.com/prantlf/ovai/compare/v0.6.0...v0.7.0) (2024-06-09)

### Features

* Recognise model gemini-1.0-pro-vision ([d042aec](https://github.com/prantlf/ovai/commit/d042aec0f10b6f20b3461e1000e9564f207a84e1))
* Support attaching images to generate and chat requests ([a3d7f15](https://github.com/prantlf/ovai/commit/a3d7f15162b77c571f3f18220e92250a3f8ec237))
* Recognise the model multimodalembedding@001 ([5c3aabe](https://github.com/prantlf/ovai/commit/5c3aabe4bcb66c83e3edbe590f2c2c8a9da45b03))

### Bug Fixes

* Use padded base64 for JWT parts ([8d01c1c](https://github.com/prantlf/ovai/commit/8d01c1ce0add49b07429956eb7e9447d3db92d9e))
* Do not log preflighting as an error ([de96929](https://github.com/prantlf/ovai/commit/de96929c08dd1bcb50d580a264b73a688c4c4191))

## [0.6.0](https://github.com/prantlf/ovai/compare/v0.5.0...v0.6.0) (2024-06-05)

### Features

* Allow forcing either IPV4 or IPV6 network connections only ([5c0a9e4](https://github.com/prantlf/ovai/commit/5c0a9e401ff450a77bb2a4b651eddd2cb604f431))

### Bug Fixes

* Allow scope and auth_uri empty in the account file ([54dbb29](https://github.com/prantlf/ovai/commit/54dbb29afc8396e64a26ad90e8ee28f09a348cc4))

## [0.5.0](https://github.com/prantlf/ovai/compare/v0.4.0...v0.5.0) (2024-06-04)

### Features

* Add chat, generate and embeddings purposes to the model info ([cbfe399](https://github.com/prantlf/ovai/commit/cbfe3996fb2128ca8b540e62c3d39487c599807e))
* Recognise model text-unicorn ([8e83e7e](https://github.com/prantlf/ovai/commit/8e83e7e2a84cdfcc170c08964435717e24a1cd68))

### Bug Fixes

* Recognise names of all PalLM embedding models ([b602248](https://github.com/prantlf/ovai/commit/b6022480dff301aed82d2dfbeaaa6018de230731))

## [0.4.0](https://github.com/prantlf/ovai/compare/v0.3.0...v0.4.0) (2024-06-04)

### Features

* Support /api/tags for listing available models ([fa5b1f0](https://github.com/prantlf/ovai/commit/fa5b1f01b37c78dcad2a4cf27681c9cb1524bc9f))
* Support /api/show for inspecting model information ([8a2c2bc](https://github.com/prantlf/ovai/commit/8a2c2bc7a0c4a37efd89282a58a77d5109d91f8c))

### Bug Fixes

* Propagate error message from requests proxied to ollama ([d2f59f1](https://github.com/prantlf/ovai/commit/d2f59f12f5c8e6d5a92c3339dcacf03f99daa639))

## [0.3.0](https://github.com/prantlf/ovai/compare/v0.2.2...v0.3.0) (2024-06-03)

### Features

* Allow forwarding requests to ollama for other than Vertex AI models ([8078a19](https://github.com/prantlf/ovai/commit/8078a19132a5b6f4e41083e944190f18f3300afd))

## [0.2.2](https://github.com/prantlf/ovai/compare/v0.2.1...v0.2.2) (2024-06-03)

### Bug Fixes

* Fix the message serlialisation in the chat response ([77f2180](https://github.com/prantlf/ovai/commit/77f218045bb2e2a533ed3538dcaf9d5ce5126ac2))

## [0.2.1](https://github.com/prantlf/ovai/compare/v0.2.0...v0.2.1) (2024-06-02)

### Bug Fixes

* Allow methods specific for each handler for CORS ([247f4c9](https://github.com/prantlf/ovai/commit/247f4c99580ba43e80fe4b32bc9c3e3f436b75d8))
* Do not close response body before reading it ([bb117c8](https://github.com/prantlf/ovai/commit/bb117c843d6fe2ab9fcdf2bde89cfef0ac8d0a54))

## [0.2.0](https://github.com/prantlf/ovai/compare/v0.1.5...v0.2.0) (2024-06-02)

### Features

* Support CORS preflighting (OPTIONS) ([0f1e6df](https://github.com/prantlf/ovai/commit/0f1e6df2e2107467be7f45eb1f00386f0ee08dc1))

### Bug Fixes

* Complete the CORS response headers ([5dd0a35](https://github.com/prantlf/ovai/commit/5dd0a35d469fe3fddc03cfd713cfab10a13fff9c))

## [0.1.5](https://github.com/prantlf/ovai/compare/v0.1.4...v0.1.5) (2024-05-12)

### Bug Fixes

* Fix automated publishing ([a6a9ee8](https://github.com/prantlf/ovai/commit/a6a9ee89b67f01260d8081406bfd46ac3344cf22))

## [0.1.4](https://github.com/prantlf/ovai/compare/v0.1.3...v0.1.4) (2024-05-12)

### Bug Fixes

* Fix automated publishing ([b17448e](https://github.com/prantlf/ovai/commit/b17448e604fe3288c8ac5ecea9c16dd254128851))

## [0.1.3](https://github.com/prantlf/ovai/compare/v0.1.2...v0.1.3) (2024-05-12)

### Bug Fixes

* Fix automated publishing ([c7830c3](https://github.com/prantlf/ovai/commit/c7830c3543eb793631bb54d7d12b8cde5fd6f37a))

## [0.1.2](https://github.com/prantlf/ovai/compare/v0.1.1...v0.1.2) (2024-05-12)

### Bug Fixes

* Fix automated publishing ([215c6ab](https://github.com/prantlf/ovai/commit/215c6aba15df20ccd0d5b5125a4724f4b0b4bb0d))

## [0.1.1](https://github.com/prantlf/ovai/compare/v0.1.0...v0.1.1) (2024-05-12)

### Bug Fixes

* Fix automated publishing ([a55e94d](https://github.com/prantlf/ovai/commit/a55e94dd37180f31086ea09d5e6f387a9eba46b0))

## [0.1.0](https://github.com/prantlf/ovai/compare/v0.0.1...v0.1.0) (2024-05-12)

### Features

* Initial release ([5c71469](https://github.com/prantlf/ovai/commit/5c71469f40862c3c3c25132e51ef3e93cdd041c2))

## 0.0.1 (2024-05-12)

Initial commit
