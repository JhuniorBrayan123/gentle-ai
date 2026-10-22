"""Pruebas de render.py y parse.py (solo biblioteca estandar)."""

import copy
import json
import re
import subprocess
import sys
import unittest
from pathlib import Path

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))

import parse  # noqa: E402
import render  # noqa: E402

TESTDATA = HERE / "testdata"


def load_example():
    return json.loads((TESTDATA / "ejemplo.json").read_text(encoding="utf-8"))


def legacy_html():
    return (TESTDATA / "legacy_word.html").read_text(encoding="utf-8")


class RenderTests(unittest.TestCase):
    def test_render_keeps_fixed_style_markers(self):
        out = render.render(load_example())
        for marker in (
            "repositorio-bookstack.s3.amazonaws.com/uploads/images/gallery/2023-06/scaled-1680-/ej7njugrmMsBHWnf-image.png",
            "Book Antiqua",
            "Times New Roman",
            "rgb(79, 129, 189)",
            "rgb(141, 179, 226)",
            "Formato de Especificación y Despliegue Micro",
            "Versión 1.0",
        ):
            self.assertIn(marker, out)
        self.assertNotIn("{{", out)
        self.assertNotIn("<!--repeat", out)
        self.assertNotIn("<!--/repeat", out)

    def test_render_escapes_values(self):
        data = load_example()
        data["nombre"] = "<script>alert(1)</script>"
        data["objetivo"] = ["<b onclick=x>hola</b>"]
        out = render.render(data)
        self.assertNotIn("<script>", out)
        self.assertNotIn("<b onclick", out)
        self.assertIn("&lt;script&gt;alert(1)&lt;/script&gt;", out)

    def test_render_does_not_reexpand_placeholders_in_values(self):
        data = load_example()
        data["observaciones"] = "{{codigo}}"
        out = render.render(data)
        self.assertIn("{{codigo}}", out)

    def test_render_rejects_password_like_keys_at_any_depth(self):
        for key in ("password", "contrasena", "contraseña", "Clave"):
            data = load_example()
            data["peticion"]["modulo_afectado"][key] = "x"
            with self.assertRaises(render.CredentialFieldError, msg=key):
                render.render(data)
        data = load_example()
        data["despliegue"][0]["pasos"][0]["password"] = "x"
        with self.assertRaises(render.CredentialFieldError):
            render.render(data)

    def test_cli_rejects_password_keys_with_nonzero_exit(self):
        data = load_example()
        data["contraseña"] = "x"
        proc = subprocess.run(
            [sys.executable, str(HERE / "render.py")],
            input=json.dumps(data).encode("utf-8"),
            capture_output=True,
        )
        self.assertNotEqual(proc.returncode, 0)
        self.assertEqual(proc.stdout, b"")
        self.assertIn(b"contrase", proc.stderr)

    def test_cli_renders_file_argument_to_stdout(self):
        proc = subprocess.run(
            [sys.executable, str(HERE / "render.py"), str(TESTDATA / "ejemplo.json")],
            capture_output=True,
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertIn("EJEMPLO-0002".encode(), proc.stdout)

    def test_render_requires_codigo_and_nombre(self):
        data = load_example()
        del data["codigo"]
        with self.assertRaises(render.SchemaError):
            render.render(data)

    def test_empty_rf_and_rnf_render_no_aplica(self):
        data = load_example()
        data["rf"] = []
        data["rnf"] = []
        out = render.render(data)
        self.assertEqual(out.count("No aplica."), 2)
        parsed = parse.parse(out)
        self.assertEqual(parsed["rf"], [])
        self.assertEqual(parsed["rnf"], [])

    def test_empty_tables_render_one_empty_row(self):
        data = load_example()
        for key in ("documentacion", "despliegue", "reversion", "implementacion", "notificaciones", "responsables"):
            data[key] = []
        out = render.render(data)
        self.assertIn("rgb(79, 129, 189)", out)
        parsed = parse.parse(out)
        for key in ("documentacion", "despliegue", "reversion", "implementacion", "notificaciones", "responsables"):
            self.assertEqual(parsed[key], [], key)

    def test_requiere_qa_marks(self):
        data = load_example()
        data["requiere_qa"] = True
        out = render.render(data)
        self.assertRegex(out, r"SI \(X\)\s+/\s+NO \( \)\.")
        data["requiere_qa"] = False
        out = render.render(data)
        self.assertRegex(out, r"SI \( \)\s+/\s+NO \(X\)\.")
        self.assertFalse(parse.parse(out)["requiere_qa"])

    def test_aviso_comercial_marks(self):
        data = load_example()
        for respuesta, expected in (("SI", r"SI \(X\)\s/\s+NO \( \)"), ("NO", r"SI \( \)\s/\s+NO \(X\)")):
            data["peticion"]["aviso_comercial"]["respuesta"] = respuesta
            out = render.render(data)
            self.assertRegex(out, expected)
            self.assertEqual(parse.parse(out)["peticion"]["aviso_comercial"]["respuesta"], respuesta)

    def test_credentials_section_is_fixed_text(self):
        out = render.render(load_example())
        self.assertIn("Usuario:</strong> [usuario de prueba]", out)
        self.assertIn("se entrega por el canal seguro del equipo", out)


class ParseTests(unittest.TestCase):
    def test_roundtrip_render_then_parse_equals_input(self):
        data = load_example()
        self.assertEqual(parse.parse(render.render(data)), data)

    def test_roundtrip_is_stable_for_sparse_input(self):
        data = {"codigo": "X-1", "nombre": "Solo lo minimo"}
        parsed = parse.parse(render.render(data))
        self.assertEqual(parsed["codigo"], "X-1")
        self.assertEqual(parsed["nombre"], "Solo lo minimo")
        self.assertEqual(parsed["objetivo"], [])
        self.assertEqual(parse.parse(render.render(parsed)), parsed)

    def test_legacy_word_page_core_fields(self):
        data = parse.parse(legacy_html())
        self.assertEqual(data["codigo"], "EJEMPLO-0001")
        self.assertEqual(data["version"], "1.0")
        self.assertEqual(
            data["nombre"],
            'Agregar el filtro "Producto" en el listado global de Historias de la APP QA',
        )
        self.assertEqual(data["redmine_url"], "https://redmine.sreasons.com/issues/XXXXX")
        self.assertEqual(len(data["objetivo"]), 1)
        self.assertTrue(data["objetivo"][0].startswith("Permitir que el usuario QA filtre"))
        self.assertEqual(len(data["solucion"]), 2)
        self.assertEqual(len(data["descripcion_usuario"]), 2)
        self.assertTrue(data["requiere_qa"])
        self.assertEqual(data["tiempo_total"], "25 min")
        self.assertEqual(data["peticion"]["aviso_comercial"]["respuesta"], "NO")
        self.assertEqual(data["peticion"]["modulo_afectado"]["respuesta"], "Submódulos: Historias")
        self.assertEqual(len(data["rf"]), 1)
        self.assertTrue(data["rf"][0].startswith("RF-01"))
        self.assertEqual(len(data["rnf"]), 1)
        self.assertTrue(data["rnf"][0].startswith("RNF-01"))

    def test_legacy_word_page_deployment_sequences(self):
        data = parse.parse(legacy_html())
        groups = {g["grupo"]: g["pasos"] for g in data["despliegue"]}
        self.assertEqual([g["grupo"] for g in data["despliegue"]], ["BASE DE DATOS", "CONFIGSERVER"])
        db = groups["BASE DE DATOS"]
        self.assertEqual(len(db), 1)
        self.assertEqual(db[0]["componente"], ["Scripts", "DBQaApp"])
        self.assertEqual(db[0]["responsable"], "DBA")
        self.assertEqual(db[0]["tiempo"], "5 min")
        self.assertEqual(db[0]["link"], "[Link al adjunto del script en Redmine]")
        jenkins = groups["CONFIGSERVER"]
        self.assertEqual(len(jenkins), 1)
        self.assertEqual(jenkins[0]["componente"], ["JENKINS", "mf-qa-app"])
        self.assertEqual(len(jenkins[0]["descripcion"]), 5)
        self.assertTrue(jenkins[0]["descripcion"][0].startswith("Ingresar a la plataforma de Jenkins"))
        self.assertIn("feature-EJEMPLO-0001", jenkins[0]["descripcion"][2])
        self.assertEqual(
            [g["grupo"] for g in data["reversion"]],
            ["", "CONFIGSERVER", "ENTREGA CONTINUA"],
        )
        self.assertEqual(data["reversion"][0]["pasos"][0]["componente"], ["Scripts", "DBQaApp"])
        self.assertEqual(len(data["reversion"][2]["pasos"]), 2)

    def test_legacy_word_page_production_tables(self):
        data = parse.parse(legacy_html())
        self.assertEqual(data["documentacion"], [])
        self.assertEqual(data["implementacion"], [])
        self.assertEqual(
            data["notificaciones"],
            [{"cargo": "QA Lead", "nombre": "[Nombre]"}, {"cargo": "Desarrollador", "nombre": "[Nombre]"}],
        )
        self.assertEqual(
            data["responsables"],
            [
                {"rol": "QA Tester", "nombre": "[Nombre]", "fecha": "DD/MM/AAAA"},
                {"rol": "AgileOps", "nombre": "[Nombre]", "fecha": "DD/MM/AAAA"},
            ],
        )
        self.assertEqual(data["observaciones"], "")

    def test_parse_never_returns_credential_values(self):
        html = legacy_html()
        marker = "Hunter2-FICTICIA"
        injected = html.replace(
            "se entrega por el canal seguro del equipo]",
            marker + "] se entrega por el canal seguro del equipo]",
        )
        self.assertIn(marker, injected)
        dumped = json.dumps(parse.parse(injected), ensure_ascii=False)
        self.assertNotIn(marker, dumped)
        self.assertNotIn("canal seguro", dumped)
        self.assertNotIn("Contraseña", dumped)
        self.assertNotIn("usuario QA de prueba", dumped)
        self.assertEqual(set(parse.parse(injected)), set(load_example()))

    def test_sections_are_located_by_heading_text_not_position(self):
        html = render.render(load_example())
        m = re.search(r"<h4>.*?Objetivo del requerimiento</strong></h4>\s*<ul>.*?</ul>", html, re.S)
        self.assertIsNotNone(m)
        block = m.group(0)
        moved = html.replace(block, "") + block
        parsed = parse.parse(moved)
        self.assertEqual(parsed["objetivo"], load_example()["objetivo"])

    def test_cli_parse_reads_stdin_and_writes_json(self):
        proc = subprocess.run(
            [sys.executable, str(HERE / "parse.py")],
            input=legacy_html().encode("utf-8"),
            capture_output=True,
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        data = json.loads(proc.stdout.decode("utf-8"))
        self.assertEqual(data["codigo"], "EJEMPLO-0001")

    def test_example_json_roundtrips_without_mutation(self):
        data = load_example()
        before = copy.deepcopy(data)
        render.render(data)
        self.assertEqual(data, before)


if __name__ == "__main__":
    unittest.main()
