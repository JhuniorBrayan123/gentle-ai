#!/usr/bin/env python3
"""Renderiza el JSON del formato "Especificacion y Despliegue Micro" a HTML.

Uso:  python render.py [datos.json | -] > pagina.html

- Lee el JSON de un archivo o de stdin ("-" o sin argumento).
- Valida claves obligatorias (codigo, nombre) y tipos.
- Escapa TODO valor antes de insertarlo en la plantilla fija
  (../assets/plantilla.html); el HTML nunca se escribe a mano.
- Rechaza cualquier clave de tipo contrasena/password/clave (S6).
- Solo biblioteca estandar de Python 3.10+.
"""

from __future__ import annotations

import html
import json
import re
import sys
import unicodedata
from pathlib import Path

TEMPLATE_PATH = Path(__file__).resolve().parent.parent / "assets" / "plantilla.html"

CREDENTIAL_KEY = re.compile(r"password|passwd|pwd|contrasena|clave|secret", re.IGNORECASE)
URL_PREFIX = re.compile(r"(?:https?://|mailto:)", re.IGNORECASE)
MARKER = re.compile(r"<!--(/?)repeat(?::([A-Za-z_]+))?-->")
PLACEHOLDER = re.compile(r"\{\{\{([^{}]+?)\}\}\}|\{\{([^{}]+?)\}\}")

PETICION_KEYS = ("aviso_comercial", "modulo_afectado", "funcionalidad_afectada", "componente_afectado")
STRING_FIELDS = (
    "codigo", "version", "fecha", "creado_por", "nombre", "validado_por",
    "redmine_url", "tiempo_total", "observaciones",
)
STRING_LISTS = ("objetivo", "solucion", "descripcion_usuario", "rf", "rnf")
REQUIRED = ("codigo", "nombre")


class SchemaError(ValueError):
    """El JSON no cumple el esquema."""


class CredentialFieldError(SchemaError):
    """El JSON contiene una clave tipo contrasena (prohibido por S6)."""


# --------------------------------------------------------------------------
# Validacion y normalizacion
# --------------------------------------------------------------------------

def _fold(key: str) -> str:
    text = unicodedata.normalize("NFD", key)
    return "".join(c for c in text if not unicodedata.combining(c))


def find_credential_keys(obj, path: str = "") -> list[str]:
    found: list[str] = []
    if isinstance(obj, dict):
        for key, value in obj.items():
            here = f"{path}.{key}" if path else str(key)
            if CREDENTIAL_KEY.search(_fold(str(key))):
                found.append(here)
            found.extend(find_credential_keys(value, here))
    elif isinstance(obj, list):
        for i, value in enumerate(obj):
            found.extend(find_credential_keys(value, f"{path}[{i}]"))
    return found


def _text(value, path: str) -> str:
    if value is None:
        return ""
    if isinstance(value, bool) or not isinstance(value, (str, int, float)):
        raise SchemaError(f"{path}: se esperaba texto")
    return str(value).replace("\r\n", "\n").replace("\r", "\n").strip()


def _text_list(value, path: str) -> list[str]:
    if value is None:
        return []
    if isinstance(value, str):
        value = [value]
    if not isinstance(value, list):
        raise SchemaError(f"{path}: se esperaba una lista de textos")
    items = [_text(v, f"{path}[{i}]") for i, v in enumerate(value)]
    return [v for v in items if v]


def _dict(value, path: str) -> dict:
    if value is None:
        return {}
    if not isinstance(value, dict):
        raise SchemaError(f"{path}: se esperaba un objeto")
    return value


def _rows(value, path: str) -> list[dict]:
    if value is None:
        return []
    if not isinstance(value, list):
        raise SchemaError(f"{path}: se esperaba una lista")
    return [_dict(v, f"{path}[{i}]") for i, v in enumerate(value)]


def _steps(value, path: str) -> list[dict]:
    out = []
    for i, group in enumerate(_rows(value, path)):
        here = f"{path}[{i}]"
        steps = []
        for j, step in enumerate(_rows(group.get("pasos"), f"{here}.pasos")):
            sp = f"{here}.pasos[{j}]"
            steps.append({
                "componente": _text_list(step.get("componente"), f"{sp}.componente"),
                "descripcion": _text_list(step.get("descripcion"), f"{sp}.descripcion"),
                "link": _text(step.get("link"), f"{sp}.link"),
                "responsable": _text(step.get("responsable"), f"{sp}.responsable"),
                "tiempo": _text(step.get("tiempo"), f"{sp}.tiempo"),
            })
        out.append({"grupo": _text(group.get("grupo"), f"{here}.grupo"), "pasos": steps})
    return out


def normalize(data) -> dict:
    """Valida `data` y devuelve una copia completa con valores por defecto."""
    if not isinstance(data, dict):
        raise SchemaError("el JSON raiz debe ser un objeto")
    bad = find_credential_keys(data)
    if bad:
        raise CredentialFieldError(
            "Campo de credencial no permitido: " + ", ".join(repr(b) for b in bad)
            + ". La skill no maneja contraseñas; la sección 5.2 usa texto fijo."
        )
    for key in REQUIRED:
        if not isinstance(data.get(key), str) or not data[key].strip():
            raise SchemaError(f"Falta el campo obligatorio: {key}")

    out: dict = {}
    for key in STRING_FIELDS:
        out[key] = _text(data.get(key), key)
    if not out["version"]:
        out["version"] = "1.0"
    for key in STRING_LISTS:
        out[key] = _text_list(data.get(key), key)

    peticion_in = _dict(data.get("peticion"), "peticion")
    out["peticion"] = {}
    for key in PETICION_KEYS:
        row = _dict(peticion_in.get(key), f"peticion.{key}")
        out["peticion"][key] = {
            "descripcion": _text(row.get("descripcion"), f"peticion.{key}.descripcion"),
            "respuesta": _text(row.get("respuesta"), f"peticion.{key}.respuesta"),
        }
    resp = out["peticion"]["aviso_comercial"]["respuesta"].upper()
    out["peticion"]["aviso_comercial"]["respuesta"] = resp if resp in ("SI", "NO") else ""

    flag = data.get("requiere_qa", False)
    if not isinstance(flag, bool):
        raise SchemaError("requiere_qa: se esperaba true o false")
    out["requiere_qa"] = flag

    out["documentacion"] = [
        {"descripcion": _text(r.get("descripcion"), f"documentacion[{i}].descripcion"),
         "link": _text(r.get("link"), f"documentacion[{i}].link")}
        for i, r in enumerate(_rows(data.get("documentacion"), "documentacion"))
    ]
    out["despliegue"] = _steps(data.get("despliegue"), "despliegue")
    out["reversion"] = _steps(data.get("reversion"), "reversion")
    out["implementacion"] = [
        {k: _text(r.get(k), f"implementacion[{i}].{k}") for k in ("inicio_fin", "corte", "observaciones")}
        for i, r in enumerate(_rows(data.get("implementacion"), "implementacion"))
    ]
    out["notificaciones"] = [
        {k: _text(r.get(k), f"notificaciones[{i}].{k}") for k in ("cargo", "nombre")}
        for i, r in enumerate(_rows(data.get("notificaciones"), "notificaciones"))
    ]
    out["responsables"] = [
        {k: _text(r.get(k), f"responsables[{i}].{k}") for k in ("rol", "nombre", "fecha")}
        for i, r in enumerate(_rows(data.get("responsables"), "responsables"))
    ]
    return out


# --------------------------------------------------------------------------
# Fragmentos HTML (todo valor pasa por esc())
# --------------------------------------------------------------------------

def esc(value) -> str:
    text = html.escape(str(value), quote=True)
    return text.replace("\r\n", "\n").replace("\n", "<br>")


def link_html(link: str) -> str:
    if not link:
        return ""
    href = link if URL_PREFIX.match(link) else "#"
    return f'<a href="{html.escape(href, quote=True)}">{esc(link)}</a>'


def componente_html(lines: list[str]) -> str:
    if not lines:
        return '<p class="MsoNormal"><br></p>'
    return "\n".join(f'<p class="MsoNormal">{esc(line)}</p>' for line in lines)


def descripcion_html(items: list[str], link: str) -> str:
    parts = []
    if len(items) == 1:
        parts.append(f"<p>{esc(items[0])}</p>")
    elif items:
        parts.append("<ol>\n" + "\n".join(f"<li>{esc(i)}</li>" for i in items) + "\n</ol>")
    if link:
        parts.append(f"<p>{link_html(link)}</p>")
    return "\n".join(parts) if parts else "<p><br></p>"


def _sequence(groups: list[dict]) -> list[dict]:
    """Contexto de plantilla para una tabla de despliegue o reversion."""
    blank = {"n": "", "componente_html": componente_html([]), "descripcion_html": descripcion_html([], ""),
             "responsable": "", "tiempo": ""}
    if not groups:
        return [{"grupo_fila": [], "pasos": [blank]}]
    out, n = [], 0
    for group in groups:
        pasos = []
        for step in group["pasos"]:
            n += 1
            pasos.append({
                "n": str(n),
                "componente_html": componente_html(step["componente"]),
                "descripcion_html": descripcion_html(step["descripcion"], step["link"]),
                "responsable": step["responsable"],
                "tiempo": step["tiempo"],
            })
        out.append({"grupo_fila": [{"grupo": group["grupo"]}], "pasos": pasos or [blank]})
    return out


def _or_blank(rows: list[dict], keys: tuple[str, ...]) -> list[dict]:
    return rows or [{k: "" for k in keys}]


def build_context(data: dict) -> dict:
    ctx = dict(data)
    ctx["redmine_link"] = link_html(data["redmine_url"]) or "<br>"
    ctx["qa_si"], ctx["qa_no"] = ("X", " ") if data["requiere_qa"] else (" ", "X")
    aviso = data["peticion"]["aviso_comercial"]["respuesta"]
    ctx["aviso_si"] = "X" if aviso == "SI" else " "
    ctx["aviso_no"] = "X" if aviso == "NO" else " "
    for key in ("objetivo", "solucion", "descripcion_usuario"):
        ctx[key] = data[key] or [""]
    for key in ("rf", "rnf"):
        ctx[key] = data[key] or ["No aplica."]
    ctx["documentacion"] = [
        {"descripcion": r["descripcion"], "link_html": link_html(r["link"])}
        for r in _or_blank(data["documentacion"], ("descripcion", "link"))
    ]
    ctx["despliegue"] = _sequence(data["despliegue"])
    ctx["reversion"] = _sequence(data["reversion"])
    ctx["implementacion"] = _or_blank(data["implementacion"], ("inicio_fin", "corte", "observaciones"))
    ctx["notificaciones"] = _or_blank(data["notificaciones"], ("cargo", "nombre"))
    ctx["responsables"] = _or_blank(data["responsables"], ("rol", "nombre", "fecha"))
    return ctx


# --------------------------------------------------------------------------
# Motor de plantilla ({{x}} escapado, {{{x}}} crudo, bloques repeat)
# --------------------------------------------------------------------------

def parse_template(template: str) -> list:
    root: list = []
    stack = [("", root)]
    pos = 0
    for m in MARKER.finditer(template):
        stack[-1][1].append(template[pos:m.start()])
        pos = m.end()
        if m.group(1):
            if len(stack) == 1:
                raise SchemaError("plantilla: cierre de repeat sin apertura")
            stack.pop()
        else:
            node = (m.group(2) or "", [])
            stack[-1][1].append(node)
            stack.append(node)
    if len(stack) != 1:
        raise SchemaError("plantilla: repeat sin cerrar")
    stack[0][1].append(template[pos:])
    return root


def _lookup(path: str, scopes: list[dict]):
    path = path.strip()
    for scope in scopes:
        if path == ".":
            if "." in scope:
                return scope["."]
            continue
        cur = scope
        for part in path.split("."):
            if isinstance(cur, dict) and part in cur:
                cur = cur[part]
            else:
                break
        else:
            return cur
    raise SchemaError(f"plantilla: marcador desconocido {path!r}")


def _expand(nodes: list, scopes: list[dict]) -> str:
    out: list[str] = []
    for node in nodes:
        if isinstance(node, str):
            def repl(m, scopes=scopes):
                if m.group(1) is not None:
                    return str(_lookup(m.group(1), scopes))
                return esc(_lookup(m.group(2), scopes))
            out.append(PLACEHOLDER.sub(repl, node))
        else:
            name, children = node
            items = _lookup(name, scopes)
            for item in items:
                scope = item if isinstance(item, dict) else {".": item}
                out.append(_expand(children, [scope] + scopes))
    return "".join(out)


def render(data, template: str | None = None) -> str:
    """Devuelve el HTML de la pagina para `data` (dict con el esquema JSON)."""
    if template is None:
        template = TEMPLATE_PATH.read_text(encoding="utf-8")
    ctx = build_context(normalize(data))
    return _expand(parse_template(template), [ctx])


# --------------------------------------------------------------------------
# CLI
# --------------------------------------------------------------------------

def _eprint(message: str) -> None:
    sys.stderr.buffer.write((message + "\n").encode("utf-8", "replace"))
    sys.stderr.buffer.flush()


def main(argv: list[str]) -> int:
    if len(argv) > 1 or (argv and argv[0] in ("-h", "--help")):
        _eprint(__doc__.strip().splitlines()[2])
        return 0 if argv and argv[0] in ("-h", "--help") else 2
    try:
        if not argv or argv[0] == "-":
            raw = sys.stdin.buffer.read()
        else:
            raw = Path(argv[0]).read_bytes()
        data = json.loads(raw.decode("utf-8-sig"))
        page = render(data)
    except (OSError, UnicodeDecodeError, json.JSONDecodeError, SchemaError) as exc:
        _eprint(f"Error: {exc}")
        return 2
    sys.stdout.buffer.write(page.encode("utf-8"))
    sys.stdout.buffer.flush()
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
