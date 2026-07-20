# 闪视前端

「闪视」移动端短视频 H5 前端工程，基于 Vue 3、TypeScript 和 Vite 构建。

## 技术栈

- Vue 3 `<script setup>` + TypeScript
- Vite 6
- Vue Router 4
- Pinia
- Tailwind CSS
- Vant 4
- Axios
- vue3-video-play
- vue-virtual-scroller
- FontAwesome 6

## 环境要求

- Node.js 20 或更高版本
- npm 10 或更高版本

## 安装与启动

从仓库根目录执行：

```bash
cd frontend
npm install
npm run dev
```

默认开发地址为 `http://localhost:5173`。

## 常用命令

```bash
# 启动开发服务器
npm run dev

# TypeScript 类型检查
npm run type-check

# ESLint 检查
npm run lint

# 测试环境构建
npm run build:test

# 生产环境构建并生成 gzip 文件
npm run build

# 预览生产构建
npm run preview
```

生产构建产物输出到 `frontend/dist/`。

## 环境配置

| 文件               | 用途         |
| ------------------ | ------------ |
| `.env.development` | 本地开发环境 |
| `.env.test`        | 测试环境     |
| `.env.production`  | 生产环境     |

接口基础地址通过 `VITE_API_BASE_URL` 配置。开发环境默认使用 `/api`，需要由本地代理或后端服务提供对应接口。

## 目录结构

```text
frontend/
├── src/
│   ├── api/          # Axios 请求与接口定义
│   ├── assets/       # 全局样式与静态资源
│   ├── components/   # 公共组件
│   ├── data/         # 开发演示数据
│   ├── router/       # 路由与守卫
│   ├── store/        # Pinia 状态模块
│   ├── types/        # TypeScript 类型定义
│   ├── utils/        # 工具函数
│   └── views/        # 页面组件
├── index.html
├── package.json
├── tailwind.config.ts
└── vite.config.ts
```

当前视频 Feed 使用 `src/data/mockVideos.ts` 中的演示数据。真实后端接入点位于 `src/api/`。

交互原型保留在仓库根目录的 `../proto/prototype.html`，仅作为 UI 参考，不参与前端打包。
