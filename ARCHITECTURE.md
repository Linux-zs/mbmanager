# 项目结构说明

```
mbmanager/
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口
├── internal/
│   ├── api/
│   │   ├── handler/               # HTTP处理器
│   │   │   ├── auth.go            # 认证处理器
│   │   │   └── handlers.go        # 其他处理器
│   │   ├── middleware/            # 中间件
│   │   │   ├── auth.go            # JWT认证中间件
│   │   │   └── cors.go            # CORS中间件
│   │   └── router.go              # 路由配置
│   ├── model/                     # 数据模型
│   │   ├── host.go                # 主机模型
│   │   ├── task.go                # 任务模型
│   │   ├── storage.go             # 存储模型
│   │   ├── notification.go        # 通知模型
│   │   ├── log.go                 # 日志模型
│   │   └── user.go                # 用户模型
│   ├── backup/                    # 备份执行器
│   │   ├── executor.go            # 执行器接口
│   │   ├── mysqldump.go           # mysqldump实现
│   │   ├── mydumper.go            # mydumper实现
│   │   └── xtrabackup.go          # xtrabackup实现
│   ├── storage/                   # 存储适配器
│   │   ├── storage.go             # 存储接口
│   │   ├── local.go               # 本地存储
│   │   ├── s3.go                  # S3存储
│   │   ├── oss.go                 # 阿里云OSS
│   │   └── nas.go                 # NAS存储
│   ├── notification/              # 通知适配器
│   │   ├── notifier.go            # 通知接口
│   │   ├── email.go               # 邮件通知
│   │   ├── dingtalk.go            # 钉钉通知
│   │   └── wecom.go               # 企业微信通知
│   ├── config/                    # 配置管理
│   │   └── config.go              # 配置加载
│   └── database/                  # 数据库初始化
│       └── db.go                  # 数据库连接和迁移
├── web/                           # 前端资源
│   ├── static/                    # 静态文件
│   │   ├── css/
│   │   ├── js/
│   │   └── img/
│   └── templates/                 # HTML模板
│       └── index.html             # 登录页面
├── docker/                        # Docker配置
│   ├── Dockerfile                 # Docker镜像构建文件
│   └── docker-compose.yml         # Docker Compose配置
├── bin/                           # 编译输出目录
├── data/                          # 数据目录
│   ├── db/                        # SQLite数据库
│   └── backups/                   # 本地备份存储
├── go.mod                         # Go模块定义
├── go.sum                         # Go模块校验
├── Makefile                       # 构建脚本
├── .gitignore                     # Git忽略文件
├── README.md                      # 项目说明
└── QUICKSTART.md                  # 快速启动指南
```

## 核心模块说明

### 1. cmd/server
应用程序入口，负责：
- 加载配置
- 初始化数据库
- 创建HTTP服务器
- 启动应用

### 2. internal/api
HTTP API层，包括：
- **handler**: 处理HTTP请求，调用service层
- **middleware**: 中间件（认证、CORS等）
- **router**: 路由配置

### 3. internal/model
数据模型层，定义：
- 数据库表结构
- GORM模型
- JSON序列化

### 4. internal/backup
备份执行器，实现：
- 备份执行器接口
- mysqldump备份
- mydumper备份
- xtrabackup备份（通过SSH）

### 5. internal/storage
存储适配器，实现：
- 存储接口
- 本地文件系统存储
- AWS S3存储
- 阿里云OSS存储
- NAS存储

### 6. internal/notification
通知适配器，实现：
- 通知接口
- 邮件通知
- 钉钉机器人通知
- 企业微信机器人通知

### 7. internal/config
配置管理，负责：
- 加载环境变量
- 提供配置访问接口

### 8. internal/database
数据库管理，负责：
- 初始化SQLite连接
- 自动迁移表结构
- 创建默认数据

## 数据流

```
HTTP请求
  ↓
Router (路由)
  ↓
Middleware (中间件：认证、CORS)
  ↓
Handler (处理器)
  ↓
Service (业务逻辑) [待实现]
  ↓
Repository (数据访问) [待实现]
  ↓
Database (数据库)
```

## 备份流程

```
任务调度器 [待实现]
  ↓
备份服务 [待实现]
  ↓
备份执行器 (mysqldump/mydumper/xtrabackup)
  ↓
生成备份文件
  ↓
存储适配器 (local/s3/oss/nas)
  ↓
保存备份文件
  ↓
通知适配器 (email/dingtalk/wecom)
  ↓
发送通知
```

## 待实现功能

### 高优先级
1. **任务调度服务** (internal/service/scheduler_service.go)
   - 使用gocron v2实现定时任务
   - 支持多种调度类型
   - 任务并发控制

2. **备份服务** (internal/service/backup_service.go)
   - 执行备份任务
   - 调用备份执行器
   - 上传到存储
   - 发送通知
   - 清理过期备份

3. **Repository层** (internal/repository/)
   - 封装数据库操作
   - 提供统一的数据访问接口

4. **Service层** (internal/service/)
   - 实现业务逻辑
   - 调用repository和其他服务

### 中优先级
5. **Vue.js前端**
   - 使用Vue 3 + Element Plus
   - 实现所有管理页面
   - 实现实时状态更新

6. **备份文件管理**
   - 浏览备份文件
   - 下载备份文件
   - 删除备份文件

7. **备份恢复功能**
   - 选择备份文件
   - 恢复到指定数据库

### 低优先级
8. **监控和告警**
   - 备份成功率统计
   - 存储空间监控
   - 异常告警

9. **增量备份**
   - 支持增量备份
   - 备份链管理

10. **备份加密**
    - 备份文件加密
    - 密钥管理

## 技术选型理由

### 后端
- **Go**: 高性能、并发支持好、部署简单
- **Gin**: 轻量级、性能好、生态成熟
- **GORM**: 功能完整、易用
- **SQLite**: 无需额外部署、适合单机应用
- **gocron**: 简单易用的任务调度库

### 前端
- **Vue 3**: 渐进式框架、学习曲线平缓
- **Element Plus**: 企业级UI组件库、组件丰富

### 存储
- **本地**: 简单直接
- **S3**: 标准对象存储协议、兼容性好
- **OSS**: 国内访问速度快
- **NAS**: 适合企业内网环境

## 开发建议

1. **代码规范**
   - 使用gofmt格式化代码
   - 遵循Go代码规范
   - 添加必要的注释

2. **错误处理**
   - 所有错误都要处理
   - 使用fmt.Errorf包装错误
   - 记录详细的错误日志

3. **测试**
   - 为核心功能编写单元测试
   - 使用testify进行断言
   - 保持测试覆盖率

4. **安全**
   - 密码使用bcrypt加密
   - 使用JWT进行认证
   - 防止SQL注入
   - 验证用户输入

5. **性能**
   - 使用连接池
   - 避免N+1查询
   - 合理使用缓存
   - 异步处理耗时操作
