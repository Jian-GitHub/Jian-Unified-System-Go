# Apollo RPC

Apollo 的完整账户与身份业务实现：Account 14、Passkeys 6、Security 4、ThirdParty 6，共 30 个 RPC 方法。领域规则和事务边界见 [重构说明](../../docs/apollo-refactoring.md)。原工程保留，未切换流量。

## 开发流程

先修改 [proto/apollo-rpc.proto](proto/apollo-rpc.proto)，再在 `proto/` 执行 `gorpcgn`。非交互 shell 等价命令：

```bash
goctl rpc protoc *.proto --go_out=../ --go-grpc_out=../ --zrpc_out=../ --style=goZero -m -I. -I/usr/include
```

消息、四组 client/server 由生成器维护；生成的 logic 仅转换消息、调用 Accounts/Identity 并映射错误。domain 为纯 Go 规则，application 通过端口协调基础设施。生成后在模板内实现，不复制旧目录内容。

## 开发运行

[etc/apollorpc.yaml](etc/apollorpc.yaml) 直接填写开发配置：DB、MLKEMKey 密钥对、WebAuthn、OAuth、SubSystem。RPC 默认回环 31100，MySQL 回环 13306 的独立 apollo_sandbox 库，无旧服务发现注册。

在独立库依次执行 [schema/account.sql](schema/account.sql)、[schema/identity.sql](schema/identity.sql)，再运行：

```bash
go run -mod=readonly . -f etc/apollorpc.yaml
```

程序启动不自动建表。此 DDL 是新工程的开发结构，不能直接对旧库执行；旧 credential 数据、provider 唯一性、邮箱空值和到期字段需要另行迁移核对。

2026-09-08 以前创建的 DDD 开发库可能缺少 `auth_version`、`login_email`、联系方式组合唯一索引或 `event_outbox`，导致登录报 MySQL 1054 或使事务能力与当前代码不一致。已存在的开发库不要重新执行建表脚本；先备份，再使用可重复执行的 [账户增量升级 SQL](schema/upgrade-account-20260908.sql)。既有主邮箱展示副本必须在解密候选与登录查询索引匹配后回填，不能直接复制通知邮箱。

RPC 启动会只读校验必需表、字段和关键唯一索引；结构不兼容时直接拒绝启动并列出缺项，不会等到登录请求再返回笼统 Internal。该校验不会自动执行 DDL 或回填数据。

通知邮箱查询使用配置中的私钥解密。账户与子系统 JWT 使用不同开发密钥。OAuth 的官方端点和回调已填写，ClientID/ClientSecret 采用明确的本地占位值；真实登录需填入注册应用的实际开发凭据。WebAuthn RP 为 localhost，访问 origin 必须匹配 YAML。会话保存在 MySQL 并原子消费，无 Redis 依赖。

RPC 按可信内部调用边界处理显式 user_id，未新增 RPC 级用户鉴权；当前只监听回环。完成重构不表示已授权对外暴露或接入旧 JQuantum。

## 验证

仓库根执行：

```bash
go test -mod=readonly -race ./apollo/apollo-rpc/...
```

显式真实库检查要求空的回环 apollo_test 库，直接读取 [测试 YAML](etc/apollorpc-test.yaml)：

```bash
go test -mod=readonly -race -count=1 ./apollo/apollo-rpc/internal/infrastructure/persistence -mysql-test-config=../../../etc/apollorpc-test.yaml
```

覆盖账户唯一性/密码持久化、重复凭据回滚、会话并发单次消费、凭据版本冲突、最后凭据保护和并发 OAuth 身份解析。默认不传参数时跳过真实库检查。全部 HTTP/RPC 双进程联调入口见 [API README](../apollo-api/README.md)，证据见 [测试记录](../../docs/quality.md)。
