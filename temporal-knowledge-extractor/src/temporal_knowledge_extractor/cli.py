from __future__ import annotations

from datetime import datetime, timedelta
from pathlib import Path
from typing import Optional

import typer
from rich import print as rprint
import json

from .models import EventEnvelope
from .pipeline import KnowledgePipeline

app = typer.Typer(help="Temporal knowledge extraction sandbox")


@app.command()
def ingest(path: Path, export: Optional[Path] = typer.Option(None, help="Optional GraphML export path")) -> None:
    """Ingest a JSON file of events and build the temporal graph."""
    pipeline = KnowledgePipeline()
    events = list(KnowledgePipeline.load_events(path))
    pipeline.ingest(events)
    pipeline.build_graph()
    pipeline.summarize()
    if export:
        pipeline.export_graphml(export)


@app.command()
def fabricate(
    count: int = typer.Option(20, min=1, help="Number of synthetic events to generate"),
    seed: Optional[int] = typer.Option(None, help="Random seed for reproducibility"),
    output: Path = typer.Option(Path("synthetic-events.json")),
) -> None:
    """Generate a synthetic dataset to experiment with temporal reasoning."""
    import random

    rng = random.Random(seed)
    teams = ["atlas", "apollo", "selene"]
    services = ["payments", "notifications", "analytics", "inventory"]
    actions = ["deployed", "scaled", "incident_opened", "incident_resolved"]

    events = []
    base_time = datetime.utcnow()

    for i in range(count):
        actor = rng.choice(teams)
        subject = rng.choice(services)
        action = rng.choice(actions)
        timestamp = base_time.replace(microsecond=0) + timedelta(seconds=int(rng.random() * 72 * 3600))
        events.append(
            EventEnvelope(
                event_type=f"service.{action}",
                occurred_at=timestamp,
                actor=actor,
                subject=subject,
                attributes={"source": "fabricated", "run": str(i)},
            )
        )

    payload = [event.model_dump(mode="json") for event in events]
    output.write_text(json.dumps(payload, indent=2))
    rprint(f"Synthetic events written to [bold]{output}[/bold]")


if __name__ == "__main__":
    app()
