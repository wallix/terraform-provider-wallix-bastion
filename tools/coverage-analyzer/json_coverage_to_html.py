import json
import sys
from pathlib import Path

def html_escape(text):
    return (text or "").replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")

def render_schema(schema_name, schema):
    html = f'<details><summary><b>{html_escape(schema_name)}</b> '
    html += f'({schema.get("covered_properties", 0)}/{schema.get("total_properties", 0)} couvertes)'
    html += '</summary>\n<ul>'
    props = schema.get("properties", {})
    if props:
        for prop, pdata in props.items():
            covered = pdata.get("covered", False)
            color = "#4caf50" if covered else "#f44336"
            html += f'<li><span style="color:{color}">{html_escape(prop)}</span></li>'
    if schema.get("missing_properties"):
        html += '<li><b>Manquantes:</b> ' + ", ".join(html_escape(p) for p in schema["missing_properties"]) + '</li>'
    html += '</ul></details>\n'
    return html

def main():
    input_path = Path("coverage-report.json")
    output_path = Path("coverage-report.html")
    if not input_path.exists():
        print("coverage-report.json non trouvé.")
        sys.exit(1)
    with open(input_path, "r") as f:
        data = json.load(f)

    html = [
        "<!DOCTYPE html><html><head><meta charset='utf-8'>",
        "<title>Rapport de Couverture API</title>",
        "<style>body{font-family:sans-serif;} summary{cursor:pointer;} .ok{color:#4caf50;} .ko{color:#f44336;}</style>",
        "</head><body>",
        "<h1>Rapport de Couverture API</h1>"
    ]

    # Résumé global
    if "coverage_percent" in data:
        html.append(f"<h2>Couverture globale : {data['coverage_percent']:.2f}%</h2>")
    if "total_endpoints" in data and "covered_endpoints" in data:
        html.append(f"<p>Endpoints couverts : {data['covered_endpoints']} / {data['total_endpoints']}</p>")

    # Schémas
    schema_analysis = data.get("schema_analysis", {})
    schema_details = schema_analysis.get("schema_details", {})
    html.append("<h2>Couverture des schémas</h2>")
    html.append(f"<p>Schémas couverts : {schema_analysis.get('covered_schemas',0)} / {schema_analysis.get('total_schemas',0)}</p>")
    html.append("<div>")
    for schema_name, schema in sorted(schema_details.items()):
        html.append(render_schema(schema_name, schema))
    html.append("</div>")

    html.append("</body></html>")
    output_path.write_text("\n".join(html), encoding="utf-8")
    print(f"Rapport HTML généré : {output_path}")

if __name__ == "__main__":
    main()