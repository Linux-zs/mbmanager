# 备份工具安装说明

## 概述

mbmanager 使用系统命令行工具执行数据库备份。在使用前，需要在系统中安装相应的备份工具。

## 支持的备份工具

### 1. mysqldump（推荐）
MySQL官方备份工具，适合中小型数据库。

**Linux安装：**
```bash
# Ubuntu/Debian
sudo apt-get install mysql-client

# CentOS/RHEL
sudo yum install mysql

# 验证安装
mysqldump --version
```

**Windows安装：**
1. 下载MySQL安装包：https://dev.mysql.com/downloads/mysql/
2. 安装时选择"Client Programs"
3. 将MySQL bin目录添加到系统PATH
4. 验证：在cmd中运行 `mysqldump --version`

**常用选项（在任务管理中配置）：**
```json
{
  "single-transaction": true,
  "routines": true,
  "triggers": true,
  "events": true,
  "hex-blob": true
}
```

### 2. mydumper
多线程备份工具，适合大型数据库，速度更快。

**Linux安装：**
```bash
# Ubuntu/Debian
sudo apt-get install mydumper

# CentOS/RHEL
sudo yum install mydumper

# 验证安装
mydumper --version
```

**Windows安装：**
1. 下载预编译版本：https://github.com/mydumper/mydumper/releases
2. 解压到指定目录
3. 将目录添加到系统PATH

**常用选项：**
```json
{
  "threads": 4,
  "compress": true,
  "chunk-filesize": 128
}
```

### 3. xtrabackup
Percona的物理备份工具，支持热备份，适合大型生产环境。

**Linux安装：**
```bash
# Ubuntu/Debian
wget https://repo.percona.com/apt/percona-release_latest.generic_all.deb
sudo dpkg -i percona-release_latest.generic_all.deb
sudo apt-get update
sudo apt-get install percona-xtrabackup-80

# CentOS/RHEL
sudo yum install https://repo.percona.com/yum/percona-release-latest.noarch.rpm
sudo yum install percona-xtrabackup-80

# 验证安装
xtrabackup --version
```

**Windows安装：**
xtrabackup 不支持Windows，建议使用mysqldump或mydumper。

**常用选项：**
```json
{
  "parallel": 4,
  "compress": true,
  "compress-threads": 4
}
```

## 当前环境检查

### Windows环境
如果你在Windows上运行mbmanager，建议：
1. 安装MySQL Client（包含mysqldump）
2. 或使用mydumper的Windows版本
3. xtrabackup不支持Windows

### Linux环境（推荐）
生产环境建议在Linux上运行，可以使用所有三种备份工具。

## 故障排查

### 问题1：执行备份任务失败，没有日志
**原因**：备份工具未安装或不在PATH中

**解决方法**：
1. 检查工具是否安装：
   ```bash
   # Windows
   where mysqldump
   where mydumper

   # Linux
   which mysqldump
   which mydumper
   which xtrabackup
   ```

2. 如果未找到，按上述说明安装

3. 确保工具在系统PATH中

### 问题2：权限错误
**原因**：MySQL用户权限不足

**解决方法**：
确保MySQL用户具有以下权限：
```sql
GRANT SELECT, RELOAD, LOCK TABLES, REPLICATION CLIENT, SHOW VIEW, EVENT, TRIGGER ON *.* TO 'backup_user'@'%';
FLUSH PRIVILEGES;
```

### 问题3：连接失败
**原因**：网络或防火墙问题

**解决方法**：
1. 检查MySQL服务是否运行
2. 检查防火墙规则
3. 使用主机管理的"测试连接"功能验证

## 备份选项说明

### mysqldump常用选项
| 选项 | 说明 | 推荐值 |
|------|------|--------|
| single-transaction | 一致性备份（InnoDB） | true |
| routines | 备份存储过程 | true |
| triggers | 备份触发器 | true |
| events | 备份事件 | true |
| hex-blob | 二进制数据十六进制 | true |
| max-allowed-packet | 最大包大小 | 67108864 |

### mydumper常用选项
| 选项 | 说明 | 推荐值 |
|------|------|--------|
| threads | 并发线程数 | 4 |
| compress | 压缩输出 | true |
| chunk-filesize | 文件分块大小(MB) | 128 |
| rows | 每个文件的行数 | 1000000 |

### xtrabackup常用选项
| 选项 | 说明 | 推荐值 |
|------|------|--------|
| parallel | 并发线程数 | 4 |
| compress | 压缩备份 | true |
| compress-threads | 压缩线程数 | 4 |
| throttle | IO限速(IOPS) | 100 |

## 配置示例

### 小型数据库（< 1GB）
```json
{
  "single-transaction": true,
  "routines": true,
  "triggers": true
}
```

### 中型数据库（1-10GB）
```json
{
  "threads": 4,
  "compress": true,
  "chunk-filesize": 128
}
```

### 大型数据库（> 10GB）
```json
{
  "parallel": 4,
  "compress": true,
  "compress-threads": 4,
  "throttle": 100
}
```

## 下一步

1. 安装所需的备份工具
2. 在主机管理中添加MySQL主机并测试连接
3. 在任务管理中创建备份任务
4. 配置适当的备份选项
5. 测试执行备份任务
6. 查看备份日志确认成功

## 技术支持

如果遇到问题，请检查：
1. 备份日志（日志管理页面）
2. 后端服务日志
3. MySQL错误日志
