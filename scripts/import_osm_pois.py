import argparse
import json
from pathlib import Path


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("source")
    parser.add_argument("destination")
    parser.add_argument("--coverage", required=True)
    args = parser.parse_args()
    source = json.loads(Path(args.source).read_text(encoding="utf-8"))
    observed = source.get("osm3s", {}).get("timestamp_osm_base")
    if not observed or not isinstance(source.get("elements"), list):
        raise SystemExit("Input must be an Overpass JSON response with an OSM timestamp.")
    items = {}
    for element in source["elements"]:
        tags = element.get("tags") or {}
        kind = "museum" if tags.get("tourism") == "museum" else tags.get("amenity")
        if kind not in ("museum", "cinema", "theatre") or not tags.get("name") or tags.get("access") in ("private", "no"):
            continue
        osm_type, osm_id = element.get("type"), element.get("id")
        point = element if osm_type == "node" else element.get("center") or {}
        lat, lon = point.get("lat"), point.get("lon")
        if osm_type not in ("node", "way", "relation") or not isinstance(osm_id, int):
            continue
        if not isinstance(lat, (int, float)) or not isinstance(lon, (int, float)) or not (-90 <= lat <= 90 and -180 <= lon <= 180):
            continue
        identifier = f"osm:{osm_type}:{osm_id}"
        items[identifier] = {
            "id": identifier,
            "name": tags["name"],
            "kind": kind,
            "latitude": lat,
            "longitude": lon,
            "source_url": f"https://www.openstreetmap.org/{osm_type}/{osm_id}",
        }
    result = {
        "attribution": "© OpenStreetMap contributors, ODbL 1.0",
        "source": "https://www.openstreetmap.org/copyright",
        "observed_at": observed,
        "coverage": args.coverage,
        "items": [items[key] for key in sorted(items)],
    }
    Path(args.destination).write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Imported {len(items)} places as of {observed}.")


if __name__ == "__main__":
    main()
