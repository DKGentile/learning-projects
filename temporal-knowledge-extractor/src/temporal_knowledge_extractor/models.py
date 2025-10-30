from __future__ import annotations

from datetime import datetime
from typing import Dict, Iterable, List, Optional, Tuple

from pydantic import BaseModel, Field, validator


class EventEnvelope(BaseModel):
    """Normalized representation of an incoming event."""

    event_type: str = Field(..., description="Semantic type of the event, e.g. issue.created.")
    occurred_at: datetime = Field(..., description="Timestamp assigned by the upstream producer.")
    actor: str = Field(..., description="Primary entity performing the action.")
    subject: str = Field(..., description="Entity that the action was taken on.")
    attributes: Dict[str, str] = Field(default_factory=dict)

    @validator("event_type", "actor", "subject")
    def lowercase_identifiers(cls, value: str) -> str:
        return value.strip().lower()


class Relationship(BaseModel):
    """Edge in the temporal knowledge graph."""

    source: str
    target: str
    relationship_type: str
    first_observed_at: datetime
    last_observed_at: datetime
    weight: float = 1.0
    context: Dict[str, str] = Field(default_factory=dict)

    def touch(self, timestamp: datetime, weight_delta: float = 1.0) -> None:
        """Update timestamps and weight when the relationship is observed again."""
        if timestamp > self.last_observed_at:
            self.last_observed_at = timestamp
        if timestamp < self.first_observed_at:
            self.first_observed_at = timestamp
        self.weight += weight_delta


class TemporalGraph(BaseModel):
    """Container for nodes and relationships observed over time."""

    nodes: Dict[str, Dict[str, str]] = Field(default_factory=dict)
    edges: Dict[Tuple[str, str, str], Relationship] = Field(default_factory=dict)

    def upsert_node(self, identifier: str, attributes: Optional[Dict[str, str]] = None) -> None:
        bucket = self.nodes.setdefault(identifier, {})
        if attributes:
            bucket.update(attributes)

    def upsert_edge(self, relationship: Relationship) -> None:
        key = (relationship.source, relationship.target, relationship.relationship_type)
        if key in self.edges:
            existing = self.edges[key]
            existing.touch(relationship.last_observed_at)
            existing.context.update(relationship.context)
        else:
            self.edges[key] = relationship

    def relationships(self) -> Iterable[Relationship]:
        return self.edges.values()

    def neighborhood(self, node: str) -> List[Relationship]:
        return [rel for rel in self.relationships() if rel.source == node or rel.target == node]
