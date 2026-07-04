#!/usr/bin/env python3
"""Spektral-Generator für die jotti-Logo-Master.

Liest die 12 grünen Master aus assets/ und erzeugt den kompletten Satz mit
kontinuierlichem Spektral-Verlauf (Rot → Violett) in einen Staging-Ordner.
Die Master werden nie direkt überschrieben; der Austausch erfolgt erst nach
visueller Freigabe von Hand.

Farbmodell:
- Pixel werden im OKLCH-Raum klassifiziert. Markenpixel sind alle Pixel mit
  Chroma >= 0.02 und Hue im Grün-Fenster [100°, 230°]; das erfasst das J
  inklusive Antialiasing-Säumen. Hintergründe (Slate, Weiß) und der
  Schriftzug liegen außerhalb (chroma-arm oder Blau ~260°) und bleiben
  byte-identisch erhalten.
- Der Ziel-Hue wandert linear entlang der Verlaufsachse über der Bounding-Box
  der Markenpixel (Standard: oben Rot 25° → unten Violett 305°, Grün 165°
  liegt in der Mitte). Helligkeit und Chroma stammen aus dem Master; Chroma
  wird bei Bedarf ans sRGB-Gamut geklemmt. Alpha bleibt unverändert.

Eingebaute Checks (brechen mit Fehler ab):
- alle 12 Varianten erzeugt
- Alpha-Kanal byte-identisch
- Nicht-Markenpixel (Neutraltöne, Hintergründe, Text) byte-identisch
- Hue-Spannweite der Markenpixel deckt das Spektrum ab (Rot bis Violett)
- Dateigrößen in der Größenordnung der bisherigen Assets (Faktor 0.25–4)

Zusätzlich entstehen Abnahme-Previews in <out>/preview/: Graustufen-Proben
aller Varianten (Druck-Check) und 16-fach vergrößerte 16px-Favicons.

Verwendung:
    python3 scripts/generate-spektral-logos.py
    python3 scripts/generate-spektral-logos.py --axis-angle 45   # Diagonale
"""

import argparse
import math
import sys
from pathlib import Path

from PIL import Image

MASTER_NAMES = [
    "jotti-icon-dark-16.png",
    "jotti-icon-dark-32.png",
    "jotti-icon-dark-64.png",
    "jotti-icon-light-16.png",
    "jotti-icon-light-32.png",
    "jotti-icon-light-64.png",
    "jotti-logo-full-dark.png",
    "jotti-logo-full-light-transparent.png",
    "jotti-logo-full-light.png",
    "jotti-logo-icon-dark.png",
    "jotti-logo-icon-light.png",
    "jotti-symbol.png",
]

CHROMA_MIN = 0.02
BRAND_HUE_WINDOW = (100.0, 230.0)
SIZE_FACTOR_RANGE = (0.25, 4.0)
# Toleranzen für den Spektrum-Check: 8-Bit-Quantisierung verschiebt Hue bei
# niedrigem Chroma um einige Grad, deshalb nicht exakt auf 25°/305° prüfen.
CHECK_HUE_RED_MAX = 60.0
CHECK_HUE_VIOLET_MIN = 270.0


def srgb_to_linear(c: float) -> float:
    return c / 12.92 if c <= 0.04045 else ((c + 0.055) / 1.055) ** 2.4


def linear_to_srgb(c: float) -> float:
    return c * 12.92 if c <= 0.0031308 else 1.055 * c ** (1 / 2.4) - 0.055


def rgb8_to_oklch(r8: int, g8: int, b8: int) -> tuple[float, float, float]:
    r = srgb_to_linear(r8 / 255)
    g = srgb_to_linear(g8 / 255)
    b = srgb_to_linear(b8 / 255)
    l = 0.4122214708 * r + 0.5363325363 * g + 0.0514459929 * b
    m = 0.2119034982 * r + 0.6806995451 * g + 0.1073969566 * b
    s = 0.0883024619 * r + 0.2817188376 * g + 0.6299787005 * b
    l, m, s = l ** (1 / 3), m ** (1 / 3), s ** (1 / 3)
    L = 0.2104542553 * l + 0.7936177850 * m - 0.0040720468 * s
    a = 1.9779984951 * l - 2.4285922050 * m + 0.4505937099 * s
    b_ = 0.0259040371 * l + 0.7827717662 * m - 0.8086757660 * s
    return L, math.hypot(a, b_), math.degrees(math.atan2(b_, a)) % 360


def oklch_to_linear_rgb(L: float, C: float, H: float) -> tuple[float, float, float]:
    h = math.radians(H)
    a, b_ = C * math.cos(h), C * math.sin(h)
    l = L + 0.3963377774 * a + 0.2158037573 * b_
    m = L - 0.1055613458 * a - 0.0638541728 * b_
    s = L - 0.0894841775 * a - 1.2914855480 * b_
    l, m, s = l ** 3, m ** 3, s ** 3
    r = 4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s
    g = -1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s
    b = -0.0041960863 * l - 0.7034186147 * m + 1.7076147010 * s
    return r, g, b


def oklch_to_rgb8(L: float, C: float, H: float) -> tuple[int, int, int]:
    """OKLCH → 8-Bit-sRGB; Chroma wird ans Gamut geklemmt (Hue/L bleiben)."""
    r, g, b = oklch_to_linear_rgb(L, C, H)
    if not all(-1e-6 <= v <= 1 + 1e-6 for v in (r, g, b)):
        lo, hi = 0.0, C
        for _ in range(24):
            mid = (lo + hi) / 2
            r, g, b = oklch_to_linear_rgb(L, mid, H)
            if all(-1e-6 <= v <= 1 + 1e-6 for v in (r, g, b)):
                lo = mid
            else:
                hi = mid
        r, g, b = oklch_to_linear_rgb(L, lo, H)
    def to8(v: float) -> int:
        return min(255, max(0, round(linear_to_srgb(min(1.0, max(0.0, v))) * 255)))
    return to8(r), to8(g), to8(b)


def is_brand(L: float, C: float, H: float) -> bool:
    return C >= CHROMA_MIN and BRAND_HUE_WINDOW[0] <= H <= BRAND_HUE_WINDOW[1]


def spektralisiere(
    img: Image.Image, axis_angle: float, hue_start: float, hue_end: float
) -> tuple[Image.Image, list[bool]]:
    """Erzeugt die Spektral-Variante; gibt Bild und Markenpixel-Maske zurück."""
    rgba = img.convert("RGBA")
    w, h = rgba.size
    src = list(rgba.getdata())
    oklch_cache: dict[tuple[int, int, int], tuple[float, float, float]] = {}

    def oklch(r: int, g: int, b: int) -> tuple[float, float, float]:
        key = (r, g, b)
        if key not in oklch_cache:
            oklch_cache[key] = rgb8_to_oklch(r, g, b)
        return oklch_cache[key]

    # Durchgang 1: Markenpixel finden und Projektion auf die Achse bestimmen
    dx, dy = math.cos(math.radians(axis_angle)), math.sin(math.radians(axis_angle))
    mask = [False] * len(src)
    proj_min, proj_max = math.inf, -math.inf
    for i, (r, g, b, a) in enumerate(src):
        if a == 0:
            continue
        if is_brand(*oklch(r, g, b)):
            mask[i] = True
            p = (i % w) * dx + (i // w) * dy
            proj_min, proj_max = min(proj_min, p), max(proj_max, p)
    if proj_min >= proj_max:
        raise SystemExit("FEHLER: keine Markenpixel gefunden")

    # Durchgang 2: Hue positionsabhängig neu zuordnen
    out = list(src)
    span = proj_max - proj_min
    for i, (r, g, b, a) in enumerate(src):
        if not mask[i]:
            continue
        L, C, _ = oklch(r, g, b)
        t = ((i % w) * dx + (i // w) * dy - proj_min) / span
        nr, ng, nb = oklch_to_rgb8(L, C, hue_start + t * (hue_end - hue_start))
        out[i] = (nr, ng, nb, a)

    result = Image.new("RGBA", (w, h))
    result.putdata(out)
    return result.convert(img.mode), mask


def check_variante(
    name: str, src: Image.Image, out_path: Path, mask: list[bool], src_size: int
) -> list[str]:
    """Prüft die geschriebene Datei gegen den Master; gibt Fehlerliste zurück."""
    errors = []
    out = Image.open(out_path)
    if out.size != src.size or out.mode != src.mode:
        return [f"{name}: Größe/Modus verändert ({src.mode}{src.size} → {out.mode}{out.size})"]
    src_px = list(src.convert("RGBA").getdata())
    out_px = list(out.convert("RGBA").getdata())

    hue_min, hue_max = math.inf, -math.inf
    for i, ((sr, sg, sb, sa), (orr, og, ob, oa)) in enumerate(zip(src_px, out_px)):
        if oa != sa:
            errors.append(f"{name}: Alpha verändert an Pixel {i}")
            break
        if not mask[i]:
            if (orr, og, ob) != (sr, sg, sb) and sa != 0:
                errors.append(f"{name}: Nicht-Markenpixel {i} verändert")
                break
        else:
            L, C, H = rgb8_to_oklch(orr, og, ob)
            if C >= 0.04:
                hue_min, hue_max = min(hue_min, H), max(hue_max, H)
    if hue_min > CHECK_HUE_RED_MAX or hue_max < CHECK_HUE_VIOLET_MIN:
        errors.append(
            f"{name}: Hue-Spannweite [{hue_min:.0f}°, {hue_max:.0f}°] deckt das "
            f"Spektrum nicht ab (erwartet <= {CHECK_HUE_RED_MAX:.0f}° bis >= {CHECK_HUE_VIOLET_MIN:.0f}°)"
        )
    factor = out_path.stat().st_size / src_size
    if not SIZE_FACTOR_RANGE[0] <= factor <= SIZE_FACTOR_RANGE[1]:
        errors.append(f"{name}: Dateigröße um Faktor {factor:.2f} verändert")
    return errors


def schreibe_previews(out_dir: Path) -> None:
    preview_dir = out_dir / "preview"
    preview_dir.mkdir(exist_ok=True)
    for name in MASTER_NAMES:
        img = Image.open(out_dir / name)
        gray = img.convert("LA") if img.mode == "RGBA" else img.convert("L")
        gray.save(preview_dir / f"grau-{name}")
        if img.size == (16, 16):
            img.resize((256, 256), Image.NEAREST).save(preview_dir / f"zoom16-{name}")


def main() -> None:
    repo_root = Path(__file__).resolve().parent.parent
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--assets-dir", type=Path, default=repo_root / "assets")
    parser.add_argument("--out", type=Path, default=repo_root / "spektral-staging")
    parser.add_argument(
        "--axis-angle", type=float, default=90.0,
        help="Richtung der Verlaufsachse in Grad: 90 = oben nach unten (Standard), "
        "0 = links nach rechts, 45 = Diagonale von links-oben nach rechts-unten",
    )
    parser.add_argument("--hue-start", type=float, default=25.0, help="OKLCH-Hue am Achsenanfang (Rot)")
    parser.add_argument("--hue-end", type=float, default=305.0, help="OKLCH-Hue am Achsenende (Violett)")
    args = parser.parse_args()

    missing = [n for n in MASTER_NAMES if not (args.assets_dir / n).is_file()]
    if missing:
        raise SystemExit(f"FEHLER: Master fehlen in {args.assets_dir}: {', '.join(missing)}")

    args.out.mkdir(parents=True, exist_ok=True)
    errors: list[str] = []
    for name in MASTER_NAMES:
        src_path = args.assets_dir / name
        src = Image.open(src_path)
        result, mask = spektralisiere(src, args.axis_angle, args.hue_start, args.hue_end)
        out_path = args.out / name
        result.save(out_path, optimize=True)
        errors += check_variante(name, src, out_path, mask, src_path.stat().st_size)
        print(f"  {name}: {sum(mask)} Markenpixel umgefärbt")

    produced = [n for n in MASTER_NAMES if (args.out / n).is_file()]
    if len(produced) != len(MASTER_NAMES):
        errors.append(f"nur {len(produced)}/{len(MASTER_NAMES)} Varianten erzeugt")
    if errors:
        raise SystemExit("CHECKS FEHLGESCHLAGEN:\n  " + "\n  ".join(errors))

    schreibe_previews(args.out)
    print(f"\nOK: {len(produced)} Varianten in {args.out}, alle Checks bestanden.")
    print(f"Previews (Graustufen, 16px-Zoom) in {args.out / 'preview'}")


if __name__ == "__main__":
    main()
