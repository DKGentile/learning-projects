# distributed-saga-coordinator

Build a fault-tolerant saga orchestrator that coordinates multi-service workflows with compensating actions.

## What You Have

- Minimal HTTP API at `cmd/orchestrator/main.go` that loads saga definitions from `config/sagas.json` and exposes `/sagas/execute`.
- Core domain models and plumbing in `internal/saga` with a no-op dispatcher and stdout logger.
- JSON configuration describing a sample “provision SaaS tenant” saga.

## What To Build

1. **Execution Engine**  
   - Implement a real `Dispatcher` that can execute saga steps concurrently while respecting dependency constraints (`internal/saga/coordinator.go`).  
   - Add retry, timeout, and compensation logic with deterministic rollback order.
2. **State Propagation**  
   - Design a payload contract so step outputs can feed subsequent steps.  
   - Persist execution traces (SQL/Redis/filesystem) and expose them via new HTTP endpoints.
3. **Observability Hooks**  
   - Replace `stdoutLogger` with structured logging or tracing (OpenTelemetry?).  
   - Emit metrics (histograms for step durations, counters for retries).

## Stretch Ideas

- Support dynamic saga registration via API + hot reloading.
- Add a temporal DSL (like JSON schema plus CEL) for assertions and guardrails.
- Embed a WASM sandbox to execute user-defined steps.

## Getting Started

```bash
go run ./cmd/orchestrator config/sagas.json
curl -X POST http://localhost:8080/sagas/execute \
  -H "Content-Type: application/json" \
  -d @- <<'JSON'
{
  "saga_name": "provision-saas-tenant",
  "trace_id": "demo-trace",
  "payload": {
    "tenant_id": "acme"
  }
}
JSON
```

Expect to receive a 422 error until you implement a dispatcher—use that as your first acceptance test.
