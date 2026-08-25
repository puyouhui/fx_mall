#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
本地环境批量添加商品分类（8 个一级分类）。

打开任一一级分类时，小程序左侧「全部」会展示该分类下全部商品。

用法：
  python3 seed_local_categories.py
  python3 seed_local_categories.py --dry-run
"""

from __future__ import annotations

import argparse
import json
import sys
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

DEFAULT_BASE_URL = "http://127.0.0.1:8082/api/mini"
DEFAULT_USERNAME = "admin"
DEFAULT_PASSWORD = "admin123"
MAP_FILE = Path(__file__).resolve().parent / "tag_category_map.json"

# 8 个一级分类（tagId 用于商品导入映射）
TAGS: List[Dict[str, Any]] = [
    {"tagId": 81980741, "tagName": "餐饮抽纸"},
    {"tagId": 81916319, "tagName": "家庭抽纸"},
    {"tagId": 82169883, "tagName": "商务用纸"},
    {"tagId": 81979637, "tagName": "提装卷纸"},
    {"tagId": 84114839, "tagName": "散装卷纸"},
    {"tagId": 81979606, "tagName": "下拉抽纸"},
    {"tagId": 92801104, "tagName": "手帕纸巾"},
    {"tagId": 92801125, "tagName": "其他纸巾"},
]


class ApiClient:
    def __init__(self, base_url: str, token: Optional[str] = None):
        self.base_url = base_url.rstrip("/")
        self.token = token

    def _url(self, path: str) -> str:
        return f"{self.base_url}{path}"

    def _request(
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
            data = json.dumps(body).encode("utf-8")

        req = urllib.request.Request(
            self._url(path),
            data=data,
            headers=headers,
            method=method,
        )
        try:
            with urllib.request.urlopen(req, timeout=20) as resp:
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
        ok, data = self._request(
            "POST",
            "/admin/login",
            {"username": username, "password": password},
            auth=False,
        )
        if not ok:
            return False, str(data)
        token = ""
        if isinstance(data, dict):
            token = data.get("token") or ""
        if not token:
            return False, "登录成功但未返回 token"
        self.token = token
        return True, token

    def get_categories(self) -> Tuple[bool, Any]:
        return self._request("GET", "/admin/categories")

    def create_category(self, name: str, parent_id: int = 0) -> Tuple[bool, Any]:
        return self._request(
            "POST",
            "/admin/categories",
            {
                "name": name,
                "parent_id": parent_id,
                "status": 1,
                "icon": "",
            },
        )

    def delete_category(self, category_id: int) -> Tuple[bool, Any]:
        return self._request("DELETE", f"/admin/categories/{category_id}")

    def update_sort(self, items: List[Dict[str, int]]) -> Tuple[bool, Any]:
        return self._request("PUT", "/admin/categories/sort", {"items": items})


def flatten_categories(nodes: Any, out: Optional[List[Dict[str, Any]]] = None) -> List[Dict[str, Any]]:
    if out is None:
        out = []
    if not isinstance(nodes, list):
        return out
    for node in nodes:
        if not isinstance(node, dict):
            continue
        out.append(node)
        children = node.get("children") or []
        if children:
            flatten_categories(children, out)
    return out


def find_root_by_name(flat: List[Dict[str, Any]], name: str) -> Optional[Dict[str, Any]]:
    for item in flat:
        if item.get("name") == name and int(item.get("parent_id") or 0) == 0:
            return item
    return None


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="本地批量添加 8 个一级商品分类")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--username", default=DEFAULT_USERNAME)
    parser.add_argument("--password", default=DEFAULT_PASSWORD)
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--map-file", default=str(MAP_FILE))
    parser.add_argument(
        "--cleanup-groups",
        action="store_true",
        help="删除旧的分组一级（抽纸/卷纸/下拉抽纸/酒店商务用纸）及其子分类",
    )
    return parser.parse_args()


OLD_GROUP_NAMES = {"抽纸", "卷纸", "下拉抽纸", "酒店商务用纸"}


def main() -> int:
    args = parse_args()

    print(f"API: {args.base_url}")
    print(f"将处理 {len(TAGS)} 个一级分类")
    for index, tag in enumerate(TAGS, start=1):
        print(f"  {index}. {tag['tagName']} (tagId={tag['tagId']})")

    if args.dry_run:
        print("dry-run 模式，未写入数据库")
        return 0

    client = ApiClient(args.base_url)
    ok, msg = client.login(args.username, args.password)
    if not ok:
        print(f"登录失败: {msg}", file=sys.stderr)
        return 1
    print("登录成功")

    ok, categories = client.get_categories()
    if not ok:
        print(f"获取现有分类失败: {categories}", file=sys.stderr)
        return 1

    # 可选：清掉之前误建的分组结构
    if args.cleanup_groups:
        for group in list(categories or []):
            if group.get("name") in OLD_GROUP_NAMES and int(group.get("parent_id") or 0) == 0:
                for child in list(group.get("children") or []):
                    cid = child.get("id")
                    if cid:
                        ok, res = client.delete_category(int(cid))
                        print(f"[清理子分类] {child.get('name')} id={cid}: {res if not ok else 'ok'}")
                gid = group.get("id")
                if gid:
                    ok, res = client.delete_category(int(gid))
                    print(f"[清理一级分组] {group.get('name')} id={gid}: {res if not ok else 'ok'}")
        ok, categories = client.get_categories()
        if not ok:
            print(f"清理后获取分类失败: {categories}", file=sys.stderr)
            return 1

    flat = flatten_categories(categories)
    mapping: List[Dict[str, Any]] = []
    created = 0
    skipped = 0
    sort_items: List[Dict[str, int]] = []

    for index, tag in enumerate(TAGS, start=1):
        name = tag["tagName"]
        tag_id = tag["tagId"]
        existing = find_root_by_name(flat, name)
        if existing:
            category_id = existing.get("id")
            print(f"[跳过] {name} 已存在，本地 id={category_id}")
            skipped += 1
            mapping.append(
                {
                    "tagId": tag_id,
                    "tagName": name,
                    "categoryId": category_id,
                    "created": False,
                }
            )
            if category_id:
                sort_items.append({"id": int(category_id), "sort": index})
            continue

        ok, result = client.create_category(name, parent_id=0)
        if not ok:
            print(f"[失败] {name}: {result}", file=sys.stderr)
            return 1

        category_id = result.get("id") if isinstance(result, dict) else None
        print(f"[创建一级] {name} -> 本地 id={category_id}")
        created += 1
        mapping.append(
            {
                "tagId": tag_id,
                "tagName": name,
                "categoryId": category_id,
                "created": True,
            }
        )
        if category_id:
            sort_items.append({"id": int(category_id), "sort": index})
            flat.append({"id": category_id, "name": name, "parent_id": 0})

    if sort_items:
        ok, msg = client.update_sort(sort_items)
        if ok:
            print("排序已更新")
        else:
            print(f"排序更新失败（分类已写入）: {msg}", file=sys.stderr)

    map_path = Path(args.map_file)
    map_path.write_text(
        json.dumps(mapping, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )
    print(f"映射已写入: {map_path}")
    print(f"完成：新建 {created}，跳过 {skipped}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
