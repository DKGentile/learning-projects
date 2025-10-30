mod engine;
mod models;

use axum::{
    extract::State,
    routing::{get, post},
    Json, Router,
};
use anyhow::Result;
use engine::Engine;
use models::{IngestRequest, QueryRequest};
use tokio::signal;
use std::net::SocketAddr;

#[tokio::main]
async fn main() -> Result<()> {
    color_eyre::install().ok();

    let engine = Engine::default();
    let app_state = AppState { engine };

    let app = Router::new()
        .route("/healthz", get(health))
        .route("/vectors", post(ingest))
        .route("/query", post(search))
        .with_state(app_state);

    let addr: SocketAddr = "127.0.0.1:8081".parse().expect("valid socket address");
    println!("vector-search-service listening on {}", addr);

    axum::serve(tokio::net::TcpListener::bind(addr).await?, app)
        .with_graceful_shutdown(shutdown_signal())
        .await?;

    Ok(())
}

async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c().await.expect("failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to install signal handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }

    println!("shutdown signal received");
}

#[derive(Clone)]
struct AppState {
    engine: Engine,
}

async fn health() -> &'static str {
    "ok"
}

async fn ingest(State(state): State<AppState>, Json(payload): Json<IngestRequest>) -> Json<serde_json::Value> {
    let id = state.engine.ingest(payload.vector, payload.metadata).await;

    Json(serde_json::json!({
        "id": id,
        "status": "queued"
    }))
}

async fn search(
    State(state): State<AppState>,
    Json(payload): Json<QueryRequest>,
) -> Json<serde_json::Value> {
    let results = state
        .engine
        .query(payload.vector, payload.k, payload.filter)
        .await;

    Json(serde_json::json!({
        "results": results
    }))
}
