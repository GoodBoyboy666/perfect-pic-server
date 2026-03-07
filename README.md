# Perfect Pic

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Gin](https://img.shields.io/badge/Framework-Gin-00ADD8?style=flat)
![SQLite](https://img.shields.io/badge/Database-SQLite-003B57?style=flat&logo=sqlite)
![License](https://img.shields.io/badge/License-MIT-green?style=flat)

**Perfect Pic** 是一个基于 Go (Gin) 开发的高性能、轻量级且功能完备的图床（图片托管）服务。采用**前后端分离架构**，使用AI辅助开发，专为个人或小型团队设计，提供安全可靠的图片存储、管理和分发功能。

📘 [项目文档](https://perfect-pic.goodboyboy.top/)

## ✨ 核心特性

* **🚀 高性能架构**
  * **多数据库适配**: 开箱即用支持 **SQLite** (零配置)，并可无缝切换至 **MySQL** 或 **PostgreSQL** 以适应生产环境。
  * **多级缓存加速**: 结合 HTTP 静态资源缓存与服务端内存缓存策略，大幅降低数据库压力，提升响应速度。
  * **Redis 持久化支持**: 可选接入 **Redis**，用于限流状态、Auth 用户状态缓存与重置密码 Token 的跨实例持久化与共享。
  * **并发与稳定性**: 针对不同数据库自动调优连接池，支持高并发读写；配合优雅停机机制，保障业务处理不中断。

* **🛡️ 安全可靠**
  * **多维安全防御**: 内置 JWT 身份认证、动态 IP 限流 (Rate Limiting) 以及生产环境安全检查，有效抵御恶意攻击。
  * **深度文件校验**: 基于文件内容 (Magic Bytes) 而非后缀名识别真实文件类型，杜绝伪装文件上传风险。
  * **数据一致性**: 核心操作（如批量删除、配额扣减）采用原子事务处理，确保文件与数据库状态始终同步。

* **⚙️ 现代架构与易用性**
  * **前后端分离**: 基于标准 RESTful API 设计，逻辑解耦。同时支持将前端资源嵌入二进制文件，既享受分离开发的灵活性，又拥有“单文件部署”的便捷性。
  * **配置热重载**: 支持在线动态调整系统参数（如限流阈值、站点设置），无需重启服务。
  * **智能配额管理**: 采用增量更新策略，无论图片数量多少，都能快速计算用户剩余存储空间。
  * **规范化存储**: 自动按日期分目录存储文件，便于运维管理与备份。

## 🛠️ 技术栈

* **语言**: Go (Golang)
* **Web 框架**: [Gin](https://github.com/gin-gonic/gin)
* **ORM**: [GORM](https://gorm.io/)
* **数据库**: SQLite, MySQL, PostgreSQL
* **缓存/持久化**: Redis (可选) / Memory
* **配置管理**: [Viper](https://github.com/spf13/viper)
* **工具库**: UUID, Captcha, Lumberjack (日志)

## 🚀 快速开始

### 1. 下载

> [!TIP]
> 带有 `embed` 字样的构建默认嵌入前端资源，开箱即用。不带该字样的构建仅为后端服务，需要自行部署前端服务。

请前往 [Releases](https://github.com/GoodBoyboy666/PerfectPic-Server/releases) 页面下载适用于您操作系统的最新版本程序。

### 2. 运行

下载后，直接在终端或命令行中运行程序。为了安全起见，生产环境**强烈建议**设置 JWT 密钥。

**Linux / macOS:**

```bash

# 赋予执行权限
chmod +x perfect-pic-server

# 设置环境变量并启动
export PERFECT_PIC_SERVER_MODE=release
export PERFECT_PIC_JWT_SECRET=perfect_pic_secret
./perfect-pic-server
```

可选参数：

```bash
./perfect-pic-server --config-dir ./config
```

**Windows (PowerShell):**

```powershell
$env:PERFECT_PIC_SERVER_MODE="release"
$env:PERFECT_PIC_JWT_SECRET="perfect_pic_secret"
.\perfect-pic-server.exe
```

可选参数：

```powershell
.\perfect-pic-server.exe --config-dir .\config
```

服务启动后，默认运行在 `http://localhost:8080`。

### 3. 初始化

访问 `http://localhost:8080/init` 即可进入初始化向导。

## ✈️ Docker 部署

如果你更喜欢使用 Docker 部署，项目提供了开箱即用的 Docker 镜像以及 Dockerfile。

### docker run

先拉取镜像：

```bash
docker pull ghcr.io/goodboyboy666/perfect-pic-server:latest
```

运行容器并持久化数据：

```bash
docker run -d \
  --name perfect-pic \
  -p 8080:8080 \
  -e PERFECT_PIC_SERVER_MODE=release \
  -e PERFECT_PIC_JWT_SECRET=perfect_pic_secret \
  -v $PWD/config:/data/config \
  -v $PWD/database:/data/database \
  -v $PWD/uploads:/app/uploads \
  ghcr.io/goodboyboy666/perfect-pic-server:latest
```

* **挂载说明**:
  * `/data/config`: 存放配置文件和邮件模板。强烈建议首次运行前在该目录下配置好 `config.yaml`。
  * `/data/database`: 存放数据库文件（默认 SQLite 路径为 `/data/database/perfect_pic.db`）。
  * `/app/uploads`: 持久化存储上传的图片。

### docker compose

项目根目录已提供 `docker-compose.yml`，可直接使用：

```bash
# 复制环境变量模板（不可直接使用，必须按需修改）
cp .env.example .env

# 后台启动
docker compose up -d
```

如需停止并移除容器：

```bash
docker compose down
```

### 自行构建镜像

```bash
# 获取构建版本信息
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse HEAD)
DATE=$(date '+%Y-%m-%d_%H:%M:%S')

# 构建镜像
docker build . \
  -t perfect-pic-server:latest \
  --build-arg APP_VERSION="$VERSION" \
  --build-arg GIT_COMMIT="$COMMIT" \
  --build-arg BUILD_TIME="$DATE" \
  --build-arg FRONTEND_REF="origin/main"
```

构建完成后，可在 `docker run` 中把镜像名替换为 `perfect-pic-server:latest`；
如果使用 `docker compose`，请将 `docker-compose.yml` 中的 `image` 改为 `perfect-pic-server:latest`。

## 🛠️手动构建

如果您想从源码编译或参与开发：

### 1. 环境要求

* Go 1.25 或更高版本
* NodeJs 22 或更高版本
* PNPM 10 或更高版本
* MySQL/PostgreSQL (可选)

### 2. 获取代码

```bash
git clone https://github.com/GoodBoyboy666/perfect-pic-server.git

cd perfect-pic-server
```

### 3. 编译运行

```bash
# 进入脚本文件夹
cd scripts/

# 赋予执行权限
chmod +x build.sh

# 执行编译脚本
./build.sh
```

最终产物位于项目根目录的 `bin` 文件夹

### 4. 前后端分离部署（非 embed 模式）

项目前端仓库为：[perfect-pic-web](https://github.com/GoodBoyboy666/perfect-pic-web)

可以将前端与后端分离部署于不同的机器，只需将来自下列的路径的请求转发至后端即可：

* /api/*
* /imgs/*
* /avatars/*

可以使用Nginx或者Caddy的反向代理处理相关请求。

## ⚙️ 配置说明

项目支持 `config.yaml` 配置文件和环境变量双重配置。

程序默认使用 `config/` 目录，可通过启动参数 `--config-dir` 指定其它目录（例如 `--config-dir /data/config`）。

### 配置文件 (config.yaml)

首次运行会自动使用默认配置，你可以在根目录或 `config/` 目录下创建 `config.yaml`：

```yaml
server:
  port: "8080"
  mode: "release" # debug / release
  trusted_proxies: "" # 逗号分隔或 CIDR，留空表示不信任代理头

database:
  type: "sqlite" # sqlite, mysql, postgres
  filename: "database/perfect_pic.db" # for sqlite  
  host: "127.0.0.1" # for mysql/postgres
  port: "3306"
  user: "root"
  password: "password"
  name: "perfect_pic"
  ssl: false

jwt:
  secret: "perfect_pic_secret"
  expiration_hours: 24

upload:
  path: "uploads/imgs"
  url_prefix: "/imgs/"
  avatar_path: "uploads/avatars"
  avatar_url_prefix: "/avatars/"

smtp:
  host: "smtp.example.com"
  port: 587
  username: "examle@example.com"
  password: "your_smtp_password"
  from: "examle@example.com"
  ssl: false

redis:
  enabled: false # 是否启用 Redis 持久化
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
  prefix: "perfect_pic"
```

### 环境变量

所有配置均可通过环境变量覆盖，前缀为 `PERFECT_PIC_`，层级用 `_` 分隔。
例如：

* `server.port` -> `PERFECT_PIC_SERVER_PORT`
* `server.trusted_proxies` -> `PERFECT_PIC_SERVER_TRUSTED_PROXIES`
* `jwt.secret` -> `PERFECT_PIC_JWT_SECRET`
* `redis.enabled` -> `PERFECT_PIC_REDIS_ENABLED`
* `redis.addr` -> `PERFECT_PIC_REDIS_ADDR`
* `redis.password` -> `PERFECT_PIC_REDIS_PASSWORD`
* `redis.db` -> `PERFECT_PIC_REDIS_DB`
* `redis.prefix` -> `PERFECT_PIC_REDIS_PREFIX`

当 `redis.enabled=true` 且可连接时，IP 限流、中间件间隔限流、重置密码 token 会写入 Redis；Redis 不可用时自动降级为内存模式。

### 邮件模板

`example` 文件夹中有有文件模板，复制至 `config` 目录即可。

## 📂 目录结构

```text
.
├── config/             # 配置文件目录
├── example/            # 示例文件 (如邮件模板)
├── frontend/           # 前端静态资源 (嵌入式)
├── internal/
│   ├── common/         # 通用错误模型与 HTTP 错误写入
│   ├── config/         # 配置加载与管理
│   ├── consts/         # 常量定义
│   ├── db/             # 数据库初始化与迁移
│   ├── di/             # Wire 依赖注入装配
│   ├── dto/            # 请求/响应 DTO
│   ├── handler/        # HTTP Handler 层
│   ├── middleware/     # Gin 中间件
│   ├── model/          # 数据模型
│   ├── repository/     # 数据访问层
│   ├── router/         # 顶层路由编排
│   ├── service/        # 领域服务与基础能力
│   ├── usecase/        # 应用编排层
│   │   ├── app/        # 前台业务用例
│   │   └── admin/      # 后台管理用例
│   ├── testutils/      # 测试辅助
│   └── utils/          # 工具函数
├── scripts/            # 构建与部署脚本
├── embed_enabled.go    # embed 构建入口
├── embed_disabled.go   # 非 embed 构建入口
├── main.go             # 程序入口
└── go.mod
```

## 📝 API 概览（部分）

### 公开接口

* `GET /api/init`: 检查是否需要初始化系统
* `POST /api/init`: 初始化管理员账号
* `POST /api/login`: 用户登录
* `POST /api/register`: 用户注册
* `POST /api/auth/passkey/login/start`: 发起 Passkey 登录挑战
  * 返回字段：`session_id`、`assertion_options`
* `POST /api/auth/passkey/login/finish`: 完成 Passkey 登录
* `GET /api/captcha`: 获取验证码元信息（`provider` + `public_config`，当 provider 为空表示已关闭验证码）
* `GET /api/webinfo`: 获取站点公开信息

### 用户接口 (需 Auth)

* `POST /api/user/upload`: 上传图片
* `GET /api/user/images`: 获取我的图库
* `DELETE /api/user/images/batch`: 批量删除图片
* `GET /api/user/profile`: 获取个人信息
* `PATCH /api/user/avatar`: 更新头像
* `POST /api/user/passkeys/register/start`: 发起 Passkey 绑定挑战
  * 返回字段：`session_id`、`creation_options`
* `POST /api/user/passkeys/register/finish`: 完成 Passkey 绑定
* `GET /api/user/passkeys`: 获取当前用户已绑定 Passkey 列表
* `PATCH /api/user/passkeys/:id/name`: 更新当前用户指定 Passkey 的名称
* `DELETE /api/user/passkeys/:id`: 删除当前用户的指定 Passkey
* 约束：单个用户最多可绑定 10 个 Passkey

### 管理员接口 (需 Admin 权限)

* `GET /api/admin/stats`: 获取服务器统计
* `GET /api/admin/users`: 用户列表管理
* `PATCH /api/admin/settings`: 动态修改系统配置

## 🤝 贡献

欢迎提交 Issue 或 Pull Request 来改进这个项目！

## 📄 许可证

[MIT License](LICENSE)
