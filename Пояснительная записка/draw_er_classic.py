from __future__ import annotations

import zipfile
from pathlib import Path
from xml.etree import ElementTree as ET

from PIL import Image, ImageDraw, ImageFont


ROOT = Path(__file__).resolve().parent
PNG_PATH = ROOT / "Схемы" / "png" / "02_er_model.png"
DOCX_PATH = ROOT / "Афонькин М.А. ВКР черновик v8.docx"

W, H = 1800, 980
BG = "#ffffff"
INK = "#2b221b"
ACCENT = "#5a3f2b"
LINE = "#9b8267"
PANEL = "#fffaf2"


def font(name: str, size: int):
    path = Path("C:/Windows/Fonts") / name
    return ImageFont.truetype(str(path), size=size)


FONT_TITLE = font("arialbd.ttf", 42)
FONT_SUB = font("arial.ttf", 24)
FONT_ENTITY = font("arialbd.ttf", 25)
FONT_FIELD = font("arial.ttf", 19)
FONT_CARD = font("arialbd.ttf", 20)


def text_center(draw: ImageDraw.ImageDraw, box: tuple[int, int, int, int], text: str, fnt, fill=INK):
    x1, y1, x2, y2 = box
    bbox = draw.textbbox((0, 0), text, font=fnt)
    x = x1 + (x2 - x1 - (bbox[2] - bbox[0])) / 2
    y = y1 + (y2 - y1 - (bbox[3] - bbox[1])) / 2
    draw.text((x, y), text, font=fnt, fill=fill)


def entity(draw: ImageDraw.ImageDraw, x: int, y: int, w: int, h: int, name: str, fields: list[str]):
    draw.rounded_rectangle((x, y, x + w, y + h), radius=14, fill=PANEL, outline=LINE, width=2)
    draw.rounded_rectangle((x, y, x + w, y + 50), radius=14, fill=ACCENT, outline=ACCENT, width=0)
    draw.rectangle((x, y + 30, x + w, y + 50), fill=ACCENT)
    draw.text((x + 18, y + 13), name, font=FONT_ENTITY, fill="white")
    yy = y + 68
    for item in fields:
        draw.text((x + 18, yy), item, font=FONT_FIELD, fill=INK)
        yy += 30


def line(draw: ImageDraw.ImageDraw, p1: tuple[int, int], p2: tuple[int, int], left: str, right: str):
    draw.line((p1, p2), fill=ACCENT, width=3)
    for p in (p1, p2):
        draw.ellipse((p[0] - 5, p[1] - 5, p[0] + 5, p[1] + 5), fill=ACCENT)
    mx = (p1[0] + p2[0]) // 2
    my = (p1[1] + p2[1]) // 2
    draw.text((p1[0] + 10, p1[1] - 28), left, font=FONT_CARD, fill=ACCENT)
    draw.text((p2[0] - 32, p2[1] - 28), right, font=FONT_CARD, fill=ACCENT)
    draw.text((mx - 8, my - 34), "", font=FONT_CARD, fill=ACCENT)


def polyline(draw: ImageDraw.ImageDraw, points: list[tuple[int, int]], left: str, right: str):
    draw.line(points, fill=ACCENT, width=3, joint="curve")
    start, end = points[0], points[-1]
    for p in (start, end):
        draw.ellipse((p[0] - 5, p[1] - 5, p[0] + 5, p[1] + 5), fill=ACCENT)
    draw.text((start[0] + 10, start[1] - 28), left, font=FONT_CARD, fill=ACCENT)
    draw.text((end[0] - 32, end[1] - 28), right, font=FONT_CARD, fill=ACCENT)


def make_png() -> None:
    image = Image.new("RGB", (W, H), BG)
    draw = ImageDraw.Draw(image)
    draw.rounded_rectangle((36, 36, W - 36, H - 36), radius=30, fill=BG, outline="#cdbda8", width=2)
    draw.text((76, 76), "Логическая модель базы данных", font=FONT_TITLE, fill=INK)
    draw.text((76, 132), "Классическое представление сущностей и связей для раздела проектирования базы данных.", font=FONT_SUB, fill="#7a604a")
    draw.rounded_rectangle((1540, 76, 1698, 126), radius=25, outline="#cdbda8", fill="#fffaf2", width=2)
    text_center(draw, (1540, 76, 1698, 126), "PostgreSQL", FONT_CARD, fill=ACCENT)

    boxes = {
        "users": (90, 230, 245, 185, ["PK id", "username", "email", "password", "profile fields"]),
        "genres": (420, 230, 245, 155, ["PK id", "name", "description"]),
        "albums": (750, 230, 245, 185, ["PK id", "FK genre_id", "title", "artist", "average_rating"]),
        "tracks": (1080, 230, 245, 185, ["PK id", "FK album_id", "title", "duration", "track_number"]),
        "user_follows": (90, 570, 245, 155, ["PK id", "FK follower_id", "FK following_id", "created_at"]),
        "reviews": (420, 570, 245, 185, ["PK id", "FK user_id", "FK album_id", "FK track_id", "scores, status"]),
        "track_genres": (750, 570, 245, 155, ["PK id", "FK track_id", "FK genre_id"]),
        "review_likes": (1080, 570, 245, 155, ["PK id", "FK user_id", "FK review_id", "created_at"]),
        "album_likes": (1410, 230, 245, 155, ["PK id", "FK user_id", "FK album_id", "created_at"]),
        "track_likes": (1410, 570, 245, 155, ["PK id", "FK user_id", "FK track_id", "created_at"]),
    }

    # Draw links first. The scheme intentionally keeps a classical 1:N notation
    # and shows the main foreign-key dependencies without implementation-only details.
    line(draw, (665, 322), (750, 322), "1", "N")      # genres -> albums
    line(draw, (995, 322), (1080, 322), "1", "N")     # albums -> tracks
    line(draw, (213, 415), (213, 570), "1", "N")      # users -> user_follows
    line(draw, (335, 660), (420, 660), "1", "N")      # users -> reviews
    polyline(draw, [(665, 610), (700, 500), (1080, 500), (1080, 610)], "1", "N")  # reviews -> review_likes
    polyline(draw, [(542, 385), (542, 500), (750, 610)], "1", "N")  # genres -> track_genres
    line(draw, (1080, 415), (995, 610), "1", "N")      # tracks -> track_genres
    line(draw, (1325, 322), (1410, 322), "1", "N")    # albums -> album_likes
    line(draw, (1325, 660), (1410, 660), "1", "N")    # tracks -> track_likes
    line(draw, (335, 310), (420, 590), "1", "N")      # users -> reviews
    line(draw, (995, 310), (665, 590), "1", "N")      # albums -> reviews
    line(draw, (1080, 350), (665, 630), "1", "N")     # tracks -> reviews

    for name, (x, y, w, h, fields) in boxes.items():
        entity(draw, x, y, w, h, name, fields)

    PNG_PATH.parent.mkdir(parents=True, exist_ok=True)
    image.save(PNG_PATH)


def replace_figure_12_image() -> None:
    ns = {
        "w": "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
        "a": "http://schemas.openxmlformats.org/drawingml/2006/main",
        "r": "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
        "rel": "http://schemas.openxmlformats.org/package/2006/relationships",
    }
    with zipfile.ZipFile(DOCX_PATH, "a") as zf:
        document_xml = zf.read("word/document.xml")
        root = ET.fromstring(document_xml)
        paragraphs = list(root.findall(".//w:body/w:p", ns))
        target_rid = None
        for index, paragraph in enumerate(paragraphs):
            text = "".join(node.text or "" for node in paragraph.findall(".//w:t", ns))
            if text.strip().startswith("Рисунок 12 -"):
                for prev in reversed(paragraphs[max(0, index - 5):index]):
                    blip = prev.find(".//a:blip", ns)
                    if blip is not None:
                        target_rid = blip.attrib.get(f"{{{ns['r']}}}embed")
                        break
                break
        if not target_rid:
            raise RuntimeError("Не найдено изображение перед подписью рисунка 12")

        rels = ET.fromstring(zf.read("word/_rels/document.xml.rels"))
        target = None
        for rel in rels.findall("rel:Relationship", ns):
            if rel.attrib.get("Id") == target_rid:
                target = rel.attrib["Target"]
                break
        if not target:
            raise RuntimeError("Не найдена связь изображения рисунка 12")

        media_path = "word/" + target
        zf.writestr(media_path, PNG_PATH.read_bytes())


if __name__ == "__main__":
    make_png()
    replace_figure_12_image()
    print(PNG_PATH)
