# Coverage: owned implementation files

Unit tests and in-process HTTP/gRPC integrations; external-process E2E is reported separately. Generated protobuf, launchers and unrelated stubs are not included.

| Scope | Covered statements | Total | Coverage |
|---|---:|---:|---:|
| HTTP API | 243 | 263 | 92.4% |
| auth store | 109 | 122 | 89.3% |
| config bridge | 71 | 74 | 95.9% |
| demo orders | 69 | 84 | 82.1% |
| gateway gRPC adapter | 72 | 88 | 81.8% |
| mapping RPC | 17 | 24 | 70.8% |
| planning | 125 | 140 | 89.3% |
| provider RPC | 74 | 86 | 86.0% |
| providers | 425 | 485 | 87.6% |
| route assembly | 361 | 452 | 79.9% |
| route handoff | 8 | 18 | 44.4% |
| sessions | 58 | 66 | 87.9% |
| **Total** | **1632** | **1902** | **85.8%** |

Required aggregate coverage: 80%. This measures statements, not complete behavioral coverage.
