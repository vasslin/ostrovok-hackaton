import argparse
import datetime
import hashlib
import json
from pathlib import Path
import subprocess
import time


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--client", default="/tmp/godnota-provider-search")
    parser.add_argument("--address", default="127.0.0.1:50158")
    parser.add_argument("--output", required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parent.parent
    output = Path(args.output)
    output.mkdir(parents=True, exist_ok=True)
    today = datetime.datetime.now().astimezone().date()
    departure = today + datetime.timedelta(days=30)
    trip = json.loads((root / "configs/examples/trip_request.kazan_weekend.json").read_text(encoding="utf-8"))
    trip["dates"].update(departure_date=departure.isoformat(), return_date=(departure + datetime.timedelta(days=2)).isoformat(), nights=2)
    ai = {"origin": "Москва", "destination": "Казань", "departure_date": departure.isoformat(), "return_date": (departure + datetime.timedelta(days=2)).isoformat(), "adults": 2, "children_ages": [], "budget": 40000, "currency": "RUB", "preferences": ["рядом с центром"]}
    frontend = {"origin": "Москва", "destination": "Казань", "startDate": departure.isoformat(), "nights": 2, "adults": 2, "childrenAges": [], "budgetMinor": 4000000, "currency": "RUB", "pets": True}
    inputs = {"provider-trip.json": trip, "ai-config.json": ai, "frontend-criteria.json": frontend}
    for name, data in inputs.items():
        (output / name).write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    base = [args.client, "-address", args.address, "-resolved", str(root / "configs/examples/resolved_trip.kazan_weekend.json")]
    results = []

    def execute(case, extra, expected_missing=None):
        started = time.monotonic()
        process = subprocess.run(base + extra, cwd=root, capture_output=True, text=True, timeout=30)
        elapsed = round((time.monotonic() - started) * 1000)
        file = f"provider-{case}.json"
        if process.returncode:
            file = f"provider-{case}.txt"
            (output / file).write_text(process.stderr, encoding="utf-8")
            status = "passed" if case == "past_date" and "InvalidArgument" in process.stderr else "failed"
            results.append({"case": case, "status": status, "elapsed_ms": elapsed, "evidence": file})
            return
        payload = json.loads(process.stdout)
        (output / file).write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        offers = payload.get("offers", [])
        issues = [item.get("code") for item in payload.get("issues", [])]
        if expected_missing:
            status = "passed" if expected_missing in payload.get("missing", []) and not offers else "failed"
        elif case == "past_date":
            status = "failed"
        elif "credentials_missing" in issues and not offers:
            status = "blocked"
        elif offers and issues:
            status = "partial"
        elif offers:
            status = "passed"
        elif issues:
            status = "blocked"
        else:
            status = "failed"
        results.append({"case": case, "status": status, "elapsed_ms": elapsed, "offers": len(offers), "issues": issues, "missing": payload.get("missing", []), "evidence": file})

    trip_path = str(output / "provider-trip.json")
    for kind in ("transport", "hotel", "activity"):
        execute(kind, ["-kind", kind, "-trip", trip_path])
    past = dict(trip)
    past["dates"] = {"departure_date": "2026-01-01", "return_date": "2026-01-03", "nights": 2}
    (output / "past-date-input.json").write_text(json.dumps(past, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    execute("past_date", ["-kind", "transport", "-trip", str(output / "past-date-input.json")])
    execute("ai_config", ["-kind", "transport", "-format", "ai_config", "-trip", str(output / "ai-config.json")], "/budget/scope")
    execute("frontend_criteria", ["-kind", "transport", "-format", "frontend_criteria", "-trip", str(output / "frontend-criteria.json")], "/travelers/pets/kind")
    fingerprint = hashlib.sha256()
    sources = [root / "api/provider/v1/provider.proto", root / "configs/poi/kazan.json"]
    for directory in ("provider-service", "providerkit", "internal/providers", "pkg/generated/provider/v1"):
        sources.extend(p for p in (root / directory).rglob("*") if p.is_file())
    for path in sorted(set(sources)):
        fingerprint.update(str(path.relative_to(root)).encode())
        fingerprint.update(path.read_bytes())
    commit = subprocess.check_output(["git", "rev-parse", "--short", "HEAD"], cwd=root, text=True).strip()
    record = {"date": today.isoformat(), "mode": "provider-service gRPC component probes; no integrated frontend", "base_commit": commit, "working_tree": "uncommitted changes", "source_sha256": fingerprint.hexdigest(), "runs": results}
    (output / "provider-runs.json").write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(results, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
