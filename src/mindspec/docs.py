import re
import subprocess
from pathlib import Path

class DocParser:
    def __init__(self, workspace):
        self.workspace = workspace

    def parse_glossary(self):
        glossary_path = self.workspace.get_glossary_path()
        if not glossary_path.exists():
            return {}

        terms = {}
        with open(glossary_path, 'r') as f:
            content = f.read()

        # Match markdown table rows: | **Term** | [Link](Target) |
        matches = re.finditer(r'\|\s*\*\*([^*]+)\*\*\s*\|\s*\[[^\]]+\]\(([^)]+)\)\s*\|', content)
        for match in matches:
            term = match.group(1).strip()
            target = match.group(2).strip()
            terms[term] = target
        return terms

    def check_health(self, strict=False):
        docs_dir = self.workspace.get_docs_dir()
        glossary = self.parse_glossary()
        project_root = self.workspace.find_project_root()

        report = {
            "docs_dir_exists": docs_dir.exists(),
            "glossary_exists": self.workspace.get_glossary_path().exists(),
            "term_count": len(glossary),
            "broken_links": [],
            "warnings": []
        }

        for term, target in glossary.items():
            # Handle relative paths in links (strip anchors for existence check)
            path_part = target.split('#')[0]
            target_path = project_root / path_part
            if not target_path.exists():
                report["broken_links"].append(f"{term} -> {target}")

        # Domain structure checks (warn by default, error in strict mode)
        domains_dir = docs_dir / "domains"
        context_map = docs_dir / "context-map.md"
        expected_domains = ["core", "context-system", "workflow"]
        domain_files = ["overview.md", "architecture.md", "interfaces.md", "runbook.md"]

        if not domains_dir.exists():
            report["warnings"].append("docs/domains/ directory not found")
        else:
            for domain in expected_domains:
                domain_dir = domains_dir / domain
                if not domain_dir.exists():
                    report["warnings"].append(f"docs/domains/{domain}/ not found")
                else:
                    for f in domain_files:
                        if not (domain_dir / f).exists():
                            report["warnings"].append(f"docs/domains/{domain}/{f} not found")

        if not context_map.exists():
            report["warnings"].append("docs/context-map.md not found")

        # In strict mode, warnings become errors
        if strict and report["warnings"]:
            report["strict_failures"] = list(report["warnings"])

        # Beads hygiene checks
        beads_report = self.check_beads_hygiene(project_root)
        report["beads"] = beads_report

        return report

    def check_beads_hygiene(self, project_root):
        """Check Beads directory health: existence, durable state, runtime leaks."""
        beads_dir = project_root / ".beads"
        report = {
            "dir_exists": beads_dir.exists(),
            "durable_state": False,
            "tracked_runtime_artifacts": [],
        }

        if not beads_dir.exists():
            return report

        # Check for durable state files
        durable_files = ["issues.jsonl", "config.yaml", "metadata.json"]
        found_durable = [f for f in durable_files if (beads_dir / f).exists()]
        report["durable_state"] = len(found_durable) > 0
        report["durable_files_found"] = found_durable

        # Check for runtime artifacts tracked by git
        runtime_patterns = {
            "bd.sock", "daemon.lock", "daemon.log", "daemon.pid",
            "sync-state.json", "last-touched", ".local_version",
            "db.sqlite", "bd.db", "redirect", ".sync.lock",
        }
        runtime_extensions = {".db", ".db-wal", ".db-shm", ".db-journal"}

        try:
            result = subprocess.run(
                ["git", "ls-files", ".beads/"],
                capture_output=True, text=True, cwd=project_root
            )
            tracked_files = result.stdout.strip().splitlines() if result.stdout.strip() else []
            for tracked in tracked_files:
                filename = Path(tracked).name
                if filename in runtime_patterns:
                    report["tracked_runtime_artifacts"].append(tracked)
                elif any(filename.endswith(ext) for ext in runtime_extensions):
                    report["tracked_runtime_artifacts"].append(tracked)
        except FileNotFoundError:
            pass  # git not available

        return report
