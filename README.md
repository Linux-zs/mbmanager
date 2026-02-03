# MBManager - MySQL Backup Manager

[English](#english) | [中文](#中文)

---

## English

### Overview

MBManager (MySQL Backup Manager) is a comprehensive web-based MySQL backup management system built with Go and Vue.js. It provides automated backup scheduling, multiple storage options, notification support, and an intuitive web interface for managing MySQL database backups.

### Features

- **🗄️ Multi-Host Management**: Manage multiple MySQL servers from a single interface with host grouping
- **⏰ Flexible Scheduling**: Support for one-time, daily, weekly, and monthly backup schedules
- **💾 Multiple Storage Options**:
  - Local storage
  - SSH remote storage
  - S3-compatible storage (MinIO, AWS S3, etc.)
- **📊 Backup Management**:
  - View backup history with filtering and grouping by host
  - Download backups directly from the web interface
  - Automatic retention policy management
  - Display storage medium information
- **🔔 Notification Support**:
  - Email notifications
  - DingTalk (钉钉)
  - WeCom (企业微信)
  - Feishu (飞书)
  - Slack
  - Custom webhooks
- **📈 Dashboard**: Real-time statistics and monitoring
- **🔐 User Authentication**: Secure login system
- **🐳 Docker Support**: Easy deployment with Docker and Docker Compose

### Technology Stack

**Backend:**
- Go 1.21+
- Gin Web Framework
- GORM (SQLite)
- Cron scheduler

**Frontend:**
- Vue.js 3
- Element Plus UI
- Vite

### Quick Start

#### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/yourusername/mbmanager.git
cd mbmanager
```

2. Start the application:
```bash
docker-compose up -d
```

3. Access the web interface at `http://localhost:8080`

4. Default credentials:
   - Username: `admin`
   - Password: `admin123`

#### Manual Installation

**Prerequisites:**
- Go 1.21 or higher
- Node.js 16+ and npm
- MySQL client tools (mysqldump)

**Backend Setup:**

```bash
# Install dependencies
go mod download

# Build the application
go build -o mbmanager ./cmd/server

# Run the application
./mbmanager
```

**Frontend Setup:**

```bash
cd web

# Install dependencies
npm install

# Build for production
npm run build

# Or run development server
npm run dev
```

### Configuration

The application uses SQLite for data storage and creates necessary directories automatically on first run:

- `/data` - Database and backup storage
- `/logs` - Application logs

### Usage

1. **Add MySQL Hosts**: Configure your MySQL servers in the Hosts management page
2. **Configure Storage**: Set up storage locations (local, SSH, or S3)
3. **Create Backup Tasks**: Define backup schedules and retention policies
4. **Set Up Notifications**: Configure notification channels for backup status alerts
5. **Monitor Backups**: View backup history and download backups as needed

### API Documentation

The application provides a RESTful API. Key endpoints:

- `POST /api/v1/auth/login` - User authentication
- `GET /api/v1/hosts` - List MySQL hosts
- `GET /api/v1/tasks` - List backup tasks
- `GET /api/v1/logs` - View backup logs
- `GET /api/v1/storages` - List storage configurations
- `GET /api/v1/notifications` - List notification configurations

### Development

```bash
# Run backend in development mode
go run ./cmd/server

# Run frontend in development mode
cd web && npm run dev
```

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### License

[MIT License](LICENSE)

---

## 中文

### 概述

MBManager (MySQL备份管理器) 是一个基于Web的MySQL数据库备份管理系统，使用Go和Vue.js构建。它提供自动化备份调度、多种存储选项、通知支持以及直观的Web界面来管理MySQL数据库备份。

### 功能特性

- **🗄️ 多主机管理**: 从单一界面管理多个MySQL服务器，支持主机分组
- **⏰ 灵活调度**: 支持一次性、每日、每周和每月备份计划
- **💾 多种存储选项**:
  - 本地存储
  - SSH远程存储
  - S3兼容存储 (MinIO、AWS S3等)
- **📊 备份管理**:
  - 查看备份历史，支持按主机筛选和分组
  - 直接从Web界面下载备份
  - 自动保留策略管理
  - 显示存储介质信息
- **🔔 通知支持**:
  - 邮件通知
  - 钉钉
  - 企业微信
  - 飞书
  - Slack
  - 自定义Webhook
- **📈 仪表板**: 实时统计和监控
- **🔐 用户认证**: 安全的登录系统
- **🐳 Docker支持**: 使用Docker和Docker Compose轻松部署

### 技术栈

**后端:**
- Go 1.21+
- Gin Web框架
- GORM (SQLite)
- Cron调度器

**前端:**
- Vue.js 3
- Element Plus UI
- Vite

### 快速开始

#### 使用Docker Compose (推荐)

1. 克隆仓库:
```bash
git clone https://github.com/yourusername/mbmanager.git
cd mbmanager
```

2. 启动应用:
```bash
docker-compose up -d
```

3. 访问Web界面: `http://localhost:8080`

4. 默认登录凭据:
   - 用户名: `admin`
   - 密码: `admin123`

#### 手动安装

**前置要求:**
- Go 1.21或更高版本
- Node.js 16+和npm
- MySQL客户端工具 (mysqldump)

**后端设置:**

```bash
# 安装依赖
go mod download

# 编译应用
go build -o mbmanager ./cmd/server

# 运行应用
./mbmanager
```

**前端设置:**

```bash
cd web

# 安装依赖
npm install

# 生产环境构建
npm run build

# 或运行开发服务器
npm run dev
```

### 配置

应用使用SQLite存储数据，首次运行时会自动创建必要的目录:

- `/data` - 数据库和备份存储
- `/logs` - 应用日志

### 使用说明

1. **添加MySQL主机**: 在主机管理页面配置MySQL服务器
2. **配置存储**: 设置存储位置 (本地、SSH或S3)
3. **创建备份任务**: 定义备份计划和保留策略
4. **设置通知**: 配置备份状态告警的通知渠道
5. **监控备份**: 查看备份历史并根据需要下载备份

### API文档

应用提供RESTful API。主要端点:

- `POST /api/v1/auth/login` - 用户认证
- `GET /api/v1/hosts` - 列出MySQL主机
- `GET /api/v1/tasks` - 列出备份任务
- `GET /api/v1/logs` - 查看备份日志
- `GET /api/v1/storages` - 列出存储配置
- `GET /api/v1/notifications` - 列出通知配置

### 开发

```bash
# 开发模式运行后端
go run ./cmd/server

# 开发模式运行前端
cd web && npm run dev
```

### 贡献

欢迎贡献! 请随时提交Pull Request。

### 许可证

[MIT License](LICENSE)

### 截图

#### 仪表板
![Dashboard](docs/screenshots/dashboard.png)

#### 备份管理
![Backup Management](docs/screenshots/backups.png)

#### 主机管理
![Host Management](docs/screenshots/hosts.png)

### 常见问题

**Q: 如何更改默认端口?**
A: 设置环境变量 `SERVER_PORT=端口号`

**Q: 支持哪些MySQL版本?**
A: 支持MySQL 5.7+和MariaDB 10.2+

**Q: 备份文件存储在哪里?**
A: 默认存储在 `/data/backups` 目录，可以在存储配置中自定义

**Q: 如何设置SSH存储?**
A: 在存储管理页面选择SSH类型，填入SSH连接信息和远程路径

### 支持

如有问题或建议，请提交Issue或联系维护者。
