use std::collections::HashMap;
use std::sync::Arc;

use parking_lot::RwLock;
use thiserror::Error;
use uuid::Uuid;

#[derive(Clone, Default)]
pub struct Engine {
    index: Arc<RwLock<InMemoryIndex>>,
}

impl Engine {
    pub async fn ingest(&self, vector: Vec<f32>, metadata: HashMap<String, String>) -> String {
        let id = Uuid::new_v4().to_string();
        let entry = VectorEntry { id: id.clone(), vector, metadata };
        self.index.write().insert(entry);
        id
    }

    pub async fn query(
        &self,
        probe: Vec<f32>,
        k: usize,
        filter: HashMap<String, String>,
    ) -> Vec<ScoredVector> {
        let index = self.index.read();
        match index.search(&probe, k, filter) {
            Ok(results) => results,
            Err(err) => {
                eprintln!("search failed: {err:?}");
                vec![]
            }
        }
    }
}

#[derive(Default)]
struct InMemoryIndex {
    store: Vec<VectorEntry>,
}

impl InMemoryIndex {
    fn insert(&mut self, entry: VectorEntry) {
        self.store.push(entry);
    }

    fn search(
        &self,
        probe: &[f32],
        k: usize,
        filter: HashMap<String, String>,
    ) -> Result<Vec<ScoredVector>, SearchError> {
        if self.store.is_empty() {
            return Ok(vec![]);
        }

        let filtered = self
            .store
            .iter()
            .filter(|entry| filter.iter().all(|(key, value)| entry.metadata.get(key) == Some(value)))
            .collect::<Vec<_>>();

        if filtered.is_empty() {
            return Ok(vec![]);
        }

        let mut scored = filtered
            .into_iter()
            .map(|entry| {
                let distance = cosine_distance(probe, &entry.vector)?;
                Ok(ScoredVector {
                    id: entry.id.clone(),
                    score: 1.0 - distance,
                    metadata: entry.metadata.clone(),
                })
            })
            .collect::<Result<Vec<_>, SearchError>>()?;

        scored.sort_by(|a, b| b.score.partial_cmp(&a.score).unwrap_or(std::cmp::Ordering::Equal));
        scored.truncate(k);
        Ok(scored)
    }
}

fn cosine_distance(lhs: &[f32], rhs: &[f32]) -> Result<f32, SearchError> {
    if lhs.len() != rhs.len() {
        return Err(SearchError::DimensionMismatch {
            expected: lhs.len(),
            got: rhs.len(),
        });
    }

    let mut dot = 0.0f32;
    let mut lhs_norm = 0.0f32;
    let mut rhs_norm = 0.0f32;

    for (l, r) in lhs.iter().zip(rhs.iter()) {
        dot += l * r;
        lhs_norm += l * l;
        rhs_norm += r * r;
    }

    let denom = (lhs_norm.sqrt() * rhs_norm.sqrt()).max(f32::MIN_POSITIVE);
    let similarity = dot / denom;
    if !similarity.is_finite() {
        return Err(SearchError::NumericalInstability);
    }

    Ok(1.0 - similarity)
}

#[derive(Debug, Error)]
enum SearchError {
    #[error("vector dimensionality mismatch (expected {expected}, got {got})")]
    DimensionMismatch { expected: usize, got: usize },
    #[error("cosine distance returned NaN")]
    NumericalInstability,
}

#[derive(Clone, Debug)]
pub struct ScoredVector {
    pub id: String,
    pub score: f32,
    pub metadata: HashMap<String, String>,
}

#[derive(Clone)]
struct VectorEntry {
    id: String,
    vector: Vec<f32>,
    metadata: HashMap<String, String>,
}
