# temporal-knowledge-extractor

Explore temporal reasoning by turning event streams into a knowledge graph you can query, visualize, and analyze.

## Current Capabilities

- `KnowledgePipeline` (`src/temporal_knowledge_extractor/pipeline.py`) ingests events, builds a graph, and exports GraphML.
- CLI (`src/temporal_knowledge_extractor/cli.py`) to ingest JSON files or fabricate synthetic datasets.
- Sample dataset in `data/sample-events.json`.

## Core Objectives

1. **Streaming Ingestion**  
   - Swap the batch ingestion for a streaming source (Kafka, Kinesis, websockets, etc.).  
   - Apply windowed aggregations to detect bursts or long-running relationships.
2. **Temporal Querying**  
   - Implement temporal predicates (e.g., “return relationships active between T1 and T2”).  
   - Provide a DSL or GraphQL layer for querying the temporal graph.
3. **Analytics Layer**  
   - Surface community detection, centrality over time, or anomaly detection signals.  
   - Expose dashboards/notebooks showing trend lines.

## Stretch Ideas

- Integrate with Neo4j or Dgraph for persistence and advanced querying.
- Add transformers/LLMs that summarize event clusters into human-readable narratives.
- Ship an interactive UI (Streamlit/React) for exploring the time-travel graph.

## Kickoff

```bash
python -m venv .venv
source .venv/bin/activate  # Windows: .venv\Scripts\activate
pip install -e .

tke ingest data/sample-events.json --export graph.graphml
```

Feed the exported GraphML into Gephi or Cytoscape to visualize the relationships. Start by designing tests for the pipeline’s aggregation logic before layering on streaming systems.
