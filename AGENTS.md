# AGENTS.md

## 适用范围

本文件适用于整个仓库。

`docs/` 是架构规范的权威来源。本文件只保留代理在改代码前必须记住的核心规则，细节以 `docs/` 为准。

## 项目形态

- 本项目是 Go modular monolith，module path 为 `modular_monolith`，Go 版本为 `1.26.3`。
- 当前技术栈：Echo v5 负责 HTTP，GORM MySQL 负责持久化初始化，`slog` 负责结构化日志，`go-playground/validator` 负责请求校验。
- 生产代码一律放在 `internal/` 下，不新增 `pkg/`。
- `cmd/main.go` 必须保持极薄：加载配置、调用 `bootstrap.Run`、等待退出信号。
- 当前实现主要是 `platform` / `bootstrap` 脚手架；新增业务代码时必须遵循 `docs/architecture.md` 中的 BC 目录结构。

## 架构规则

- 每个 Bounded Context 使用 DDD Lite 分层：`domain/`、`app/`、`ports/`、`adapters/`，以及 `internal/{bc}/module.go`。
- 依赖方向只能向内：
  - `domain` 只 import 标准库和合法的 `internal/shared` 业务原语。
  - `app` 可以 import 自己的 `domain` 和 `shared`；禁止 import `ports`、`adapters`、`bootstrap` 或其他 BC。
  - `ports` 只负责把入站协议或事件翻译为 app command / query 调用。
  - `adapters` 只实现 `domain` 或 `app` 拥有的接口。
- `internal/{bc}/module.go` 是 BC-level composition root，定义 `Deps`、`Module`、`NewModule(Deps) (*Module, error)`。
  - 可以装配本 BC 的 app、ports、非跨 BC adapters。
  - 禁止打开 DB / Redis / HTTP server。
  - 禁止 import 任何其他 BC，包括对方的 `ports/module`。
  - 禁止构造跨 BC adapter；这些 adapter 由 bootstrap 构造后通过 `Deps` 注入。
- `internal/bootstrap` 是 process-level composition root：打开 platform 资源、实例化各 BC module、装配跨 BC adapter、挂载 router、注册事件订阅、管理 graceful shutdown。
  - bootstrap 可以 import platform 包、BC 根包、跨 BC adapter 包。
  - bootstrap 禁止直接 import BC 的 `domain`、`app`、`ports` 或持久化 adapter。
- `internal/platform` 只能放与业务无关的技术底座；`internal/shared` 只放至少两个 BC 语义完全一致的业务原语。

## 跨 BC 通信

只允许三种方式：

- 低频同步调用：走目标 BC 的 `ports/module` contract。
- 异步协作：走 `ports/module/events.go` 中带版本的 Integration Event。
- 高频跨 BC 读：优先维护本地物化视图。

禁止 import 其他 BC 的 `domain`、`app`、`adapters`。调用方在自己的 `app/command/services.go` 定义最小接口；类似 `adapters/shopclient` 的 adapter 调用目标 `ports/module`，并负责 DTO / error 翻译。

Module Contract 和 Integration Event 是稳定传输契约。不要把 domain entity 暴露到 BC 边界外。

## CQRS、事件与事务

- Command 放在 `app/command`，一个写用例一个文件，统一暴露 `Handle(ctx, cmd) error`。
- Query 放在 `app/query`，依赖 `ReadModel` 接口，返回 DTO，不返回 domain 对象。
- Domain Event 只在 BC 内部流转；Integration Event 是带版本的公开契约，由持久化 adapter 在保存后翻译并发布。
- 新代码优先使用 `PeekEvents()`，仅在 commit 且 publish 成功后 `ClearEvents()`；不要引入 `PullEvents()`。
- 事务语义错误属于 platform，例如 `internal/platform/txerr`，不要放在 domain 或 adapter 包。
- 不要通过 `context.Context` 传递 `*sql.Tx`。

## 命名

- 包名小写、单数、面向业务概念；不要创建 `utils`、`helpers`、`common`、`misc`、`tools`。
- 文件名使用 `snake_case`。
- Domain 类型使用业务名，不加 `Entity`、`VO`、`DTO` 后缀。
- `DTO` 后缀只用于 app/query 和 ports/module 的传输对象。
- 接口默认放在调用方；例外是 domain repository，放在 `domain/{aggregate}/repository.go`。
- Handler 方法统一命名为 `Handle`。

## 测试与验证

- 修改 Go 文件后运行 `gofmt`。
- 默认验证命令是 `go test ./...`。
- 集成测试使用 `//go:build integration`，通过 `go test -tags=integration ./...` 单独运行。
- E2E 测试使用 `//go:build e2e`，单独运行。
- 测试文件默认与被测代码同包同目录；跨 BC E2E 测试除外。
- Domain 测试必须是纯单测，不起 DB / HTTP，也不使用 mock。
- App 测试 fake / mock 接口，不 mock SQL；简单接口优先手写 fake。
- Ports 测试只验证协议翻译和对 app handler 的委托。
- 持久化 adapter 测试使用真实持久化，并放在 integration build tag 后。

## 文档与 ADR

- 不确定规则时按顺序阅读：`docs/architecture.md`、`docs/cross-context.md`、`docs/naming.md`、`docs/testing.md`、`docs/anti-patterns.md`、`docs/adr/`。
- 架构规则变更、新增跨 BC 通信方式、契约语义变更、新增核心基础设施、新增 `shared/` 或 `platform/` 组件、任何有意破例，都必须写 ADR。
- `docs/anti-patterns.md` 是 code review 的 blocking 检查清单。
