## 宝塔（Linux）运行：编译与启动

### 1) 本地编译 Linux 可执行文件

在你的电脑上执行：

```bash
cd /Users/puyouhui/Desktop/mini_mall/go_backend
bash build_bt.sh
```

默认输出：
- `go_backend/dist/go_backend_linux_amd64`

如果你的服务器是 ARM 架构（少数情况）：

```bash
GOARCH=arm64 bash build_bt.sh
```

输出：
- `go_backend/dist/go_backend_linux_arm64`

### 2) 上传到服务器

把生成的文件上传到宝塔服务器（例如 `/www/wwwroot/mini_mall/go_backend/`），并重命名为 `go_backend`：

```bash
mv go_backend_linux_amd64 go_backend
chmod +x go_backend
```

### 3) 启动（最简单方式）

```bash
./go_backend
```

默认端口：`8082`

### 4) 宝塔里长期运行（推荐）

建议用宝塔的 **“进程守护/ Supervisor”** 来托管：
- 启动命令：`/www/wwwroot/mini_mall/go_backend/go_backend`
- 工作目录：`/www/wwwroot/mini_mall/go_backend`
- 重启策略：自动重启

> 说明：当前后端配置写在代码里（`internal/config/config.go`），若服务器的数据库/MinIO 等配置与本地不同，需要改代码后重新编译，或再做一次“配置改为环境变量/配置文件”的改造。




