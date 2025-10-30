"""Temporal Knowledge Extractor package.

The package exposes primitives for ingesting event streams, projecting them into
temporal knowledge graphs, and querying emergent relationships over time.
"""

from .pipeline import KnowledgePipeline
from .models import EventEnvelope, Relationship, TemporalGraph

__all__ = [
    "KnowledgePipeline",
    "EventEnvelope",
    "Relationship",
    "TemporalGraph",
]
