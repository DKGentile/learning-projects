# vector-search-service

Design and ship a production-grade vector similarity API in Rust. The current code stands up an Axum server with very naive storage—your job is to evolve it into a low-latency ANN (approximate nearest neighbor) service.

## What You Have

- HTTP handlers in `src/main.rs` with `/vectors` for ingestion and `/query` for search.
- `Engine` abstraction (`src/engine.rs`) backed by an in-memory cosine similarity search.
- JSON request models in `src/models.rs`.

## Mission Objectives

1. **Indexing Strategy**  
   - Replace the in-memory Vec with a proper ANN index (HNSW, IVF+PQ, product quantization, etc.).  
   - Plumb configurable distance metrics and dimensionality validation.
2. **Persistence + Snapshotting**  
   - Add background tasks to persist the index to disk and restore it on boot.  
   - Support collection-level TTLs or metadata-based pruning.
3. **Observability & Testing**  
   - Integrate tracing (`tracing` crate) with request IDs and timing.  
   - Build load-testing scripts (Locust/k6) and unit tests that cover numerical stability.

## Stretch Ideas

- Multi-tenant separation with API keys and per-tenant quotas.
- Streaming ingestion via Kafka / NATS + backpressure handling.
- Generate typed SDKs for TypeScript/Python using OpenAPI.

## Quick Check

```bash
cargo run

curl -X POST http://localhost:8081/vectors \
  -H "Content-Type: application/json" \
  -d '{"vector":[0.1,0.2,0.3],"metadata":{"domain":"demo"}}'

curl -X POST http://localhost:8081/query \
  -H "Content-Type: application/json" \
  -d '{"vector":[0.1,0.2,0.3],"k":3}'
```

Start by writing property tests for `cosine_distance` to lock in correctness before swapping in a new index implementation.
