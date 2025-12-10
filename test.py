import math
import json
from typing import List, Dict, Tuple
from datetime import datetime


class DeliveryRouteOptimizer:
    """配送路线优化核心算法（纯计算版，不依赖界面）。"""

    def __init__(self):
        self.locations: List[Dict] = []
        self.name_to_index: Dict[str, int] = {}
        self.distance_cache: Dict[Tuple[int, int], float] = {}

    def add_location(self, name: str, lat: float, lng: float) -> int:
        """添加位置；若已存在返回其索引。"""
        if name in self.name_to_index:
            return self.name_to_index[name]

        idx = len(self.locations)
        self.locations.append({"name": name, "lat": lat, "lng": lng, "index": idx})
        self.name_to_index[name] = idx
        return idx

    def calculate_distance(self, idx1: int, idx2: int) -> float:
        """计算两点球面距离（公里），带缓存。"""
        key = (min(idx1, idx2), max(idx1, idx2))
        if key in self.distance_cache:
            return self.distance_cache[key]

        a, b = self.locations[idx1], self.locations[idx2]
        R = 6371.0
        lat1, lon1 = math.radians(a["lat"]), math.radians(a["lng"])
        lat2, lon2 = math.radians(b["lat"]), math.radians(b["lng"])
        dlat, dlon = lat2 - lat1, lon2 - lon1
        h = (
            math.sin(dlat / 2) ** 2
            + math.cos(lat1) * math.cos(lat2) * math.sin(dlon / 2) ** 2
        )
        dist = R * 2 * math.atan2(math.sqrt(h), math.sqrt(1 - h))

        self.distance_cache[key] = dist
        return dist

    def nearest_neighbor(self, start_idx: int) -> Tuple[List[int], float]:
        """最近邻生成初始路线。"""
        n = len(self.locations)
        if n <= 1:
            return [start_idx], 0.0

        unvisited = set(range(n))
        unvisited.remove(start_idx)
        route = [start_idx]
        total = 0.0
        current = start_idx

        while unvisited:
            nearest_idx, nearest_dist = None, float("inf")
            for nxt in unvisited:
                d = self.calculate_distance(current, nxt)
                if d < nearest_dist:
                    nearest_dist, nearest_idx = d, nxt
            route.append(nearest_idx)
            total += nearest_dist
            unvisited.remove(nearest_idx)
            current = nearest_idx

        return route, total

    def two_opt_optimization(
        self, route: List[int], iterations: int = 100
    ) -> Tuple[List[int], float]:
        """2-opt 改善路线。"""

        def route_distance(indices: List[int]) -> float:
            return sum(
                self.calculate_distance(indices[i], indices[i + 1])
                for i in range(len(indices) - 1)
            )

        best_route = route.copy()
        best_distance = route_distance(best_route)
        n = len(best_route)

        for _ in range(iterations):
            improved = False
            for i in range(1, n - 2):
                for j in range(i + 1, n):
                    if j - i == 1:
                        continue
                    new_route = best_route[:i] + best_route[i:j][::-1] + best_route[j:]
                    new_distance = route_distance(new_route)
                    if new_distance < best_distance:
                        best_route, best_distance, improved = (
                            new_route,
                            new_distance,
                            True,
                        )
            if not improved:
                break

        return best_route, best_distance

    def optimize_route(
        self,
        start_point: str,
        optimization_iterations: int = 100,
    ) -> Dict:
        """执行完整优化流程。"""
        if start_point not in self.name_to_index:
            raise ValueError(f"起点 '{start_point}' 不存在")

        start_idx = self.name_to_index[start_point]
        initial_route, _ = self.nearest_neighbor(start_idx)
        optimized_route, _ = self.two_opt_optimization(
            initial_route, optimization_iterations
        )

        def route_distance(indices: List[int]) -> float:
            return sum(
                self.calculate_distance(indices[i], indices[i + 1])
                for i in range(len(indices) - 1)
            )

        initial_distance = route_distance(initial_route)
        optimized_distance = route_distance(optimized_route)

        initial_names = [self.locations[i]["name"] for i in initial_route]
        optimized_names = [self.locations[i]["name"] for i in optimized_route]
        improvement = (
            (initial_distance - optimized_distance) / initial_distance * 100
            if initial_distance > 0
            else 0.0
        )

        return {
            "initial": {
                "route": initial_names,
                "distance": round(initial_distance, 2),
                "coordinates": [
                    (self.locations[i]["lat"], self.locations[i]["lng"])
                    for i in initial_route
                ],
            },
            "optimized": {
                "route": optimized_names,
                "distance": round(optimized_distance, 2),
                "coordinates": [
                    (self.locations[i]["lat"], self.locations[i]["lng"])
                    for i in optimized_route
                ],
            },
            "improvement_percent": round(improvement, 1),
            "locations": self.locations,
            "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        }

    def clear_data(self):
        """清空存储。"""
        self.locations.clear()
        self.name_to_index.clear()
        self.distance_cache.clear()


def optimize_delivery_route(
    locations: List[Dict[str, float]],
    start_point: str,
    iterations: int = 100,
) -> Dict:
    """
    外部直接调用的简化接口：
    locations: [{'name': str, 'lat': float, 'lng': float}, ...]
    start_point: 起点名称，需在 locations 中存在
    iterations: 2-opt 迭代次数
    """
    optimizer = DeliveryRouteOptimizer()
    for item in locations:
        optimizer.add_location(item["name"], float(item["lat"]), float(item["lng"]))
    return optimizer.optimize_route(start_point, iterations)


def format_report(result: Dict) -> str:
    """将结果格式化为文本报告，便于打印或保存。"""
    lines = [
        "=" * 60,
        "配送路线优化报告",
        f"生成时间: {result['timestamp']}",
        "=" * 60,
        "",
        "【初始路线】",
        f"总距离: {result['initial']['distance']} km",
        "路线顺序: " + " → ".join(result["initial"]["route"]),
        "",
        "【优化后路线】",
        f"总距离: {result['optimized']['distance']} km",
        "路线顺序: " + " → ".join(result["optimized"]["route"]),
        "",
        "【优化效果】",
        f"距离减少: {result['improvement_percent']}%",
        f"节约距离: {result['initial']['distance'] - result['optimized']['distance']:.2f} km",
    ]
    if result["initial"]["distance"] > 0:
        time_saved = (
            (result["initial"]["distance"] - result["optimized"]["distance"]) / 40 * 60
        )
        lines.append(f"节约时间: {time_saved:.1f} 分钟 (按40km/h估算)")
    lines.append("=" * 60)
    return "\n".join(lines)


def format_route_as_json(result: Dict, use_optimized: bool = True) -> str:
    """
    按路线顺序输出 JSON，便于复制到 HTML：
    [
      {"name": "...", "lat": ..., "lng": ...},
      ...
    ]
    """
    key = "optimized" if use_optimized else "initial"
    route = result[key]["route"]
    coords = result[key]["coordinates"]
    payload = [
        {"name": name, "lat": lat, "lng": lng}
        for name, (lat, lng) in zip(route, coords)
    ]
    return json.dumps(payload, ensure_ascii=False, indent=2)


def plot_route(
    result: Dict, save_path: str = "route.png", use_optimized: bool = True
) -> None:
    """
    将路线绘制成图片保存到本地。
    use_optimized=True 绘制优化后路线，False 绘制初始路线。
    """
    try:
        import matplotlib

        matplotlib.use("Agg")  # 使用无界面后端，便于服务器/无显示环境保存图片
        import matplotlib.pyplot as plt
        # 兼容中文及负号显示，防止图片出现乱码或方块
        matplotlib.rcParams["font.sans-serif"] = ["SimHei", "Arial Unicode MS", "DejaVu Sans"]
        matplotlib.rcParams["axes.unicode_minus"] = False
    except ImportError:
        raise ImportError(
            "缺少 matplotlib，请先安装：python3 -m pip install matplotlib"
        )

    key = "optimized" if use_optimized else "initial"
    coords = result[key]["coordinates"]
    names = result[key]["route"]
    lats = [c[0] for c in coords]
    lngs = [c[1] for c in coords]

    plt.figure(figsize=(8, 6))
    plt.plot(lngs, lats, "o-", color="tab:blue", linewidth=2, markersize=6)
    plt.plot(lngs[0], lats[0], "ro", markersize=10, label="起点")
    plt.plot(lngs[-1], lats[-1], "go", markersize=8, label="终点")

    for i, (x, y, name) in enumerate(zip(lngs, lats, names)):
        plt.annotate(
            f"{i + 1}.{name}",
            (x, y),
            xytext=(5, 5),
            textcoords="offset points",
            fontsize=9,
            bbox=dict(boxstyle="round,pad=0.2", fc="white", alpha=0.7),
        )

    plt.title(
        f"{'优化' if use_optimized else '初始'}路线 - 距离 {result[key]['distance']} km"
    )
    plt.xlabel("经度")
    plt.ylabel("纬度")
    plt.grid(True, alpha=0.3)
    plt.legend()
    plt.gca().set_aspect("equal", adjustable="box")
    plt.tight_layout()
    plt.savefig(save_path, dpi=200)
    plt.close()
    print(f"路线图已保存到: {save_path}")


def _demo():
    """示例：直接传入起点和门店列表，打印结果。"""
    example_locations = [
        {"name": "仓库", "lat": 24.958168, "lng": 102.716387},
        {"name": "门店1", "lat": 25.007088, "lng": 102.681202},
        {"name": "门店2", "lat": 25.019417, "lng": 102.687725},
        {"name": "门店3", "lat": 25.019067, "lng": 102.682918},
        {"name": "门店4", "lat": 25.015061, "lng": 102.676181},
        {"name": "门店5", "lat": 24.999349, "lng": 102.682103},
        {"name": "门店6", "lat": 25.027505, "lng": 102.687081},
        {"name": "门店7", "lat": 25.032677, "lng": 102.681888},
        {"name": "门店8", "lat": 25.024511, "lng": 102.663821},
        {"name": "门店9", "lat": 25.030305, "lng": 102.660989},
        {"name": "门店10", "lat": 25.024939, "lng": 102.641333},
        {"name": "门店11", "lat": 25.015839, "lng": 102.650989},
        {"name": "门店12", "lat": 24.997015, "lng": 102.702359},
        {"name": "门店13", "lat": 24.88405, "lng": 102.830268}
    ]

    result = optimize_delivery_route(
        example_locations,
        start_point="仓库",
        iterations=500,
    )
    print(format_report(result))
    # 输出优化后顺序的 JSON，便于复制到 HTML 中查看
    print("\n优化后顺序（JSON）：")
    print(format_route_as_json(result, use_optimized=True))
    # 生成路线可视化图片（优化后与初始各一张）
    plot_route(result, save_path="route_optimized.png", use_optimized=True)
    plot_route(result, save_path="route_initial.png", use_optimized=False)


if __name__ == "__main__":
    _demo()
