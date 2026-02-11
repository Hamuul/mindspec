import sys
import click
from .workspace import Workspace
from .docs import DocParser

@click.group()
def cli():
    """Mindspec: Spec-Driven Development and Self-Documentation System."""
    pass

@cli.command()
def doctor():
    """Check the health of the current workspace documentation."""
    ws = Workspace()
    root = ws.find_project_root()
    parser = DocParser(ws)
    health = parser.check_health()
    exit_code = 0

    click.echo(f"Workspace Root: {root}")
    click.echo(f"Docs Directory: {'[OK]' if health['docs_dir_exists'] else '[MISSING]'}")
    click.echo(f"GLOSSARY.md: {'[OK]' if health['glossary_exists'] else '[MISSING]'} ({health['term_count']} terms)")

    if health['broken_links']:
        click.echo("\nBroken Links in Glossary:")
        for link in health['broken_links']:
            click.echo(f"  - {link}")
    elif health['glossary_exists']:
        click.echo("Glossary links verified.")

    # Beads hygiene
    beads = health.get("beads", {})
    click.echo("")
    if not beads.get("dir_exists"):
        click.echo("Beads: [MISSING] .beads/ directory not found")
        click.echo("  Run `beads init` to initialize Beads in this repo.")
    else:
        click.echo("Beads: [OK] .beads/ directory exists")
        if beads.get("durable_state"):
            found = ", ".join(beads.get("durable_files_found", []))
            click.echo(f"  Durable state: [OK] ({found})")
        else:
            click.echo("  Durable state: [MISSING] No durable state files found (issues.jsonl, config.yaml, metadata.json)")

        tracked_runtime = beads.get("tracked_runtime_artifacts", [])
        if tracked_runtime:
            exit_code = 1
            click.echo("  Runtime artifacts: [ERROR] The following runtime files are tracked by git:")
            for artifact in tracked_runtime:
                click.echo(f"    - {artifact}")
            click.echo("  Add these to .beads/.gitignore and run `git rm --cached <file>`")
        else:
            click.echo("  Runtime artifacts: [OK] None tracked by git")

    sys.exit(exit_code)

@cli.group()
def context():
    """Context pack management."""
    pass

@context.command(name="init")
@click.option("--spec", required=True, help="Specification ID (e.g., 001)")
def context_init(spec):
    """Initialize a context pack for a specific specification."""
    click.echo(f"Initializing Context Pack for Spec: {spec} (Prototype)")
    # Future: Logic to pull doc sections and memory
    click.echo("Created: docs/specs/001-skeleton/context-pack.md")

if __name__ == "__main__":
    cli()
