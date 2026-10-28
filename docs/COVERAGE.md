# Coverage: owned implementation files

Unit tests and in-process HTTP/gRPC integrations; external-process E2E is reported separately. Generated protobuf, launchers and unrelated stubs are not included.

| Scope | Covered statements | Total | Coverage |
|---|---:|---:|---:|
| HTTP API | 220 | 229 | 96.1% |
| auth store | 109 | 122 | 89.3% |
| config bridge | 71 | 74 | 95.9% |
| gateway gRPC adapter | 61 | 77 | 79.2% |
| mapping RPC | 17 | 24 | 70.8% |
| planning | 91 | 96 | 94.8% |
| provider RPC | 16 | 19 | 84.2% |
| providers | 262 | 309 | 84.8% |
| route handoff | 5 | 8 | 62.5% |
| sessions | 75 | 78 | 96.2% |
| **Total** | **927** | **1036** | **89.5%** |

Required aggregate coverage: 80%. This measures statements, not complete behavioral coverage.
