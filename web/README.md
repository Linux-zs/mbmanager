# MBManager Frontend

Vue.js 3 + Element Plus 前端界面

## 安装依赖

```bash
cd web
npm install
```

## 开发模式

```bash
npm run dev
```

访问: http://localhost:3000

## 构建生产版本

```bash
npm run build
```

构建后的文件在 `dist` 目录

## 功能特性

- ✅ 用户登录/登出
- ✅ 仪表盘统计
- ✅ 主机管理（CRUD + 测试连接）
- ✅ 任务管理（CRUD + 立即执行 + 调度配置）
- ✅ 存储管理（CRUD + 测试连接）
- ✅ 通知管理（CRUD + 测试通知）
- ✅ 备份日志查询
- ✅ 用户管理（CRUD）

## 技术栈

- Vue 3 - 渐进式JavaScript框架
- Vue Router - 路由管理
- Pinia - 状态管理
- Element Plus - UI组件库
- Axios - HTTP客户端
- Vite - 构建工具

## 项目结构

```
web/
├── src/
│   ├── api/           # API接口
│   ├── assets/        # 静态资源
│   ├── components/    # 公共组件
│   ├── router/        # 路由配置
│   ├── store/         # 状态管理
│   ├── utils/         # 工具函数
│   ├── views/         # 页面组件
│   ├── App.vue        # 根组件
│   └── main.js        # 入口文件
├── index.html         # HTML模板
├── package.json       # 依赖配置
└── vite.config.js     # Vite配置
```

## 默认账号

- 用户名: admin
- 密码: admin123

## API代理

开发模式下，所有 `/api` 请求会被代理到 `http://localhost:8080`

## 注意事项

1. 确保后端服务已启动（端口8080）
2. 首次运行需要安装依赖
3. 生产环境需要配置正确的API地址
