use serde::Deserialize;
use std::collections::HashMap;

#[derive(Debug, Deserialize)]
pub struct IngestRequest {
    pub vector: Vec<f32>,
    #[serde(default)]
    pub metadata: HashMap<String, String>,
}

#[derive(Debug, Deserialize)]
pub struct QueryRequest {
    pub vector: Vec<f32>,
    #[serde(default = "default_k")]
    pub k: usize,
    #[serde(default)]
    pub filter: HashMap<String, String>,
}

fn default_k() -> usize {
    5
}
