# ADR-0007: BC 自带 Module 工厂,bootstrap 只做进程级编排

- **状态**:Accepted
- **日期**:2026-05-20
- **决策者**:架构组

## 背景

最初版本的设计把每个 BC 的装配代码集中放在 `internal/bootstrap/{bc}_module.go`,bootstrap 直接 import 各 BC 的 domain / app / adapters / ports 内部包,负责实例化所有组件。

这种设计有几个问题:

1. **`bootstrap` 包变成"上帝包"**,知道每个 BC 的所有内部包名;BC 内部目录调整都要去 bootstrap 改
2. **BC 不自给自足**:把 `internal/order/` 整个目录搬到另一个仓库,跑不起来——装配代码留在原仓库的 bootstrap 里
3. **BC 单独测试困难**:测试想拉起一个完整 BC 必须复制 bootstrap 里的装配逻辑
4. **bootstrap 体量随 BC 数量线性增长**,且不同 BC 的装配代码物理上相邻容易耦合
5. **跨 BC 依赖关系不显式**:在 bootstrap 里看 order 的装配代码,要从 `command.PlaceOrderHandler{Orders: ..., Products: ...}` 反推 order 依赖 shop

需要把"BC 装配"从 bootstrap 中剥离。

## 决策

采用**两层 Composition Root**:

### 1. BC-level Composition Root:`internal/{bc}/module.go`

每个 BC 在自己根目录下持有 `module.go`,包名为 BC 名(如 `package order`),定义:

- `Deps` 结构体:**显式声明** BC 装配所需的外部依赖,其中跨 BC 依赖**只能用本 BC 自己定义的小接口**(如 `command.ProductsService`),禁止直接持有对方 `ports/module` 类型
- `Module` 结构体:**显式暴露** BC 对外的接入点(HTTP router、`ports/module` 实例、event subscriptions 等)
- `NewModule(Deps) (*Module, error)` 工厂函数:装配本 BC 所有内部部件;**不**新建跨 BC adapter(那由 bootstrap 实例化)

### 2. Process-level Composition Root:`internal/bootstrap/server.go`

只做"进程级编排":

- 加载 config,打开技术底座(logger / DB / Redis / eventbus / HTTP server)
- 决定本进程运行哪些 BC
- **按依赖顺序**实例化各 BC 的 `Module`
- **实例化跨 BC adapter**(`{bc}/adapters/{other_bc}client/`),把其包装成本 BC 接口类型后通过 `Deps` 注入下游 BC——跨 BC wiring 集中在这里
- 把各 BC 暴露的 `HTTPRouter` 挂到 HTTP server,把 `EventSubs` 注册到 eventbus
- 启动 + graceful shutdown

### 强制约束

- `bootstrap` **允许** import 的范围:各 BC 根包(`order` / `shop` / ...)、各 BC 的 `adapters/{other_bc}client/`、`platform/`
- `bootstrap` **禁止** import BC 的其他内部包(domain / app / ports / 其他 adapter)
- `{bc}/module.go` **禁止** import **任何**其他 BC 的包,**包括对方的 `ports/module/`**(详细原因见下方"为什么 module.go 不能 import 对方 ports/module")
- `{bc}/module.go` **禁止** import 自己 BC 的 `adapters/{other_bc}client/`(那由 bootstrap 实例化后通过 `Deps` 注入)
- 跨 BC 依赖**必须**出现在 `Deps` 字段中,字段类型只能是本 BC 自己定义的接口(`command.ProductsService` 等)
- `{bc}/module.go` **禁止**打开 DB / Redis / HTTP server(那是 bootstrap 的责任)

#### 为什么 `module.go` 不能 import 对方 `ports/module/`

Go `internal/` 包边界规则:`a/internal/b/c` 只能被 `a/` 子树下的包 import。当本 BC 被剥离到独立仓库后,跨仓库的 `internal/.../ports/module/` 在物理上将不再可 import——这一行会让"剥离零变更"的承诺直接破产。

把跨 BC 依赖**反转为本 BC 自定义接口**(`command.ProductsService`),`module.go` 物理上不依赖任何其他 BC,才能真正剥离零变更。对方 `ports/module` 的 import 只下沉到 `adapters/{other_bc}client/` 这一层——这一层本就是跨 BC 胶水,拆服务时会被远端 client 替换,代价合理。

## 替代方案

### 方案 A:全部集中在 `internal/bootstrap/`(初版方案)

- **优点**:一处看到全部装配,顶层视角强
- **缺点**:已在背景中列出五条;尤其是"BC 不自给自足"和"bootstrap 上帝包"
- **放弃原因**:违反模块化原则——既然 BC 是模块,装配就应归模块自己

### 方案 B:每个 BC 开 `internal/{bc}/bootstrap/` 子目录

- **优点**:与顶层 `bootstrap/` 概念呼应
- **缺点**:子目录嵌套加深;且容易与 `internal/bootstrap/` 在 grep / IDE 跳转时混淆
- **放弃原因**:不开子目录、直接 BC 根放 `module.go` 更轻量

### 方案 C:每个 BC 开 `internal/{bc}/module/` 子目录

- **优点**:与现有 `ports/`、`adapters/` 子目录风格一致
- **缺点**:`internal/{bc}/module/` 与 `internal/{bc}/ports/module/` 概念冲突——一个是"BC 内部装配",一个是"BC 对外契约"
- **放弃原因**:命名冲突会持续误导新人

### 方案 D:用 DI 框架(wire / fx)生成装配代码

- **优点**:零模板代码,依赖图自动推导
- **缺点**:增加学习曲线;生成代码可读性下降;wire 在跨 BC wiring 的可读性反而不如手写
- **放弃原因**:本期不上 DI 框架。如未来引入,`{bc}/module.go` 内部可以改用 wire 实现 `NewModule`,接口签名不变,迁移平滑

## 后果

### 正向影响

- BC 模块完整内聚:装配 + domain + app + ports + adapters 都在 `internal/{bc}/` 下
- 拆服务零迁移:`internal/{bc}/` 整体搬走 + 新写 `cmd/{bc}-service/main.go` 即可
- BC 可独立测试:测试直接 `order.NewModule(testDeps)`,跨 BC 依赖用 fake
- 跨 BC 依赖显式可见:`grep "Deps{"` 在 bootstrap 一眼看清依赖图
- bootstrap 不随 BC 数量爆炸,新增 BC 只加 3-5 行代码

### 负向影响 / 代价

- 每个 BC 多一个 `module.go` 文件(但减少了 bootstrap 的对应文件)
- 跨 BC wiring 需要按依赖顺序在 bootstrap 中安排(有循环依赖会暴露在装配阶段——这其实是好事)
- BC 之间的初始化顺序是显式的,不能像 DI 框架那样自动推导

### Deps / Module 的设计准则

`Deps` 字段按"是否需要可替换性"分四类,**不要求统一全部用接口**:

| 类别 | 类型形态 | 例子 | 理由 |
|---|---|---|---|
| 基础设施技术句柄 | **具体类型** | `*sql.DB`、`*redis.Client`、`*http.Client` | 标准库/官方驱动的标准用法;由 platform 持有,生命周期归 bootstrap;BC 无需替换 |
| 横切能力 | **接口** | `Clock`、`IDGen`、`Tracer`、`*slog.Logger`(已是结构化句柄) | 测试需要替换;BC 之间共享语义但实现可不同 |
| 平台抽象 | **接口** | `eventbus.Bus` | monolith → microservices 演进时实现切换(in-memory ↔ Kafka/NATS) |
| **跨 BC 依赖** | **本 BC 自己定义的接口**(`command.ProductsService` 等) | 见 [`../architecture.md`](../architecture.md) §4.2.1 | 拆服务时实现切换为远端 RPC;**绝不允许直接持有对方 `ports/module` 类型** |

#### 反例

```go
// 反例:跨 BC 依赖直接用对方类型
type Deps struct {
    ShopModule shopmod.ShopModule // 拆服务时 import 失败
}
```

#### 正例

```go
// internal/order/app/command/services.go
type ProductsService interface {
    GetProduct(ctx context.Context, id string) (ProductInfo, error)
}

// internal/order/module.go
type Deps struct {
    DB       *sql.DB              // 基础设施句柄,具体类型 OK
    EventBus eventbus.Bus         // 平台抽象,接口
    Logger   *slog.Logger         // 已是结构化句柄
    Clock    clock.Clock          // 横切接口
    IDGen    idgen.IDGen          // 横切接口
    Products command.ProductsService // 跨 BC 依赖,本 BC 自定义接口
}
```

#### 其它规则

- `Module` 字段都是**已组装好的实例**,只有 bootstrap 关心
- `NewModule` 返回 `error` 而不是 panic,允许 bootstrap 决定失败策略
- **`module.go` 不允许 import 任何其他 BC 的包**(包括对方的 `ports/module/`)——这是"剥离零变更"的物理保证,违反将在 Go `internal/` 边界规则下让跨仓库迁移失败

### 演进路径

#### 同步调用

| 阶段 | `order.Deps.Products` 的实现 |
|---|---|
| Monolith 单进程 | `order/adapters/shopclient/` 包装 `shopmod.ShopModule`,bootstrap 实例化并注入 |
| Monolith 多进程 | bootstrap 注入 `order/adapters/shophttp/`(新增)的远端 client |
| Microservices | `internal/order/` 搬到独立仓库,**不带 `adapters/shopclient/`**;新 `cmd/order-service/main.go` 装配 order,使用 `adapters/shophttp/` |

#### 异步事件投递

| 阶段 | `platform/eventbus` 实现 |
|---|---|
| Monolith 单进程 | in-memory pub/sub;单聚合 `Save` / `Update{Aggregate}` commit 后 best-effort publish,UoW commit 后统一 flush pending events;可靠投递需 Outbox |
| Monolith 多进程 | broker client(Kafka / NATS / RabbitMQ);可引入 Outbox 保证投递可靠性 |
| Microservices | 同上 |

**关键**:`platform/eventbus.Bus` 接口签名在所有阶段不变,只换实现。

#### 始终不变(零变更)

- `internal/{bc}/domain/`
- `internal/{bc}/app/`(含 `command/services.go` 接口签名)
- `internal/{bc}/ports/{http,ws,stream,module}/`
- `internal/{bc}/module.go`(`Deps`、`Module`、`NewModule` 签名)
- `internal/{bc}/adapters/mysql/`、`adapters/redis/` 等仅依赖技术底座的 adapter
- `internal/{bc}/adapters/{vendor}/`(如 stripe)

#### 逻辑不变、import path 需切换

- `internal/{bc}/ports/eventsub/`:handler 逻辑不变,但 monolith 阶段 import 对方 `internal/.../ports/module/events.go` 的事件 DTO,拆服务后需改为从独立契约包或代码生成的 DTO 包 import

#### 必然变化(`adapters/{other_bc}client/` 与 `adapters/{other_bc}http/` 替换)

跨 BC adapter 是 hexagonal 模式中"承担 transport 变化"的层——本就该被替换。

### 后续需要做的事

- [ ] 在 `import-boundary` 规则中加入:
  - bootstrap 允许 import 各 BC 根包 + 各 BC 的 `adapters/{other_bc}client/`
  - bootstrap 不允许 import BC 的其他内部包
  - `{bc}/module.go` 不允许 import 任何其他 BC 的包(包括对方 `ports/module/`)
- [ ] BC 脚手架默认生成 `module.go` 三件套(Deps / Module / NewModule)
- [ ] 提供 `bootstrap.StartForTest(...)` 测试辅助,内部仍走标准 module 工厂
- [ ] `platform/eventbus.Bus` 接口预留 broker 实现的扩展点(envelope 含 event_id、headers)

## 相关文档

- [`../architecture.md`](../architecture.md) §4.2
- [`../anti-patterns.md`](../anti-patterns.md) §9
- [`../naming.md`](../naming.md) §1.2、§1.3
- [`../testing.md`](../testing.md) §4.5
- [ADR-0001 分层](./0001-layering.md)
- [ADR-0002 Module Contract](./0002-module-contract.md)
