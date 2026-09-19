# Apollo API

Apollo 的完整 HTTP 接入实现，共 30 个路由：账户资料与安全设置、Passkey、子系统令牌和 GitHub/Google 身份。新目录独立运行，未替换原有服务。实现和兼容差异见 [重构说明](../../docs/apollo-refactoring.md)。

## 开发流程

先修改 [api/apollo.api](api/apollo.api)，再在 `api/` 执行 `goapign`。非交互 shell 使用核对后的等价命令：

```bash
goctl api go -api *.api -dir ../ --style=goZero
```

routes/types 由生成器维护，其他可编辑模板内实施业务；不复制旧框架或业务代码。API logic 调用 Sessions 或四组生成 RPC client，领域规则在 RPC；handler 负责请求、响应和 OAuth 浏览器关联，middleware 补充严格 JWT 校验。

## 开发运行

先准备并启动新 RPC，然后在本目录执行：

```bash
go run -mod=readonly . -f etc/apollo-api.yaml
```

[开发 YAML](etc/apollo-api.yaml) 直接读取，默认 API 回环 21100、RPC 回环 31100。Auth 开发密钥已配置；多个 API 实例使用不同 Snowflake.NodeID。`GeoIP.Database` 指向 GeoLite2 数据库，邮箱、Passkey 和第三方 OAuth 的所有新账户按客户端 IP 保存国家代码，无法确认国家时保存 `UNKNOWN`；`DefaultLanguage` 提供无需客户端语言输入流程的默认值。WebAuthn 使用 localhost 域名，与 RPC 的 RPOrigins 一致。镜像应从仓库根以 `docker build -f apollo/apollo-api/deploy/Dockerfile .` 构建，以包含 GeoLite2 数据文件。

OAuth 起始响应设置短期关联 Cookie，回调要求同一浏览器携带它；前端 fetch 需使用 credentials: include。优先同源代理，跨源 CORS 由实际网关配置。成功回调 302 至 FrontendURL，token 位于查询参数。真实 GitHub/Google 登录需在新 RPC YAML 中填写注册应用的 ClientID/ClientSecret；当前明确的本地占位值不能用于真实提供者登录。

Turnstile 使用 [官方开发测试配置](https://developers.cloudflare.com/turnstile/troubleshooting/testing/)，hostname/action 精确匹配 dummy 响应；开发联调不代表真实人机验证验收。

## 测试

仓库根执行离线检查：

```bash
go test -mod=readonly -race ./apollo/apollo-api/... ./apollo/apollo-rpc/...
go vet -mod=readonly ./apollo/apollo-api/... ./apollo/apollo-rpc/...
```

完整双进程联调需要回环 13306 上独立、空的 apollo_test MySQL 库，以及能调用 Cloudflare 官方 dummy 验证端点的网络。数据库连接直接读取 RPC 测试 YAML，拒绝覆盖已有表：

```bash
go test -mod=readonly -race -count=1 -v ./apollo/apollo-api/integration -integration-config=../../apollo-rpc/etc/apollorpc-test.yaml
```

路径相对于 integration 包。测试在临时目录构建并启动两个带竞态检测的真实服务进程，实际经过生成路由/JWT 中间件/客户端/server、真实 MySQL 和 WebAuthn ES256 签名；OAuth 提供者使用校验 PKCE 的本地 fixture。测试清理进程和表，MySQL 实例由调用方管理。默认不传参数时跳过集成检查；详情见 [测试记录](../../docs/quality.md)。
