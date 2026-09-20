# Hephaestus 收入后端

Google Sheets 手动快照导入已接入 Apollo 统一授权：[配置与使用说明](docs/google-sheets.md)。真实 Google 使用前需完成 Cloud 凭据与回调配置。

目录保持 `hephaestus-api` / `hephaestus-rpc`，启动入口均为 `hephaestusrpc.go`。
独立 Go module 和数据库承载收入记录、汇总、CSV/XLSX 导入；Invoice 后端不在本次范围。

## 统一身份与权限

**所有登录、会话及子系统访问授权由 Apollo 管理。**

- 不再提供 `POST /api/v1/session` 密码登录，没有 `LocalAccount`、bcrypt 登录、
  本地账户初始化工具或本地会话表。
- `GET /api/v1/auth/start` 创建浏览器绑定的短期 HttpOnly 状态 Cookie，发起 Apollo
  S256 PKCE 授权。`GET /api/v1/auth/callback` 验证 state/签名/期限后，在服务端交换授权码。
- Apollo 保存一次性授权码和子系统会话。收入 API 只持有 Apollo 发出的不透明会话
  Cookie；收入 RPC 每次在线调用 Apollo `/v1/sso/introspect`，不复制账户密码或 JWT 密钥。
- Apollo 控制 client 启用状态、scope 4、授权撤销、会话期限和账户失效版本。
  API 与 RPC 均校验身份；业务仓储继续按该身份对应的 owner 隔离记录。
- Apollo 不可用时返回服务不可用；权限不足返回 403；会话无效返回 401。
  不会退回本地登录、信任前端 owner ID 或使用旧会话绕过 Apollo。
- 修改请求继续要求 CSRF；版本控制、幂等、来源限制和 RPC 服务凭据继续生效。

`owners` 现在只是收入业务所有者映射。`apollo_subject` 为唯一外部身份标识，
本地 `id` 仅用于外键和事务锁，不能登录或授予权限。迁移保留旧 owner ID 及业务数据，
旧 owner 的 `apollo_subject` 留空；不得自动按姓名、邮箱或数字 ID 认领旧记录。
如需绑定旧收入，须明确提供经核实的 Apollo subject 和对应旧 owner，单独迁移数据归属。

## 本地配置

各项目使用本地 MySQL **3306**，连接串直接读取各自 `etc/*.yaml`；不使用旧的
13316 Docker 开发库作为运行配置，不把收入和 Apollo 数据放入同一业务库。

- 收入 RPC：`hephaestus-rpc/etc/hephaestusrpc.yaml` 的 `DB.DataSource`、`Apollo`。
- 收入 API：`hephaestus-api/etc/hephaestus-api.yaml` 的 `Apollo`、`SSO`、`Browser`。
- Apollo DDD RPC：`apollo/apollo-rpc-ddd/etc/apollorpc.yaml` 的 `DB`、`SSO.Clients`。
- Apollo DDD API：`apollo/apollo-api-ddd/etc/apollo-api.yaml` 的 `SSO.RPCSecret`。

`Apollo.ClientSecret` 与 Apollo 注册的 Hephaestus client secret 必须一致；
`SSO.RPCSecret` 仅由 Apollo API/RPC 持有。当前浏览器 origin 为
`http://192.168.2.7:15173`，Apollo 授权页为 `http://192.168.2.7:20551/authorize`，
精确回跳为 `http://192.168.2.7:15173/api/v1/auth/callback`。
使用其他 hostname/HTTPS 时，需同时更新回跳注册、公开 origin、CORS、Secure Cookie
和 Apollo WebAuthn RP origins；不要混用 localhost 与 127.0.0.1 浏览器入口。

手动修改访问地址时，需要同时修改以下配置并重启四个服务：

- `hephaestus-api/etc/hephaestus-api.yaml`：`Browser.AllowedOrigins`、
  `SSO.AuthorizeURL`、`SSO.PublicURL`；
- `apollo-rpc-ddd/etc/apollorpc.yaml`：`SSO.Clients[].RedirectURIs`；
- `apollo-api-ddd/etc/apollo-api.yaml`：`FrontendURL`；
- 使用 OAuth 时，还要修改 `apollo-rpc-ddd/etc/apollorpc.yaml` 中各 provider 的
  `RedirectURL`，并在 GitHub/Google 控制台登记完全相同的回调地址。

Vite 的 `proxy.target` 是服务器内部 API 地址，可以继续使用 `127.0.0.1`；它不会成为
浏览器跳转地址。局域网 HTTP 地址只允许私有 IP。Passkey/WebAuthn 若要从其他设备
使用，应改用 HTTPS 域名，并同步设置 `WebAuthn.RPID` 与 `RPOrigins`。

## 迁移与启动

先确保各自配置的本地数据库已存在且连接账号可用。迁移时需要相应表的 DDL 权限；
迁移工具不会创建用户或把 Apollo 的数据库凭据复制给收入服务。

在 `apollo/apollo-rpc-ddd` 执行：

```sh
go run ./cmd/sso-migrate
```

在 `hephaestus` 执行：

```sh
bash scripts/local-db.sh
bash scripts/build.sh
./bin/hephaestus-rpc -f hephaestus-rpc/etc/hephaestusrpc.yaml
# 另一个终端：
./bin/hephaestus-api -f hephaestus-api/etc/hephaestus-api.yaml
```

将初步设计内嵌的 19 条 AimerHQ 记录绑定到已核实的 Apollo 账户时，可执行：

```sh
mysql --default-character-set=utf8mb4 -h127.0.0.1 -P3306 \
  -u<hephaestus-user> -p hephaestus < scripts/seed-initial-design.sql
```

脚本按 Apollo subject 查找 owner，并以来源和稳定 `source_key` 幂等导入；不会按显示姓名
推断身份，也不会覆盖之后在系统内编辑过的记录。当前参考数据汇总为 19 条、税前
NZD 1,021.00、费用 NZD 254.66。

`schema-migrate` 会初始化空库或将 v1/v2/v3 升至 v4，可重跑；v2 删除旧会话和密码字段，v3 新增来源软删除标记，v4 允许相同内容新建同步批次并保留历史，
保留收入业务表。正常启动只检查 schema，不隐式迁移。Apollo 的新增 SSO 表支持
账户/授权删除时清理关联会话。迁移后需重启旧版本 API/RPC，不能继续运行密码登录版二进制。

AimerHQ 工作表导入先阅读 [工资与只读规则](docs/aimerhq-rules.md)。映射页选择规则预设后，必须明确工资列并确认，才可预览。来源删除只隐藏本地来源、停止后续导入，保留工作记录与历史，不修改 Google 文档。

默认端口：Apollo API/RPC 21100/31100，收入 API/RPC 18101/19101，
Apollo/Hephaestus 前端 20551/15173。

## 验证与生成

```sh
go test -race ./...
go vet ./...
bash scripts/generate.sh
```

协议修改遵循先 `.api/.proto`、实际 goctl 生成、再实现业务。新授权端点也遵循此流程。
生成命令见 `scripts/generate.sh`，与 `/home/jian/.zshrc` 的 GoZero alias 一致。

`python3 scripts/smoke_test.py --bin-dir /tmp` 是可复现的隔离集成测试：需要
`/tmp/apollo-sso-api`、`/tmp/apollo-sso-rpc`、`/tmp/heph-sso-api`、`/tmp/heph-sso-rpc`
四个当前构建及 Docker。它仅为测试创建临时 MySQL 容器和两个临时数据库，
使用 13326/22100/32100/18201/19201，不修改任何运行 YAML 或本地 3306 业务数据。
默认结束后清理；`--keep-running` 供浏览器验收，退出时清理。
验证包含真实 Apollo 注册/登录、一次性回跳、跨账户隔离、旧数据迁移、CSRF、
导入确认、幂等/并发版本、重启持久性、子系统退出和 Apollo 全局退出。
