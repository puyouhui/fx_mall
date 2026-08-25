#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
本地批量上传 products/ 目录商品（基础字段）。

前置：
1. go_backend 已启动（图片唯一名修复后需重启）
2. 已执行 scripts/seed_local_categories/seed_local_categories.py
3. 本地 MinIO 可用

用法：
  python3 upload_local_products.py
  python3 upload_local_products.py --limit 5
  python3 upload_local_products.py --dry-run
  # 修复已导入商品的重复图片（只用本地 img_* 重新上传）
  python3 upload_local_products.py --fix-images
"""

from __future__ import annotations

import argparse
import hashlib
import json
import mimetypes
import sys
import time
import uuid
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parents[1]
DEFAULT_PRODUCTS_DIR = REPO_ROOT / "products"
DEFAULT_MAP_FILE = (
    REPO_ROOT / "scripts" / "seed_local_categories" / "tag_category_map.json"
)
DEFAULT_STATE_FILE = SCRIPT_DIR / "upload_state.json"
DEFAULT_BASE_URL = "http://127.0.0.1:8082/api/mini"
DEFAULT_USERNAME = "admin"
DEFAULT_PASSWORD = "admin123"

IMAGE_EXTS = {".jpg", ".jpeg", ".png", ".gif", ".webp"}


class ApiClient:
    def __init__(self, base_url: str, token: Optional[str] = None):
        self.base_url = base_url.rstrip("/")
        self.token = token

    def _url(self, path: str) -> str:
        return f"{self.base_url}{path}"

    def _json_request(
        self,
        method: str,
        path: str,
        body: Optional[dict] = None,
        auth: bool = True,
    ) -> Tuple[bool, Any]:
        data = None
        headers = {"Content-Type": "application/json", "Accept": "application/json"}
        if auth and self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        if body is not None:
            data = json.dumps(body, ensure_ascii=False).encode("utf-8")

        req = urllib.request.Request(
            self._url(path), data=data, headers=headers, method=method
        )
        try:
            with urllib.request.urlopen(req, timeout=60) as resp:
                payload = json.loads(resp.read().decode("utf-8"))
        except urllib.error.HTTPError as exc:
            raw = exc.read().decode("utf-8", errors="replace")
            try:
                payload = json.loads(raw)
            except Exception:
                return False, raw or str(exc)
        except Exception as exc:
            return False, str(exc)

        if isinstance(payload, dict) and payload.get("code") == 200:
            return True, payload.get("data")
        message = ""
        if isinstance(payload, dict):
            message = payload.get("message") or ""
        return False, message or payload

    def login(self, username: str, password: str) -> Tuple[bool, str]:
        ok, data = self._json_request(
            "POST",
            "/admin/login",
            {"username": username, "password": password},
            auth=False,
        )
        if not ok:
            return False, str(data)
        token = data.get("token") if isinstance(data, dict) else ""
        if not token:
            return False, "登录成功但未返回 token"
        self.token = token
        return True, token

    def get_default_uom_category_id(self) -> Tuple[bool, Any]:
        return self._json_request("GET", "/admin/uom/default-category")

    def get_product(self, product_id: int) -> Tuple[bool, Any]:
        return self._json_request("GET", f"/admin/products/{product_id}")

    def create_product(self, payload: dict) -> Tuple[bool, Any]:
        return self._json_request("POST", "/admin/products", payload)

    def update_product(self, product_id: int, payload: dict) -> Tuple[bool, Any]:
        return self._json_request("PUT", f"/admin/products/{product_id}", payload)

    def upload_image(self, file_path: Path) -> Tuple[bool, Any]:
        boundary = f"----CursorFormBoundary{uuid.uuid4().hex}"
        filename = file_path.name
        content_type = mimetypes.guess_type(filename)[0] or "application/octet-stream"
        file_bytes = file_path.read_bytes()

        body = b"".join(
            [
                f"--{boundary}\r\n".encode(),
                (
                    f'Content-Disposition: form-data; name="file"; filename="{filename}"\r\n'
                ).encode(),
                f"Content-Type: {content_type}\r\n\r\n".encode(),
                file_bytes,
                b"\r\n",
                f"--{boundary}--\r\n".encode(),
            ]
        )
        headers = {
            "Content-Type": f"multipart/form-data; boundary={boundary}",
            "Accept": "application/json",
        }
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"

        req = urllib.request.Request(
            self._url("/admin/products/upload"),
            data=body,
            headers=headers,
            method="POST",
        )
        try:
            with urllib.request.urlopen(req, timeout=120) as resp:
                payload = json.loads(resp.read().decode("utf-8"))
        except urllib.error.HTTPError as exc:
            raw = exc.read().decode("utf-8", errors="replace")
            try:
                payload = json.loads(raw)
            except Exception:
                return False, raw or str(exc)
        except Exception as exc:
            return False, str(exc)

        if isinstance(payload, dict) and payload.get("code") == 200:
            data = payload.get("data") or {}
            return True, data.get("imageUrl") or data.get("url") or data
        message = ""
        if isinstance(payload, dict):
            message = payload.get("message") or ""
        return False, message or payload


def load_json(path: Path) -> Any:
    return json.loads(path.read_text(encoding="utf-8"))


def load_tag_map(path: Path) -> Dict[int, int]:
    if not path.exists():
        raise FileNotFoundError(
            f"缺少映射文件: {path}\n请先运行 scripts/seed_local_categories/seed_local_categories.py"
        )
    rows = load_json(path)
    mapping: Dict[int, int] = {}
    for row in rows:
        mapping[int(row["tagId"])] = int(row["categoryId"])
    return mapping


def load_state(path: Path) -> Dict[str, Any]:
    if not path.exists():
        return {"imported": {}}
    try:
        data = load_json(path)
        if not isinstance(data, dict):
            return {"imported": {}}
        data.setdefault("imported", {})
        return data
    except Exception:
        return {"imported": {}}


def save_state(path: Path, state: Dict[str, Any]) -> None:
    path.write_text(json.dumps(state, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def list_product_dirs(products_dir: Path) -> List[Path]:
    dirs = []
    for child in sorted(products_dir.iterdir()):
        if child.is_dir() and (child / "info.json").exists():
            dirs.append(child)
    return dirs


def local_images(product_dir: Path) -> List[Path]:
    """只取本地 img_* 真实图片，并按内容去重。"""
    files: List[Path] = []
    seen_hash = set()
    candidates = [
        path
        for path in product_dir.iterdir()
        if path.is_file()
        and path.suffix.lower() in IMAGE_EXTS
        and path.name.startswith("img_")
    ]
    for path in sorted(candidates, key=lambda p: p.name):
        digest = hashlib.md5(path.read_bytes()).hexdigest()
        if digest in seen_hash:
            continue
        seen_hash.add(digest)
        files.append(path)
    return files


def upload_local_images(client: ApiClient, product_dir: Path) -> Tuple[bool, Any]:
    image_files = local_images(product_dir)
    if not image_files:
        return False, "目录内没有本地 img_* 图片"
    image_urls: List[str] = []
    for img in image_files:
        ok, url = client.upload_image(img)
        if not ok:
            return False, f"{img.name}: {url}"
        image_urls.append(str(url))
        time.sleep(0.02)
    return True, image_urls


def resolve_tag_id(info: dict) -> Optional[int]:
    if info.get("tagId"):
        return int(info["tagId"])
    tags = info.get("tags") or []
    if tags and tags[0].get("tagId"):
        return int(tags[0]["tagId"])
    return None


def build_specs(info: dict) -> List[dict]:
    price = float(info.get("price") or 0)
    skus = info.get("skus") or []
    if skus:
        specs = []
        for sku in skus:
            name = (sku.get("name") or sku.get("formatNum") or "默认规格").strip()
            specs.append(
                {
                    "name": name,
                    "wholesale_price": price,
                    "retail_price": price,
                    "cost": price,
                    "description": sku.get("formatNum") or "",
                    "delivery_count": 1.0,
                }
            )
        return specs
    return [
        {
            "name": "默认规格",
            "wholesale_price": price,
            "retail_price": price,
            "cost": price,
            "description": "",
            "delivery_count": 1.0,
        }
    ]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="本地批量上传 products 商品")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--username", default=DEFAULT_USERNAME)
    parser.add_argument("--password", default=DEFAULT_PASSWORD)
    parser.add_argument("--products-dir", default=str(DEFAULT_PRODUCTS_DIR))
    parser.add_argument("--map-file", default=str(DEFAULT_MAP_FILE))
    parser.add_argument("--state-file", default=str(DEFAULT_STATE_FILE))
    parser.add_argument("--limit", type=int, default=0, help="仅处理前 N 个，0 表示全部")
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument(
        "--force",
        action="store_true",
        help="忽略本地上传状态，强制重新创建（可能产生重复商品）",
    )
    parser.add_argument(
        "--fix-images",
        action="store_true",
        help="对已导入商品，用本地 img_* 重新上传并更新 images 字段",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    products_dir = Path(args.products_dir).resolve()
    map_file = Path(args.map_file).resolve()
    state_file = Path(args.state_file).resolve()

    if not products_dir.exists():
        print(f"商品目录不存在: {products_dir}", file=sys.stderr)
        return 1

    try:
        tag_map = load_tag_map(map_file)
    except Exception as exc:
        print(str(exc), file=sys.stderr)
        return 1

    product_dirs = list_product_dirs(products_dir)
    if args.limit > 0:
        product_dirs = product_dirs[: args.limit]

    print(f"API: {args.base_url}")
    print(f"商品目录: {products_dir}")
    print(f"待处理: {len(product_dirs)} 个")
    print(f"tag 映射: {len(tag_map)} 条")
    if args.fix_images:
        print("模式: 修复图片（本地 img_* 重新上传）")

    if args.dry_run:
        for d in product_dirs[:10]:
            info = load_json(d / "info.json")
            tag_id = resolve_tag_id(info)
            imgs = local_images(d)
            print(
                f"  {d.name} | {info.get('title')} | tagId={tag_id} -> categoryId={tag_map.get(tag_id)} | 本地图 {len(imgs)} 张"
            )
            for img in imgs:
                print(f"      - {img.name}")
        if len(product_dirs) > 10:
            print(f"  ... 其余 {len(product_dirs) - 10} 个省略")
        print("dry-run 结束")
        return 0

    client = ApiClient(args.base_url)
    ok, msg = client.login(args.username, args.password)
    if not ok:
        print(f"登录失败: {msg}", file=sys.stderr)
        return 1
    print("登录成功")

    ok, uom_data = client.get_default_uom_category_id()
    if not ok or not isinstance(uom_data, dict) or not uom_data.get("id"):
        print(f"获取默认计量单位类别失败: {uom_data}", file=sys.stderr)
        return 1
    uom_category_id = int(uom_data["id"])
    print(f"默认计量单位类别 id={uom_category_id}")

    state = load_state(state_file)
    imported: Dict[str, Any] = state["imported"]

    success = 0
    skipped = 0
    failed = 0

    for index, product_dir in enumerate(product_dirs, start=1):
        info_path = product_dir / "info.json"
        try:
            info = load_json(info_path)
        except Exception as exc:
            print(f"[{index}] 读取失败 {product_dir.name}: {exc}", file=sys.stderr)
            failed += 1
            continue

        goods_id = str(info.get("goods_id") or product_dir.name)
        title = (info.get("title") or info.get("goodsNum") or goods_id).strip()

        if args.fix_images:
            record = imported.get(goods_id)
            if not record or not record.get("productId"):
                print(f"[{index}] 跳过（未导入） {title}")
                skipped += 1
                continue
            product_id = int(record["productId"])
            ok, image_urls = upload_local_images(client, product_dir)
            if not ok:
                print(f"[{index}] 图片上传失败 {title}: {image_urls}", file=sys.stderr)
                failed += 1
                continue
            ok, existing = client.get_product(product_id)
            if not ok or not isinstance(existing, dict):
                print(f"[{index}] 获取商品失败 {title}: {existing}", file=sys.stderr)
                failed += 1
                continue
            payload = {
                "name": existing.get("name") or title,
                "description": existing.get("description") or "",
                "category_id": existing.get("category_id") or record.get("categoryId"),
                "uom_category_id": existing.get("uom_category_id") or uom_category_id,
                "supplier_id": existing.get("supplier_id"),
                "is_special": bool(existing.get("is_special")),
                "images": image_urls,
                "specs": existing.get("specs") or build_specs(info),
                "status": existing.get("status", 1),
            }
            ok, result = client.update_product(product_id, payload)
            if not ok:
                print(f"[{index}] 更新图片失败 {title}: {result}", file=sys.stderr)
                failed += 1
                continue
            success += 1
            print(f"[{index}] 已修复图片 {title} -> {len(image_urls)} 张")
            continue

        if not args.force and goods_id in imported:
            print(f"[{index}] 跳过已导入 {title} ({goods_id})")
            skipped += 1
            continue

        tag_id = resolve_tag_id(info)
        if not tag_id or tag_id not in tag_map:
            print(f"[{index}] 失败 {title}: 无法映射 tagId={tag_id}", file=sys.stderr)
            failed += 1
            continue
        category_id = tag_map[tag_id]

        ok, image_urls = upload_local_images(client, product_dir)
        if not ok:
            print(f"[{index}] 失败 {title}: {image_urls}", file=sys.stderr)
            failed += 1
            continue

        payload = {
            "name": title,
            "description": (info.get("subTitle") or info.get("goodsNum") or "").strip(),
            "category_id": category_id,
            "uom_category_id": uom_category_id,
            "is_special": False,
            "images": image_urls,
            "specs": build_specs(info),
            "status": 1,
        }
        ok, result = client.create_product(payload)
        if not ok:
            print(f"[{index}] 创建失败 {title}: {result}", file=sys.stderr)
            failed += 1
            continue

        product_id = result.get("id") if isinstance(result, dict) else None
        imported[goods_id] = {
            "productId": product_id,
            "title": title,
            "categoryId": category_id,
            "tagId": tag_id,
            "dir": product_dir.name,
        }
        save_state(state_file, state)
        success += 1
        print(
            f"[{index}] 成功 {title} -> productId={product_id}, categoryId={category_id}, images={len(image_urls)}"
        )

    print(f"完成：成功 {success}，跳过 {skipped}，失败 {failed}")
    print(f"状态文件: {state_file}")
    return 0 if failed == 0 else 2


if __name__ == "__main__":
    raise SystemExit(main())
