import json
import sys
from collections import Counter
from pathlib import Path

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt


def load(path):
    return json.loads(path.read_text(encoding="utf-8"))


def main():
    folder = Path(sys.argv[1] if len(sys.argv) > 1 else "docs/evidence/2026-10-28")
    groups = [
        ("Frontend demo", load(folder / "frontend-runs.json")["runs"]),
        ("Provider gRPC probes", load(folder / "provider-runs.json")["runs"]),
        ("Gateway route probe", load(folder / "gateway-runs.json")["runs"]),
    ]
    colors = {"passed": "#20826b", "failed": "#c4514f", "partial": "#c18a25", "blocked": "#788496"}
    fig, ax = plt.subplots(figsize=(9, 4.8))
    for row, (name, runs) in enumerate(groups):
        counts = Counter(run["status"] for run in runs)
        left = 0
        for status in ("passed", "partial", "failed", "blocked"):
            count = counts[status]
            if count:
                ax.barh(row, count, left=left, height=0.54, color=colors[status], label=status.title() if row == 0 else None)
                ax.text(left + count / 2, row, str(count), ha="center", va="center", color="white", fontweight="bold")
                left += count
        ax.text(left + 0.12, row, f"n={len(runs)}", va="center", fontsize=10)
    ax.set_yticks(range(len(groups)), [g[0] for g in groups])
    ax.invert_yaxis()
    ax.set_xlabel("Number of individual checks")
    ax.set_xlim(0, max(len(g[1]) for g in groups) + 1.2)
    ax.set_title("Observed checks — 28 October 2026", loc="left", fontweight="bold", fontsize=15)
    handles = [plt.Rectangle((0, 0), 1, 1, color=colors[k]) for k in ("passed", "partial", "failed", "blocked")]
    ax.legend(handles, ["Passed", "Partial", "Failed", "Blocked"], ncol=4, loc="lower right", frameon=False)
    ax.spines[["top", "right", "left"]].set_visible(False)
    ax.tick_params(axis="y", length=0)
    fig.text(0.12, 0.02, "Different modes are not comparable. Provider checks are component checks; no integrated journey passed.", fontsize=9, color="#56606e")
    fig.tight_layout(rect=(0, 0.06, 1, 1))
    fig.savefig(folder / "scenario-outcomes.svg", bbox_inches="tight")
    fig.savefig(folder / "scenario-outcomes.png", dpi=160, bbox_inches="tight")
    plt.close(fig)

    provider = load(folder / "provider-runs.json")["runs"]
    fig, ax = plt.subplots(figsize=(9, 4.8))
    names = [run["case"].replace("_", " ") for run in provider]
    values = [run["elapsed_ms"] / 1000 for run in provider]
    bars = ax.barh(names, values, color="#367e9f")
    ax.invert_yaxis()
    ax.bar_label(bars, labels=[f"{v:.2f} s" for v in values], padding=4, fontsize=9)
    ax.set_xlim(0, max(values, default=0) * 1.25 + 0.1)
    ax.set_xlabel("Elapsed time per check (seconds)")
    ax.set_title("Provider gRPC probes — single local run", loc="left", fontweight="bold", fontsize=15)
    ax.spines[["top", "right", "left"]].set_visible(False)
    ax.tick_params(axis="y", length=0)
    fig.text(0.12, 0.02, "One observation per check; missing credentials and network errors affect timings. Not a benchmark.", fontsize=9, color="#56606e")
    fig.tight_layout(rect=(0, 0.06, 1, 1))
    fig.savefig(folder / "provider-latency.svg", bbox_inches="tight")
    fig.savefig(folder / "provider-latency.png", dpi=160, bbox_inches="tight")
    plt.close(fig)


if __name__ == "__main__":
    main()
