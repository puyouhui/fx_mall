# 供应商后台管理系统

供应商后台管理系统，用于供应商管理商品、订单等业务。

## 技术栈

- Vue 3
- Vite
- Element Plus
- Vue Router
- Axios

## 项目结构

```
supplier_console/
├── src/
│   ├── api/              # API 接口
│   │   └── auth.js       # 认证相关接口
│   ├── layout/           # 布局组件
│   │   └── Layout.vue    # 主布局
│   ├── router/           # 路由配置
│   │   └── index.js      # 路由定义
│   ├── utils/            # 工具函数
│   │   └── request.js    # Axios 请求封装
│   ├── views/            # 页面组件
│   │   ├── Login.vue     # 登录页
│   │   ├── Dashboard.vue # 仪表盘
│   │   ├── Products.vue  # 商品管理
│   │   └── Orders.vue    # 订单管理
│   ├── App.vue           # 根组件
│   └── main.js           # 入口文件
├── index.html            # HTML 模板
└── vite.config.js        # Vite 配置
```

## 开发

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## API 配置

API 基础 URL 配置在 `src/utils/request.js` 中，默认指向：

```
http://localhost:8082/api/supplier
```

请根据实际后端地址修改。

## 功能模块

- **登录认证**：供应商登录功能
- **仪表盘**：数据统计展示
- **商品管理**：商品列表、添加、编辑、删除
- **订单管理**：订单列表、订单详情查看

## 注意事项

1. 项目使用 localStorage 存储 token，请确保在生产环境中使用更安全的存储方式
2. API 接口需要根据实际后端接口进行调整
3. 部分功能（如商品编辑、订单详情）需要进一步完善
