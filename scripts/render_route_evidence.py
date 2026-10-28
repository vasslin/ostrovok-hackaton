import json
import sys
from pathlib import Path
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt

folder = Path(sys.argv[1] if len(sys.argv) > 1 else "docs/evidence/2026-10-28")
record = json.loads((folder / "route-booking-evidence.json").read_text(encoding="utf-8"))
demo = record["backend_demo"]
labels = ["Один город", "Два города"]
values = [demo["single_city_candidates"], demo["two_city_candidates"]]
fig, ax = plt.subplots(figsize=(8.8, 4.4))
bars = ax.barh(labels, values, color=["#31826f", "#5189b9"], height=0.48)
ax.invert_yaxis()
for bar, value in zip(bars, values):
    ax.text(value + .15, bar.get_y() + bar.get_height() / 2, str(value), va="center", fontweight="bold")
ax.set_xlim(0, max(values) + 2)
ax.set_xlabel("Число найденных вариантов")
ax.set_title("Сборка маршрутов: один автоматический backend-прогон", loc="left", fontweight="bold")
ax.spines[["top", "right", "left"]].set_visible(False)
ax.tick_params(axis="y", length=0)
fig.text(.13, .10, "Синтетические данные · один тестовый сеанс · не пользователи и не реальные бронирования", fontsize=9, color="#48566a")
fig.text(.13, .045, "Демозаказ: создание, повтор, имитация оплаты и отмены проверены; AI и frontend не участвовали.", fontsize=9, color="#48566a")
fig.subplots_adjust(left=.2, right=.95, bottom=.28, top=.85)
for ext in ("svg", "png"):
    fig.savefig(folder / f"route-demo-outcomes.{ext}", dpi=180, facecolor="white")
