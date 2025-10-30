from __future__ import annotations

import json
from pathlib import Path
from typing import Iterable, List, Sequence

import networkx as nx
from rich.console import Console
from rich.table import Table

from .models import EventEnvelope, Relationship, TemporalGraph

console = Console()


class KnowledgePipeline:
    """High-level façade for ingesting events and surfacing temporal insights."""

    def __init__(self) -> None:
        self._events: List[EventEnvelope] = []
        self._graph = TemporalGraph()

    def ingest(self, events: Sequence[EventEnvelope]) -> None:
        self._events.extend(sorted(events, key=lambda e: e.occurred_at))

    def build_graph(self) -> TemporalGraph:
        for event in self._events:
            self._graph.upsert_node(event.actor, {"type": "actor"})
            self._graph.upsert_node(event.subject, {"type": event.event_type})

            relationship = Relationship(
                source=event.actor,
                target=event.subject,
                relationship_type=event.event_type,
                first_observed_at=event.occurred_at,
                last_observed_at=event.occurred_at,
                weight=1.0,
                context=event.attributes,
            )
            self._graph.upsert_edge(relationship)

        return self._graph

    def export_graphml(self, destination: Path) -> None:
        graph = nx.DiGraph()
        for node, attrs in self._graph.nodes.items():
            graph.add_node(node, **attrs)
        for relationship in self._graph.relationships():
            graph.add_edge(
                relationship.source,
                relationship.target,
                key=relationship.relationship_type,
                weight=relationship.weight,
                first_observed_at=relationship.first_observed_at.isoformat(),
                last_observed_at=relationship.last_observed_at.isoformat(),
                **relationship.context,
            )

        nx.write_graphml(graph, destination)
        console.log(f"Graph exported to [bold]{destination}[/bold]")

    def summarize(self, top_k: int = 5) -> None:
        if not self._graph.edges:
            console.print("No relationships discovered yet. Run build_graph() first.")
            return

        table = Table("Relationship", "Observations", "Active Window")
        for relationship in sorted(
            self._graph.relationships(), key=lambda r: r.weight, reverse=True
        )[:top_k]:
            window = f"{relationship.first_observed_at.date()} → {relationship.last_observed_at.date()}"
            descriptor = f"{relationship.source} -[{relationship.relationship_type}]-> {relationship.target}"
            table.add_row(descriptor, f"{relationship.weight:.1f}", window)

        console.print(table)

    @staticmethod
    def load_events(path: Path) -> Iterable[EventEnvelope]:
        raw = json.loads(path.read_text())
        for event in raw:
            yield EventEnvelope.model_validate(event)
