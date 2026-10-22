#!/usr/bin/env python3
"""Convierte el HTML de una pagina "Especificacion y Despliegue Micro" a JSON.

Uso:  python parse.py [pagina.html | -] > datos.json

- Funciona con paginas generadas por la plantilla y con paginas legacy
  importadas desde Word (estilos mso-*, comentarios <!--[endif]-->, spans
  anidados, numeracion "1.&nbsp;&nbsp;").
- Ubica secciones por el texto del encabezado y tablas por el texto de su
  cabecera, nunca por posicion.
- Nunca extrae credenciales: la tabla 5.2 se ignora por completo (S6).
- Solo biblioteca estandar de Python 3.10+.
"""

from __future__ import annotations

import json
import re
import sys
import unicodedata
from html.parser import HTMLParser
from pathlib import Path

VOID = {"br", "img", "hr", "col", "input", "meta", "link", "area", "base", "wbr", "source", "track", "embed", "param"}
SKIP = {"script", "style", "head", "title"}
HEADINGS = {"h1", "h2", "h3", "h4", "h5", "h6"}
BLOCKS = HEADINGS | {"p", "div", "ul", "ol", "li", "table", "tr", "td", "th", "tbody", "thead", "tfoot",
                     "blockquote", "pre", "section", "article", "main"}
CONTAINERS = {"div", "section", "article", "main", "body", "html", "blockquote", "#root"}
P_BOUNDARY = {"td", "th", "li", "table", "div", "body", "html", "blockquote", "#root"}


# --------------------------------------------------------------------------
# Arbol DOM minimo (tolerante al HTML de Word)
# --------------------------------------------------------------------------

class Node:
    __slots__ = ("tag", "attrs", "children")

    def __init__(self, tag: str, attrs: dict | None = None):
        self.tag = tag
        self.attrs = attrs or {}
        self.children: list = []


class _Builder(HTMLParser):
    def __init__(self) -> None:
        super().__init__(convert_charrefs=True)
        self.root = Node("#root")
        self.stack = [self.root]

    def _in_scope(self, tags: set, boundary: set) -> bool:
        for node in reversed(self.stack):
            if node.tag in tags:
                return True
            if node.tag in boundary:
                return False
        return False

    def _pop_to(self, tags: set) -> None:
        while len(self.stack) > 1:
            node = self.stack.pop()
            if node.tag in tags:
                return

    def handle_starttag(self, tag, attrs):
        if tag == "li" and self._in_scope({"li"}, {"ul", "ol", "table", "td", "th"}):
            self._pop_to({"li"})
        elif tag in ("td", "th"):
            if self._in_scope({"td", "th"}, {"tr", "table"}):
                self._pop_to({"td", "th"})
        elif tag == "tr":
            if self._in_scope({"td", "th"}, {"tr", "table"}):
                self._pop_to({"td", "th"})
            if self._in_scope({"tr"}, {"table"}):
                self._pop_to({"tr"})
        elif tag in ("tbody", "thead", "tfoot"):
            if self._in_scope({"tr", "td", "th"}, {"table"}):
                self._pop_to({"tr"})
            if self._in_scope({"tbody", "thead", "tfoot"}, {"table"}):
                self._pop_to({"tbody", "thead", "tfoot"})
        elif tag in BLOCKS and self._in_scope({"p"}, P_BOUNDARY):
            self._pop_to({"p"})
        node = Node(tag, {k: (v or "") for k, v in attrs})
        self.stack[-1].children.append(node)
        if tag not in VOID:
            self.stack.append(node)

    def handle_startendtag(self, tag, attrs):
        self.handle_starttag(tag, attrs)
        if tag not in VOID and len(self.stack) > 1 and self.stack[-1].tag == tag:
            self.stack.pop()

    def handle_endtag(self, tag):
        if tag in VOID:
            return
        for i in range(len(self.stack) - 1, 0, -1):
            if self.stack[i].tag == tag:
                del self.stack[i:]
                return

    def handle_data(self, data):
        self.stack[-1].children.append(data)


def build_tree(html_text: str) -> Node:
    builder = _Builder()
    builder.feed(html_text)
    builder.close()
    return builder.root


# --------------------------------------------------------------------------
# Texto
# --------------------------------------------------------------------------

def norm(text: str) -> str:
    return re.sub(r"\s+", " ", text.replace("​", "")).strip()


def fold(text: str) -> str:
    decomposed = unicodedata.normalize("NFD", norm(text))
    return "".join(c for c in decomposed if not unicodedata.combining(c)).lower()


class Line:
    __slots__ = ("kind", "text", "anchors")

    def __init__(self, kind: str, text: str, anchors: list):
        self.kind = kind
        self.text = text
        self.anchors = anchors


class _Lines:
    """Aplana un nodo en lineas de texto (parrafos, items, saltos <br>)."""

    def __init__(self) -> None:
        self.lines: list[Line] = []
        self.buf: list[str] = []
        self.anchors: list[tuple[str, str]] = []
        self.kind = "p"

    def flush(self) -> None:
        text = norm("".join(self.buf))
        if text:
            self.lines.append(Line(self.kind, text, list(self.anchors)))
        self.buf, self.anchors = [], []

    def walk(self, node: Node) -> None:
        for child in node.children:
            if isinstance(child, str):
                self.buf.append(child)
            elif child.tag in SKIP:
                continue
            elif child.tag == "br":
                self.flush()
            elif child.tag in BLOCKS:
                self.flush()
                saved = self.kind
                if child.tag == "li":
                    self.kind = "li"
                self.walk(child)
                self.flush()
                self.kind = saved
            elif child.tag == "a":
                start = len(self.buf)
                self.walk(child)
                self.anchors.append((child.attrs.get("href", ""), norm("".join(self.buf[start:]))))
            else:
                self.walk(child)


def block_lines(node: Node) -> list[Line]:
    collector = _Lines()
    collector.walk(node)
    collector.flush()
    return collector.lines


def cell_text(cell: Node) -> str:
    return "\n".join(line.text for line in block_lines(cell))


def anchor_value(href: str, text: str) -> str:
    return href if href not in ("", "#") else text


def link_value(cell: Node) -> str:
    lines = block_lines(cell)
    for line in lines:
        if line.anchors:
            href, text = line.anchors[0]
            return anchor_value(href, text)
    return "\n".join(line.text for line in lines)


# --------------------------------------------------------------------------
# Estructura del documento
# --------------------------------------------------------------------------

SECTION_PHRASES = (
    ("objetivo", "objetivo del requerimiento"),
    ("solucion", "descripcion general de la solucion"),
    ("usuario", "descripcion a nivel de usuario"),
    ("rnf", "catalogo de requerimientos no funcionales"),
    ("rf", "catalogo de requerimientos funcionales"),
    ("qa", "requiere pruebas de calidad"),
    ("docs", "documentacion relacionada"),
    ("creds", "credenciales"),
    ("dep", "secuencia de despliegue"),
    ("rev", "secuencia de reversion"),
)
NEUTRAL_PHRASES = (
    "catalogo de requerimientos",
    "despliegue",
    "pase a produccion",
    "fecha y hora de implementacion",
    "notificaciones a usuarios",
    "responsables durante",
)


def heading_key(text: str):
    """Clave de seccion, "" para un encabezado neutro, None si no es encabezado."""
    folded = fold(text)
    if not folded or len(folded) > 220:
        return None
    folded = re.sub(r"^[\d.\s]+", "", folded)
    for key, phrase in SECTION_PHRASES:
        if folded.startswith(phrase):
            return key
    for phrase in NEUTRAL_PHRASES:
        if folded.startswith(phrase):
            return ""
    return None


def top_blocks(node: Node):
    for child in node.children:
        if isinstance(child, str) or child.tag in SKIP:
            continue
        if child.tag in CONTAINERS:
            yield from top_blocks(child)
        elif child.tag in HEADINGS or child.tag in ("p", "ul", "ol", "table"):
            yield child


def table_rows(table: Node) -> list[list[Node]]:
    rows: list[list[Node]] = []

    def walk(node: Node) -> None:
        for child in node.children:
            if isinstance(child, str):
                continue
            if child.tag == "tr":
                cells = [c for c in child.children if not isinstance(c, str) and c.tag in ("td", "th")]
                rows.append(cells)
            elif child.tag in ("tbody", "thead", "tfoot"):
                walk(child)

    walk(table)
    return rows


def colspan(cell: Node) -> int:
    try:
        return int(cell.attrs.get("colspan", "1") or "1")
    except ValueError:
        return 1


def header_of(rows: list[list[Node]]) -> tuple[str, ...]:
    return tuple(fold(cell_text(c)) for c in rows[0]) if rows else ()


def table_kind(rows: list[list[Node]]) -> str:
    for row in rows[:2]:
        if len(row) >= 2 and fold(cell_text(row[0])).rstrip(": ") == "codigo":
            return "cover"
    head = header_of(rows)
    if head == ("peticion", "descripcion", "respuesta"):
        return "peticion"
    if head == ("descripcion", "link"):
        return "docs"
    if head == ("usuario",):
        return "creds"
    if len(head) >= 2 and head[0] == "#" and head[1].startswith("componente"):
        return "seq"
    if head and head[0].startswith("fecha y hora de inicio y fin de implementacion"):
        return "impl"
    if head == ("cargo", "nombre"):
        return "notif"
    if head == ("rol", "nombre", "fecha"):
        return "resp"
    if head == ("observaciones",):
        return "obs"
    return ""


def yes_no(text: str) -> str:
    folded = fold(text)
    si = re.search(r"\bsi\s*\(\s*x\s*\)", folded)
    no = re.search(r"\bno\s*\(\s*x\s*\)", folded)
    if si and not no:
        return "SI"
    if no and not si:
        return "NO"
    return ""


COVER_LABELS = {
    "codigo": "codigo",
    "version": "version",
    "fecha de la version": "fecha",
    "creado por": "creado_por",
    "nombre requerimiento": "nombre",
    "validado por": "validado_por",
}


def parse_cover(rows, out: dict) -> None:
    for row in rows:
        if len(row) < 2:
            continue
        label = fold(cell_text(row[0])).rstrip(": ")
        if label.startswith("link de redmine"):
            out["redmine_url"] = link_value(row[1])
        elif label in COVER_LABELS:
            out[COVER_LABELS[label]] = cell_text(row[1])


PETICION_LABELS = (
    ("aviso comercial", "aviso_comercial"),
    ("modulo afectado", "modulo_afectado"),
    ("funcionalidad afectada", "funcionalidad_afectada"),
    ("componente", "componente_afectado"),
)


def parse_peticion(rows, out: dict) -> None:
    for row in rows[1:]:
        if len(row) < 3:
            continue
        label = fold(cell_text(row[0]))
        for prefix, key in PETICION_LABELS:
            if label.startswith(prefix):
                respuesta = cell_text(row[2])
                if key == "aviso_comercial":
                    respuesta = yes_no(respuesta)
                out["peticion"][key] = {"descripcion": cell_text(row[1]), "respuesta": respuesta}
                break


def parse_docs(rows) -> list[dict]:
    items = []
    for row in rows[1:]:
        if len(row) < 2:
            continue
        item = {"descripcion": cell_text(row[0]), "link": link_value(row[1])}
        if item["descripcion"] or item["link"]:
            items.append(item)
    return items


def parse_description(cell: Node) -> tuple[list[str], str]:
    items: list[str] = []
    link = ""
    for line in block_lines(cell):
        if (not link and line.kind == "p" and len(line.anchors) == 1
                and norm(line.anchors[0][1]) == line.text):
            link = anchor_value(*line.anchors[0])
        else:
            items.append(line.text)
    return items, link


def parse_sequence(rows) -> tuple[list[dict], str]:
    groups: list[dict] = []
    current: dict | None = None
    total = ""
    for row in rows[1:]:
        if not row:
            continue
        if fold(cell_text(row[0])) == "tiempo total":
            total = cell_text(row[-1]) if len(row) > 1 else ""
            continue
        if len(row) == 1 or (len(row) < 5 and colspan(row[0]) >= 2):
            current = {"grupo": cell_text(row[0]), "pasos": []}
            groups.append(current)
            continue
        if len(row) < 5:
            continue
        descripcion, link = parse_description(row[2])
        step = {
            "componente": [line.text for line in block_lines(row[1])],
            "descripcion": descripcion,
            "link": link,
            "responsable": cell_text(row[3]),
            "tiempo": cell_text(row[4]),
        }
        if not (step["componente"] or descripcion or link or step["responsable"] or step["tiempo"]):
            continue
        if current is None:
            current = {"grupo": "", "pasos": []}
            groups.append(current)
        current["pasos"].append(step)
    return groups, total


def parse_simple_rows(rows, keys: tuple[str, ...]) -> list[dict]:
    items = []
    for row in rows[1:]:
        values = [cell_text(c) for c in row[: len(keys)]]
        values += [""] * (len(keys) - len(values))
        if any(values):
            items.append(dict(zip(keys, values)))
    return items


def parse_list(node: Node) -> list[str]:
    return [line.text for line in block_lines(node) if line.text]


def empty_result() -> dict:
    return {
        "codigo": "", "version": "", "fecha": "", "creado_por": "", "nombre": "", "validado_por": "",
        "redmine_url": "",
        "objetivo": [], "solucion": [], "descripcion_usuario": [],
        "peticion": {k: {"descripcion": "", "respuesta": ""} for k in (
            "aviso_comercial", "modulo_afectado", "funcionalidad_afectada", "componente_afectado")},
        "rf": [], "rnf": [],
        "requiere_qa": False,
        "documentacion": [],
        "despliegue": [], "reversion": [],
        "tiempo_total": "",
        "implementacion": [], "notificaciones": [], "responsables": [],
        "observaciones": "",
    }


LIST_TARGETS = {"objetivo": "objetivo", "solucion": "solucion", "usuario": "descripcion_usuario",
                "rf": "rf", "rnf": "rnf"}


def parse(html_text: str) -> dict:
    """Devuelve el dict del esquema JSON para el HTML de una pagina."""
    out = empty_result()
    section = ""
    qa_found = False
    seq_seen = 0
    first_title = ""

    for block in top_blocks(build_tree(html_text)):
        if block.tag in HEADINGS or block.tag == "p":
            text = " ".join(line.text for line in block_lines(block))
            key = heading_key(text)
            if key is not None:
                section = key
                continue
            if block.tag == "h3" and not first_title and text:
                first_title = text
            if section == "qa" and not qa_found and yes_no(text):
                out["requiere_qa"] = yes_no(text) == "SI"
                qa_found = True
        elif block.tag in ("ul", "ol"):
            target = LIST_TARGETS.get(section)
            if target:
                out[target].extend(parse_list(block))
        else:  # table
            rows = table_rows(block)
            kind = table_kind(rows)
            if kind == "cover":
                parse_cover(rows, out)
            elif kind == "peticion":
                parse_peticion(rows, out)
            elif kind == "docs":
                out["documentacion"] = parse_docs(rows)
            elif kind == "seq":
                groups, total = parse_sequence(rows)
                seq_seen += 1
                target = "reversion" if section == "rev" or (section != "dep" and seq_seen == 2) else "despliegue"
                out[target] = groups
                if total:
                    out["tiempo_total"] = total
            elif kind == "impl":
                out["implementacion"] = parse_simple_rows(rows, ("inicio_fin", "corte", "observaciones"))
            elif kind == "notif":
                out["notificaciones"] = parse_simple_rows(rows, ("cargo", "nombre"))
            elif kind == "resp":
                out["responsables"] = parse_simple_rows(rows, ("rol", "nombre", "fecha"))
            elif kind == "obs":
                out["observaciones"] = "\n".join(cell_text(c) for row in rows[1:] for c in row[:1] if cell_text(c))
            # kind == "creds": se ignora a proposito (S6)

    for key in ("rf", "rnf"):
        if [v.lower() for v in out[key]] == ["no aplica."]:
            out[key] = []
    if not out["nombre"]:
        out["nombre"] = first_title
    return out


# --------------------------------------------------------------------------
# CLI
# --------------------------------------------------------------------------

def main(argv: list[str]) -> int:
    if len(argv) > 1 or (argv and argv[0] in ("-h", "--help")):
        sys.stderr.buffer.write((__doc__.strip().splitlines()[2] + "\n").encode("utf-8"))
        return 0 if argv and argv[0] in ("-h", "--help") else 2
    try:
        raw = sys.stdin.buffer.read() if not argv or argv[0] == "-" else Path(argv[0]).read_bytes()
        data = parse(raw.decode("utf-8-sig"))
    except (OSError, UnicodeDecodeError) as exc:
        sys.stderr.buffer.write(f"Error: {exc}\n".encode("utf-8", "replace"))
        return 2
    sys.stdout.buffer.write((json.dumps(data, ensure_ascii=False, indent=2) + "\n").encode("utf-8"))
    sys.stdout.buffer.flush()
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
