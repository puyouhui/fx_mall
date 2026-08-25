# 风行商城多端业务系统

面向商品采购、供应商备货、员工代客开单和同城配送场景的多端商城系统。

仓库采用单体仓库（monorepo）组织方式：一个 Go 后端同时服务微信小程序、运营管理后台、供应商后台，以及员工端、配送端和移动管理端三个 Flutter 应用。系统还包含 Hiprint 打印中转服务和批量调价工具。

> 根目录 README 是项目总览。各子项目中的 README、构建说明和脚本提供更具体的操作信息；如果内容与实际代码冲突，以代码和构建脚本为准。

## 目录

- [系统组成](#系统组成)
- [业务流程](#业务流程)
- [技术架构](#技术架构)
- [目录结构](#目录结构)
- [本地开发](#本地开发)
- [配置说明](#配置说明)
- [构建与部署](#构建与部署)
- [接口与鉴权](#接口与鉴权)
- [开发约定](#开发约定)
- [测试与检查](#测试与检查)
- [已知问题与改进方向](#已知问题与改进方向)

## 系统组成

| 模块 | 目录 | 技术栈 | 主要职责 |
| --- | --- | --- | --- |
| 业务后端 | `go_backend/` | Go 1.20、Gin、MySQL、MinIO | 认证、商品、订单、支付、配送、结算及 WebSocket |
| 客户小程序 | `mini_app/` | UniApp、Vue | 商品浏览、采购单、下单、支付、订单、优惠券、积分 |
| 运营管理后台 | `admin_console/` | Vue 3、Vite、Element Plus | 商品、用户、订单、供应商、配送、财务及系统设置 |
| 供应商后台 | `supplier_console/` | Vue 3、Vite、Element Plus | 商品查看、备货记录、历史订单、付款对账 |
| 员工端 | `employees_app/` | Flutter | 客户维护、代客开单、销售订单、收款申请和销售分成 |
| 配送端 | `distribution_app/` | Flutter | 接单、分供应商取货、路线规划、配送和收入统计 |
| 移动管理端 | `super_app/` | Flutter | 商品、订单、供应商付款和经营统计 |
| 打印中转 | `print_transfer/` | Node.js、Socket.IO | 连接管理后台与 Electron Hiprint 客户端 |
| 批量调价工具 | `scripts/batch_price_tool/` | Python | 通过表格批量维护商品价格 |

## 业务流程

### 采购与履约

```text
客户浏览商品
  -> 加入采购单
  -> 自主下单 / 销售员代客开单
  -> 货到付款或微信在线支付
  -> 供应商查看待备货商品
  -> 配送员接单
  -> 按供应商取货
  -> 路线规划与配送
  -> 送达、收款或确认收货
```

### 财务结算

```text
订单完成
  -> 计算供应商应付款
  -> 计算配送员收入
  -> 计算销售员分成
  -> 管理员审核、计入与结算
```

订单还涉及优惠券、积分抵扣、配送费、加急费、微信退款、收款审核和推荐奖励等业务。修改订单金额或状态时，应同时检查这些关联流程。

## 技术架构

```text
微信小程序 ───────────────┐
运营管理后台 ─────────────┤
供应商后台 ───────────────┤
员工 Flutter App ─────────┼──> Gin API ──> MySQL
配送 Flutter App ─────────┤       │
移动管理 Flutter App ─────┘       ├──> MinIO（图片）
                                  ├──> 微信支付 API
管理后台 ──> Hiprint 中转服务      ├──> 地图/天气服务
配送端 <──── WebSocket ────────────└──> 飞书通知
```

后端是当前系统的业务中心，入口为 `go_backend/cmd/main.go`，统一 API 前缀为 `/api/mini`。主要代码分层如下：

- `internal/api/`：HTTP Handler、鉴权中间件和第三方回调。
- `internal/model/`：数据结构、SQL 查询和核心业务计算。
- `internal/database/`：数据库连接、建表和兼容性升级。
- `internal/config/`：服务、数据库、对象存储和第三方服务配置。
- `internal/utils/`：JWT、密码、MinIO、地图、距离及路线工具。
- `internal/notify/`：飞书通知。

## 目录结构

```text
.
├── go_backend/               # Go 后端
│   ├── cmd/                  # 服务和数据迁移程序入口
│   └── internal/
│       ├── api/              # 接口及鉴权
│       ├── config/           # 配置
│       ├── database/         # 数据库初始化与升级
│       ├── model/            # 模型和业务逻辑
│       ├── notify/           # 通知服务
│       └── utils/            # 通用工具
├── mini_app/                 # UniApp 微信小程序
│   ├── api/                  # 请求封装与业务 API
│   ├── pages/                # 页面
│   └── utils/                # 分享等公共逻辑
├── admin_console/            # Vue 运营管理后台
├── supplier_console/         # Vue 供应商后台
├── employees_app/            # Flutter 员工端
├── distribution_app/         # Flutter 配送端
├── super_app/                # Flutter 移动管理端
├── print_transfer/           # Hiprint 打印中转
└── scripts/
    └── batch_price_tool/     # 批量调价工具
```

## 本地开发

### 一键启动后端和两个 Web 后台

首次完成依赖和数据库配置后，可从仓库根目录启动主要本地服务：

```bash
./start-local.sh
```

脚本会同时启动：

- 本地 MinIO：`http://localhost:9000`（管理界面 `http://localhost:9001`）
- 后端：`http://localhost:8082`
- 运营管理后台：`http://localhost:5173/admin/`
- 供应商后台：`http://localhost:21321/supplier/`

按 `Ctrl+C` 可一起停止三个服务。后端本地环境写在被 Git 忽略的 `go_backend/.env.local`；首次配置可复制 `go_backend/.env.example`。

本地上传文件保存在 Docker Volume `zhujiang_minio_data` 的 `zhujiang` Bucket 中，不会写入线上 MinIO。删除并重建 MinIO 容器不会删除该 Volume；如需清空本地文件，应先确认数据不再需要再删除 Volume。

运行 Flutter 真机应用时，可使用根目录脚本自动注入当前电脑的局域网地址：

```bash
./run-flutter-local.sh employees_app
./run-flutter-local.sh distribution_app
./run-flutter-local.sh super_app
```

### 环境要求

按需要安装对应工具：

- Go 1.20 或兼容版本。
- MySQL 5.7+/8.0。
- 可访问的 MinIO 服务。
- Node.js 16+ 和 npm，用于两个 Vue 后台及打印服务。
- Flutter SDK，Dart 版本需满足各 App 的 `pubspec.yaml`。
- HBuilderX；构建微信小程序时还需要微信开发者工具。
- Python 3，仅使用批量调价工具时需要。

### 1. 启动后端

当前后端配置来自 `go_backend/internal/config/config.go`。首次本地运行前，请准备 MySQL 数据库和 MinIO，并使用本地配置替换对应连接信息。

```bash
cd go_backend
go mod download
go run ./cmd
```

默认监听地址：

```text
http://localhost:8082
```

健康验证可先请求一个公开接口：

```bash
curl http://localhost:8082/api/mini/categories
```

后端启动时会自动创建或补充数据库表，并在没有管理员数据时创建初始管理员。首次部署后必须立即修改初始密码。生产数据库升级前应先备份，不要将启动时自动升级当作可回滚的迁移方案。

### 2. 启动运营管理后台

```bash
cd admin_console
npm install
npm run dev
```

API 地址由 `admin_console/src/utils/request.js` 决定。当前本地访问配置仍指向线上 API；需要连接本地后端时，请将开发地址调整为：

```text
http://localhost:8082/api/mini
```

### 3. 启动供应商后台

```bash
cd supplier_console
npm install
npm run dev
```

本地 API 默认地址：

```text
http://localhost:8082/api/mini/supplier
```

### 4. 运行微信小程序

1. 使用 HBuilderX 打开 `mini_app/`。
2. 在 `mini_app/api/request.js` 中确认 `BASE_URL`。
3. 选择“运行到微信开发者工具”。
4. 在微信开发者工具中配置合法域名、AppID 和调试选项。

本地联调时，真机不能使用电脑的 `localhost`，应改为同一局域网内电脑的 IP 地址。
当前本地配置使用 `http://192.168.1.87:8082/api/mini`；电脑局域网 IP 变化后，需要同步修改 `mini_app/api/request.js` 和 `go_backend/.env.local` 中的 `MINIO_BASE_URL`。

### 5. 运行 Flutter 应用

三个 App 的操作方式一致：

```bash
cd employees_app          # 或 distribution_app / super_app
flutter pub get
flutter run --dart-define=APP_ENV=device
```

常用环境值：

- `device`：局域网真机开发地址。
- `emulator`：Android 模拟器地址，通常使用 `10.0.2.2` 访问宿主机。
- `prod`：生产 API 地址。

具体地址以各 App 的 `lib/utils/config.dart` 为准。切换网络后，通常需要更新 `devBaseUrl`。

### 6. 启动打印中转服务

```bash
cd print_transfer
npm install
npm run serve
```

服务配置位于 `print_transfer/config.json`。生产环境建议启用 HTTPS，并使用独立强 Token。详细配置参见 `print_transfer/README.md` 和 `print_transfer/SSL-CERTIFICATE.md`。

## 配置说明

### API 路径

| 环境/调用方 | 基础路径 |
| --- | --- |
| Go 服务原始路径 | `http://localhost:8082/api/mini` |
| 小程序生产路径 | `https://api.sscchh.com/api/mini` |
| Web 后台生产代理路径 | `/api_mall/mini` |
| 管理员接口 | `{baseUrl}/admin` |
| 供应商接口 | `{baseUrl}/supplier` |
| 员工/配送接口 | `{baseUrl}/employee` |

`/api_mall/mini` 是生产 Nginx 对外代理路径，不是 Gin 内部注册的原始路径。排查 404 时，应同时检查前端基础地址和网关重写规则。

### 后端配置项

后端主要需要以下配置：

- HTTP 端口和超时。
- MySQL 地址、账号、密码和数据库名。
- MinIO 地址、Access Key、Secret Key、Bucket 和公共访问 URL。
- 微信小程序 AppID、AppSecret。
- 地图服务 Key。
- 配送端与管理后台 WebSocket 地址。
- 微信支付、飞书、打印地址等部分运行配置存储在系统设置表中。

> 安全要求：不要在提交、日志、截图或文档中暴露数据库密码、JWT 密钥、MinIO 密钥、小程序密钥、微信支付证书和地图 Key。当前配置仍存在硬编码，建议优先迁移到环境变量或密钥管理服务，并轮换已经进入版本历史的凭据。

## 构建与部署

### Go 后端

生成 Linux AMD64 静态二进制：

```bash
cd go_backend
./build_bt.sh
```

生成 Linux ARM64 二进制：

```bash
./build_bt.sh arm64
```

输出位于 `go_backend/dist/`。

### Vue 后台

```bash
cd admin_console          # 或 supplier_console
npm install
npm run build
```

构建产物位于 `dist/`。仓库中的 Dockerfile 使用“先构建、再复制 dist 到 Nginx 镜像”的方式：

```bash
docker build -t admin-console:latest .
```

运营管理后台容器默认监听 `10080`，供应商后台容器默认监听 `5174`。API 反向代理由外层网关负责，子项目自带的 Nginx 配置只提供静态文件和 History 路由回退。

### Flutter Android

各 App 已提供生产构建脚本：

```bash
cd distribution_app      # 或 employees_app / super_app
./build-prod.sh
```

等价的核心命令为：

```bash
flutter build apk --release --dart-define=APP_ENV=prod
```

APK 默认输出到：

```text
build/app/outputs/flutter-apk/app-release.apk
```

### 打印中转服务

打印服务支持直接运行、打包和 Docker 部署，参见：

- `print_transfer/README.md`
- `print_transfer/README-DOCKER.md`
- `print_transfer/BUILD.md`

## 接口与鉴权

后端目前按身份划分四类接口：

| 身份 | 路径 | 鉴权方式 |
| --- | --- | --- |
| 小程序用户 | `/api/mini/mini-app/users/*` | 小程序用户 Bearer Token |
| 管理员 | `/api/mini/admin/*` | 管理员 Bearer Token |
| 供应商 | `/api/mini/supplier/*` | 供应商 Bearer Token |
| 员工/配送员 | `/api/mini/employee/*` | 员工 Bearer Token |

公开接口包括商品、分类、轮播图、小程序登录、各角色登录，以及微信支付回调等。WebSocket 接口在 Handler 内验证 Token。

项目统一响应通常采用以下结构：

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

新增接口时，应复用已有响应方法和对应身份的鉴权中间件，避免在前端单独适配另一套响应结构。

## 开发约定

### 修改业务时的检查范围

- 商品：同时检查小程序、管理后台、供应商端、员工开单和计量单位。
- 订单：同时检查支付、配送、退款、供应商货款、配送收入及销售分成。
- 用户与地址：同时检查小程序用户、销售员代客维护、地图坐标和配送端展示。
- API 路径：同步检查 Vue 请求封装、Flutter `config.dart` 和生产 Nginx 重写。
- 图片字段：确认 MinIO 公共 URL、旧图片兼容和默认图片逻辑。
- 状态值：先查找所有端对该状态的判断，不要只修改后端常量或单个页面。

### 数据库变更

目前表结构升级集中在 `go_backend/internal/database/database.go`。新增字段或索引时应做到：

1. 对已有数据库可重复执行。
2. 明确默认值和旧数据兼容策略。
3. 在生产执行前完成备份。
4. 同步更新模型查询中的 `SELECT`、`INSERT` 和 `Scan` 顺序。

长期建议引入带版本号和回滚说明的数据库迁移工具。

### Git 提交

- 不提交真实密码、Token、证书私钥和本地环境配置。
- 不提交 `node_modules/`、Flutter 构建目录、APK、Vue `dist/` 或 Go 构建产物。
- 一个提交尽量只处理一个业务主题，并在提交信息中说明受影响的端。

## 测试与检查

当前自动化测试覆盖较少，提交前至少执行与改动相关的检查：

```bash
# Go
cd go_backend
gofmt -w <修改过的 Go 文件>
go test ./...

# Vue
cd admin_console          # 或 supplier_console
npm run build

# Flutter
cd employees_app          # 或 distribution_app / super_app
flutter analyze
flutter test
```

小程序应至少在微信开发者工具中验证：登录、首页加载、采购单、下单、支付分支和订单详情。

涉及支付、退款、金额计算、订单状态、配送完成或财务结算的改动，需要在隔离测试环境中完成端到端验证，禁止直接使用生产订单试错。

## 已知问题与改进方向

以下是接手项目时应优先关注的技术债：

1. **配置安全**：部分数据库、对象存储和第三方服务凭据仍硬编码在源码中，打印服务目录还包含证书文件。
2. **支付缓存**：微信预支付数据保存在后端进程内存中，重启或多实例部署可能导致回调无法取得原始下单数据。
3. **数据库迁移**：大量建表和升级逻辑集中在启动流程中，缺少明确版本与回滚机制。
4. **单文件过大**：订单、配送、员工销售、数据库初始化及部分页面文件职责过重。
5. **自动化测试不足**：后端与两个 Vue 后台几乎没有业务测试，Flutter 测试也以模板为主。
6. **环境配置分散**：多个端分别维护 API、图片和地图地址，切换环境容易遗漏。
7. **历史兼容逻辑较多**：订单状态、支付方式和图片地址存在旧数据兼容分支，重构前需要先确认线上数据情况。

建议的治理顺序：先完成密钥轮换与配置外置，再补订单/支付/金额计算测试，然后引入数据库迁移，最后逐步拆分超大文件和统一多端环境配置。
