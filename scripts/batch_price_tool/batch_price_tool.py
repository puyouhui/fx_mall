#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
批量改价工具 - 按分类批量调整商品成本、批发价、零售价
使用后台 API 批量请求，带 GUI 界面
"""

import tkinter as tk
from tkinter import ttk, messagebox, scrolledtext
import requests
import json
import threading
from typing import Optional, List, Dict, Any

# ============ 配置 ============
DEFAULT_BASE_URL = "https://mall.sscchh.com/api_mall/mini"


class ApiClient:
    """后台 API 客户端"""
    
    def __init__(self, base_url: str, token: Optional[str] = None):
        self.base_url = base_url.rstrip("/")
        self.token = token
        self.session = requests.Session()
        self.session.headers["Content-Type"] = "application/json"
    
    def _url(self, path: str) -> str:
        return f"{self.base_url}{path}"
    
    def _headers(self) -> dict:
        h = {"Content-Type": "application/json"}
        if self.token:
            h["Authorization"] = f"Bearer {self.token}"
        return h
    
    def login(self, username: str, password: str) -> tuple[bool, str]:
        """登录，返回 (成功, token 或错误信息)"""
        try:
            r = self.session.post(
                self._url("/admin/login"),
                json={"username": username, "password": password},
                timeout=15
            )
            data = r.json()
            if r.status_code == 200 and data.get("code") == 200:
                self.token = data.get("data", {}).get("token") or data.get("token")
                if self.token:
                    return True, self.token
            return False, data.get("message", "登录失败")
        except Exception as e:
            return False, str(e)
    
    def get_categories(self) -> tuple[bool, Any]:
        """获取分类列表（树形），返回 (成功, 数据 或 错误)"""
        try:
            r = self.session.get(self._url("/admin/categories"), headers=self._headers(), timeout=15)
            data = r.json()
            if r.status_code == 200 and data.get("code") == 200:
                return True, data.get("data", data)
            return False, data.get("message", "获取分类失败")
        except Exception as e:
            return False, str(e)
    
    def get_products_by_category(self, category_id: int, page_num: int = 1, page_size: int = 500) -> tuple[bool, Any]:
        """按分类获取商品列表"""
        try:
            r = self.session.get(
                self._url("/admin/products"),
                params={"categoryId": category_id, "pageNum": page_num, "pageSize": page_size},
                headers=self._headers(),
                timeout=30
            )
            data = r.json()
            if r.status_code == 200 and data.get("code") == 200:
                d = data.get("data", data)
                if isinstance(d, dict):
                    return True, d.get("list", []), d.get("total", 0)
                return True, d if isinstance(d, list) else [], 0
            return False, [], 0
        except Exception as e:
            return False, [], 0
    
    def update_product(self, product_id: int, product_data: dict) -> tuple[bool, str]:
        """更新商品"""
        try:
            r = self.session.put(
                self._url(f"/admin/products/{product_id}"),
                json=product_data,
                headers=self._headers(),
                timeout=15
            )
            data = r.json()
            if r.status_code == 200 and data.get("code") == 200:
                return True, ""
            return False, data.get("message", "更新失败")
        except Exception as e:
            return False, str(e)


def flatten_categories(cats: list, prefix: str = "") -> list[tuple[int, str]]:
    """扁平化分类树为 (id, 显示名) 列表"""
    result = []
    for c in cats:
        name = c.get("name", "未命名")
        cid = c.get("id")
        if cid is not None:
            result.append((cid, f"{prefix}{name}"))
        children = c.get("children", [])
        if children:
            result.extend(flatten_categories(children, prefix=f"{prefix}{name} / "))
    return result


def apply_price_delta(specs: list, cost_delta: float, wholesale_delta: float, retail_delta: float) -> list:
    """对每个规格应用价格调整"""
    new_specs = []
    for s in specs:
        ns = dict(s)
        ns["cost"] = max(0, (ns.get("cost") or 0) + cost_delta)
        ns["wholesale_price"] = max(0, (ns.get("wholesale_price") or ns.get("wholesalePrice") or 0) + wholesale_delta)
        ns["retail_price"] = max(0, (ns.get("retail_price") or ns.get("retailPrice") or 0) + retail_delta)
        new_specs.append(ns)
    return new_specs


class BatchPriceApp:
    def __init__(self):
        self.root = tk.Tk()
        self.root.title("批量改价工具")
        self.root.geometry("580x620")
        self.root.minsize(500, 500)
        
        self.client: Optional[ApiClient] = None
        self.categories_flat: list = []
        
        self._build_ui()
    
    def _build_ui(self):
        main = ttk.Frame(self.root, padding=12)
        main.pack(fill=tk.BOTH, expand=True)
        
        # 登录区
        login_fr = ttk.LabelFrame(main, text="登录", padding=8)
        login_fr.pack(fill=tk.X, pady=(0, 8))
        ttk.Label(login_fr, text="API 地址:").grid(row=0, column=0, sticky=tk.W, pady=2)
        self.base_url_var = tk.StringVar(value=DEFAULT_BASE_URL)
        ttk.Entry(login_fr, textvariable=self.base_url_var, width=50).grid(row=0, column=1, sticky=tk.EW, padx=4, pady=2)
        ttk.Label(login_fr, text="用户名:").grid(row=1, column=0, sticky=tk.W, pady=2)
        self.username_var = tk.StringVar()
        ttk.Entry(login_fr, textvariable=self.username_var, width=30).grid(row=1, column=1, sticky=tk.W, padx=4, pady=2)
        ttk.Label(login_fr, text="密码:").grid(row=2, column=0, sticky=tk.W, pady=2)
        self.password_var = tk.StringVar()
        ttk.Entry(login_fr, textvariable=self.password_var, width=30, show="*").grid(row=2, column=1, sticky=tk.W, padx=4, pady=2)
        ttk.Button(login_fr, text="登录", command=self._do_login).grid(row=3, column=1, sticky=tk.W, pady=4)
        login_fr.columnconfigure(1, weight=1)
        
        # 分类选择
        cat_fr = ttk.LabelFrame(main, text="选择分类", padding=8)
        cat_fr.pack(fill=tk.X, pady=(0, 8))
        ttk.Label(cat_fr, text="分类:").grid(row=0, column=0, sticky=tk.W, pady=2)
        self.category_combo = ttk.Combobox(cat_fr, width=55, state="readonly")
        self.category_combo.grid(row=0, column=1, sticky=tk.EW, padx=4, pady=2)
        ttk.Button(cat_fr, text="刷新分类", command=self._refresh_categories).grid(row=1, column=1, sticky=tk.W, pady=4)
        cat_fr.columnconfigure(1, weight=1)
        
        # 价格调整
        price_fr = ttk.LabelFrame(main, text="价格调整（正数为涨价，负数为降价）", padding=8)
        price_fr.pack(fill=tk.X, pady=(0, 8))
        ttk.Label(price_fr, text="成本调整(元):").grid(row=0, column=0, sticky=tk.W, pady=2)
        self.cost_delta_var = tk.StringVar(value="0")
        ttk.Entry(price_fr, textvariable=self.cost_delta_var, width=15).grid(row=0, column=1, sticky=tk.W, padx=4, pady=2)
        ttk.Label(price_fr, text="批发价调整(元):").grid(row=1, column=0, sticky=tk.W, pady=2)
        self.wholesale_delta_var = tk.StringVar(value="0")
        ttk.Entry(price_fr, textvariable=self.wholesale_delta_var, width=15).grid(row=1, column=1, sticky=tk.W, padx=4, pady=2)
        ttk.Label(price_fr, text="零售价调整(元):").grid(row=2, column=0, sticky=tk.W, pady=2)
        self.retail_delta_var = tk.StringVar(value="0")
        ttk.Entry(price_fr, textvariable=self.retail_delta_var, width=15).grid(row=2, column=1, sticky=tk.W, padx=4, pady=2)
        
        # 执行
        btn_fr = ttk.Frame(main)
        btn_fr.pack(fill=tk.X, pady=(0, 8))
        self.run_btn = ttk.Button(btn_fr, text="开始批量改价", command=self._do_batch_update)
        self.run_btn.pack(side=tk.LEFT, padx=(0, 8))
        self.progress_var = tk.StringVar(value="")
        ttk.Label(btn_fr, textvariable=self.progress_var).pack(side=tk.LEFT)
        
        # 日志
        log_fr = ttk.LabelFrame(main, text="执行日志", padding=4)
        log_fr.pack(fill=tk.BOTH, expand=True, pady=(0, 8))
        self.log_text = scrolledtext.ScrolledText(log_fr, height=14, wrap=tk.WORD, state=tk.DISABLED)
        self.log_text.pack(fill=tk.BOTH, expand=True)
    
    def _log(self, msg: str):
        self.log_text.configure(state=tk.NORMAL)
        self.log_text.insert(tk.END, msg + "\n")
        self.log_text.see(tk.END)
        self.log_text.configure(state=tk.DISABLED)
        self.root.update_idletasks()
    
    def _do_login(self):
        base = self.base_url_var.get().strip()
        username = self.username_var.get().strip()
        password = self.password_var.get()
        if not base or not username or not password:
            messagebox.showwarning("提示", "请填写 API 地址、用户名和密码")
            return
        self.run_btn.configure(state=tk.DISABLED)
        self._log("正在登录...")
        
        def task():
            client = ApiClient(base)
            ok, res = client.login(username, password)
            self.root.after(0, lambda: self._on_login_done(ok, res, client))
        
        threading.Thread(target=task, daemon=True).start()
    
    def _on_login_done(self, ok: bool, res, client: ApiClient):
        self.run_btn.configure(state=tk.NORMAL)
        if ok:
            self.client = client
            self._log("登录成功")
            self._refresh_categories()
        else:
            self._log(f"登录失败: {res}")
            messagebox.showerror("登录失败", str(res))
    
    def _refresh_categories(self):
        if not self.client:
            self._log("请先登录")
            return
        self._log("正在加载分类...")
        
        def task():
            ok, data = self.client.get_categories()
            self.root.after(0, lambda: self._on_categories_done(ok, data))
        
        threading.Thread(target=task, daemon=True).start()
    
    def _on_categories_done(self, ok: bool, data):
        if not ok:
            self._log(f"获取分类失败: {data}")
            messagebox.showerror("错误", str(data))
            return
        cats = data if isinstance(data, list) else []
        self.categories_flat = flatten_categories(cats)
        names = [n for _, n in self.categories_flat]
        self.category_combo["values"] = names
        if names:
            self.category_combo.current(0)
        self._log(f"已加载 {len(self.categories_flat)} 个分类")
    
    def _do_batch_update(self):
        if not self.client:
            messagebox.showwarning("提示", "请先登录")
            return
        idx = self.category_combo.current()
        if idx < 0 or idx >= len(self.categories_flat):
            messagebox.showwarning("提示", "请选择分类")
            return
        
        try:
            cost_d = float(self.cost_delta_var.get().strip() or "0")
            wholesale_d = float(self.wholesale_delta_var.get().strip() or "0")
            retail_d = float(self.retail_delta_var.get().strip() or "0")
        except ValueError:
            messagebox.showwarning("提示", "价格调整请输入数字")
            return
        
        if cost_d == 0 and wholesale_d == 0 and retail_d == 0:
            messagebox.showwarning("提示", "请至少填写一项非零的价格调整")
            return
        
        category_id, category_name = self.categories_flat[idx]
        self.run_btn.configure(state=tk.DISABLED)
        self.progress_var.set("执行中...")
        self._log(f"开始批量改价: 分类={category_name}, 成本{cost_d:+.2f}, 批发{wholesale_d:+.2f}, 零售{retail_d:+.2f}")
        
        def task():
            self._run_batch(category_id, category_name, cost_d, wholesale_d, retail_d)
        
        threading.Thread(target=task, daemon=True).start()
    
    def _run_batch(self, category_id: int, category_name: str, cost_d: float, wholesale_d: float, retail_d: float):
        products: List[Dict] = []
        page = 1
        while True:
            ok, lst, total = self.client.get_products_by_category(category_id, page, 200)
            if not ok or not lst:
                break
            products.extend(lst)
            if len(lst) < 200 or len(products) >= total:
                break
            page += 1
        
        total_products = len(products)
        self.root.after(0, lambda: self._log(f"该分类下共 {total_products} 个商品"))
        
        success = 0
        fail = 0
        for i, p in enumerate(products):
            pid = p.get("id")
            name = p.get("name", "")
            self.root.after(0, lambda i=i, t=total_products: self.progress_var.set(f"处理中 {i+1}/{t}"))
            
            specs = p.get("specs") or []
            if not specs:
                self.root.after(0, lambda n=name: self._log(f"  跳过 [{n}] (无规格)"))
                continue
            
            new_specs = apply_price_delta(specs, cost_d, wholesale_d, retail_d)
            payload = dict(p)
            payload["specs"] = new_specs
            
            ok, err = self.client.update_product(pid, payload)
            if ok:
                success += 1
                self.root.after(0, lambda n=name: self._log(f"  ✓ [{n}]"))
            else:
                fail += 1
                self.root.after(0, lambda n=name, e=err: self._log(f"  ✗ [{n}] {e}"))
        
        self.root.after(0, lambda: self._on_batch_done(success, fail, total_products))
    
    def _on_batch_done(self, success: int, fail: int, total: int):
        self.run_btn.configure(state=tk.NORMAL)
        self.progress_var.set("")
        msg = f"完成: 成功 {success}, 失败 {fail}, 共 {total} 个商品"
        self._log(msg)
        messagebox.showinfo("完成", msg)
    
    def run(self):
        self.root.mainloop()


def main():
    print("正在启动批量改价工具（图形界面）...")
    try:
        app = BatchPriceApp()
        app.run()
    except Exception as e:
        print("启动失败:", e)
        raise


if __name__ == "__main__":
    main()
