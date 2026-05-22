# 反例集合

本文档收集**架构层面的反例**,与之对应的正例与原因。

**这是 code review 的检查清单**。表中任何一条出现在 PR 中,都应被标记为 blocking,直到修正或写 ADR 解释为什么本次破例。

> 关联文档:[`architecture.md`](./architecture.md)、[`cross-context.md`](./cross-context.md)、[`naming.md`](./naming.md)

## 1. 分层与依赖方向

### 1.1 domain 引用了外层

**反例**

```go
// internal/order/domain/order/order.go
import (
    "database/sql"                            // 反例
    "myshop/internal/order/adapters/mysql"    // 反例
    "myshop/internal/shop/domain/product"     // 反例
)
```

**正例**:domain 只依赖标准库 + `shared/`。持久化通过 `repository.go` 接口反转。跨 BC 的产品概念通过 `app/command/services.go` 接口反转。

**原因**:打破"内层不知外层",domain 不再可独立测试,也无法跨持久化技术迁移。

---

### 1.2 app 直接引用 adapter / ports

**反例**

```go
// internal/order/app/command/place_order.go
import (
    "myshop/internal/order/adapters/mysql"  // 反例
)

func (h *Handler) Handle(ctx context.Context, cmd PlaceOrder) error {
    repo := mysql.NewOrderRepository(db)  // 反例:app 不应知道 mysql
    ...
}
```

**正例**:app Handler 通过构造函数接收 `order.Repository` 接口,具体实现在 bootstrap 注入。

---

### 1.3 ports 跨 BC 直接调对方 app

**反例**

```go
// internal/order/ports/http/order_handler.go
import shopapp "myshop/internal/shop/app"  // 反例
```

**正例**:order 的 HTTP handler 只调自己 BC 的 `app.Application`。需要 shop 的能力时,由 order/app/command 通过 services.go 接口请求,adapter 走 `shop/ports/module/`。

## 2. 跨 BC 通信

### 2.1 跨 BC 直接 import 对方 domain / app

**反例**

```go
// internal/order/app/command/place_order.go
import "myshop/internal/shop/domain/product"  // 反例

func (h *Handler) Handle(ctx, cmd PlaceOrder) error {
    var p product.Product
    ...
}
```

**正例**:order/app 只依赖自定义的 `ProductInfo` DTO(来自 services.go),`order/adapters/shopclient/` 负责调 `shop/ports/module/` 并转换 DTO。

---

### 2.2 跨 BC adapter 直接 import 对方 adapter / domain

**反例**

```go
// internal/order/adapters/shopclient/products_service.go
import (
    "myshop/internal/shop/adapters/mysql"     // 反例
    "myshop/internal/shop/domain/product"     // 反例
)
```

**正例**:**只允许** import `myshop/internal/shop/ports/module`。

---

### 2.3 跨 BC adapter 命名歧义

**反例**

```text
order/adapters/shop/products_service.go
```

容易被误读为 "shop BC 内嵌在 order 下"。

**正例**

```text
order/adapters/shopclient/products_service.go
```

明确这是"shop 的客户端适配器"。

---

### 2.4 Module Contract 暴露 domain 对象

**反例**

```go
// internal/shop/ports/module/contract.go
import "myshop/internal/shop/domain/product"

type ShopModule interface {
    GetProduct(ctx, id string) (*product.Product, error)  // 反例:返回 domain entity
}
```

**正例**:返回 `ProductDTO`(在 `ports/module/types.go` 定义),domain entity 不出 BC 边界。

---

### 2.5 Integration Event 没有版本

**反例**

```go
// internal/order/ports/module/events.go
type OrderPlaced struct { ... }  // 反例:无版本
```

**正例**:`OrderPlacedV1`,事件契约的版本是契约的一部分。

---

### 2.6 用 Domain Event 直接做跨 BC 通信

**反例**:把 `domain/order/events.go` 中的 `OrderPlaced`(含聚合内部字段)直接通过 eventbus 跨 BC 发布。

**正例**:domain event 是 BC 内部产物;跨 BC 用 `ports/module/events.go` 中的 `OrderPlacedV1`,字段精简、稳定。**翻译在 repository 的 `Save` / `Update{Aggregate}` 内部完成**(见 §7.4),app 不感知 Integration Event。

---

### 2.7 Module Contract 用 HTTP 自调自

**反例**:`shop/ports/module/contract.go` 的实现里 `http.Get("http://localhost:8080/shop/products/..")`。

**正例**:Module Contract 是**进程内调用**,直接调 `shop/app.Application`,零网络。等真的拆服务时再换 adapter 走 HTTP。

## 3. 共享与技术底座

### 3.1 `shared/` 放接口

**反例**

```go
// internal/shared/userservice.go
type UserService interface { ... }  // 反例
```

**正例**:接口是行为契约,放在使用方。shared 只放**值类型 + 纯函数**。

---

### 3.2 `shared/` 出现 `utils.go` / `helpers.go` / `common.go`

**反例**:`shared/utils.go`、`shared/helpers/string.go`。

**正例**:按业务概念命名。如果想不出业务概念,说明这段代码本来就不该 shared。

---

### 3.3 `shared/` 单 BC 使用

**反例**:某个值类型只在 `payments` BC 使用,却放在 `shared/`。

**正例**:进入 shared 的条件:**≥ 2 个 BC 都使用**且**语义完全一致**。否则各 BC 自己定义,即便代码看起来一样。

---

### 3.4 `platform/` 出现业务概念

**反例**:`platform/order_id_generator.go`、`platform/payment_helper.go`。

**正例**:platform 只放与业务无关的技术底座(config / mysql / redis / eventbus / httpserver / logging)。业务相关的工具放对应 BC。

## 4. domain 设计

### 4.1 Aggregate 字段公开

**反例**

```go
type Order struct {
    ID     string
    Status Status
    Items  []OrderItem
}
```

任何代码都能 `o.Status = "Canceled"`,绕过业务规则。

**正例**:字段全部小写,通过业务方法访问与修改。

```go
type Order struct {
    id     OrderID
    status Status
    items  []OrderItem
}

func (o *Order) ID() OrderID { return o.id }
func (o *Order) Cancel() error { ... } // 业务方法
```

---

### 4.2 Aggregate 直接持有其它聚合实例

**反例**

```go
type Order struct {
    user *user.User  // 反例:聚合 A 持有聚合 B
}
```

**正例**:聚合之间**只通过 ID 引用**。需要 user 信息时,query 侧通过 ReadModel 组合,command 侧通过 services.go 取所需字段。

---

### 4.3 `NewOrder` 接收数据库主键

**反例**

```go
func NewOrder(id string, ...) (*Order, error)  // 反例:接收已存在的 id,语义混乱
```

**正例**:

- `NewOrder(...)` 在内部生成 ID,做全量校验,创建新聚合
- `UnmarshalOrderFromDatabase(...)` 接收 id + 所有字段,仅恢复状态,**只允许 adapters/mysql 调用**

---

### 4.4 业务方法名是数据语义

**反例**:`order.SetStatus(StatusCanceled)`。

**正例**:`order.Cancel(reason string) error`。

业务动词表达**为什么**改、是否被允许、产生什么事件。`SetStatus` 只表达**做了什么**,把规则推给调用方。

---

### 4.5 typed error 在 adapter 之外泄漏

**反例**

```go
// internal/order/app/command/place_order.go
if errors.Is(err, gorm.ErrRecordNotFound) { ... }  // 反例
```

**正例**:adapter 必须把 `gorm.ErrRecordNotFound` 翻译为 `order.ErrOrderNotFound`,app 层只感知领域错误。

## 5. CQRS

### 5.1 Query 返回 domain 对象

**反例**

```go
func (h GetOrderHandler) Handle(ctx, q GetOrder) (*order.Order, error)  // 反例
```

**正例**:返回 `OrderDTO`(`app/query/types.go` 定义)。Query 与聚合解耦,可以独立优化(读库、join、cache)。

---

### 5.2 Query 走 domain Repository

**反例**

```go
type GetOrderHandler struct {
    orders order.Repository  // 反例:query 用了 domain Repository
}
```

**正例**:query 用 `OrderReadModel` 接口(`app/query/read_model.go`),实现在 `adapters/{persistence}/order_read_model.go`。

---

### 5.3 Command 返回查询结果

**反例**

```go
func (h PlaceOrderHandler) Handle(ctx, cmd PlaceOrder) (*OrderDTO, error)  // 反例
```

**正例**:command 只返回 error(必要时返回新建实体的 ID)。下游需要详细数据时再走 query。

```go
func (h PlaceOrderHandler) Handle(ctx, cmd PlaceOrder) (OrderID, error)
```

## 6. ports 与 adapters

### 6.1 HTTP handler 内有业务逻辑

**反例**

```go
func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
    ...
    if total > 10000 {  // 反例:业务规则
        // 折扣
    }
    ...
}
```

**正例**:handler 只做 parse → call app → respond。业务规则在 domain / app。

---

### 6.2 HTTP handler 直接调 repository

**反例**

```go
func (h *OrderHandler) GetOrder(...) {
    o, _ := h.repo.FindByID(...)  // 反例
    ...
}
```

**正例**:handler 调 `h.app.Queries.GetOrder.Handle(...)`,query handler 再调 ReadModel。

---

### 6.3 Repository 接口在 adapter 包内

**反例**

```go
// internal/order/adapters/mysql/repository.go
type Repository interface { ... }  // 反例:接口定义在实现层
```

**正例**:`Repository` 接口定义在 `domain/order/repository.go`,adapter 实现它。

---

### 6.4 一个 repository 跨多个聚合

**反例**:`order_repository.go` 同时操作 Order 和 Product 两个聚合。

**正例**:Repository 与聚合一一对应。跨聚合协作在 app 层用多个 repository + UnitOfWork。

## 7. 事务与事件

### 7.1 事务横跨多个 BC

**反例**:一个 SQL 事务同时改 `order` 库和 `payment` 库。

**正例**:**绝不跨 BC 事务**。用 Integration Event 实现最终一致性。

---

### 7.2 订阅方不幂等

**反例**:`payments` 收到 `OrderPlacedV1` 直接创建支付,不做去重;重复投递会产生多条支付记录。

**正例**:任何 integration event 订阅方必须幂等。**本期**由 app/command 的业务唯一键保证(如 `InitializePayment` 以 `order_id` 唯一约束);`ports/eventsub/` handler 仅 decode → 委托 app。详见 [`cross-context.md`](./cross-context.md) §3.5。

---

### 7.3 app/command 缺少业务幂等键

**反例**:`InitializePayment` 每次被调用都 INSERT 新支付,无 `order_id` 唯一约束或状态检查。

**正例**:每个 eventsub 订阅链路的 app command 必须能以**稳定业务唯一键**防重复(唯一约束、upsert、显式状态检查)。code review 须能指出具体键与实现。

**演进注意**:接入外部 broker 后,可额外引入 `processed_events` 表以 `event_id` dedup,与业务幂等形成双层防线;届时业务幂等仍是必要条件,不能仅靠 `event_id` dedup 替代。

---

### 7.4 在 app/command 内翻译 Integration Event

**反例**

```go
// internal/order/app/command/place_order.go
import ordermod "myshop/internal/order/ports/module" // 反例:app 不能 import ports

func (h PlaceOrderHandler) Handle(ctx, cmd PlaceOrder) error {
    o, _ := order.NewOrder(...)
    h.orders.Save(ctx, o)
    h.eventbus.Publish(ordermod.OrderPlacedV1{ ... }) // 反例
    ...
}
```

**正例**:**Integration Event 翻译与发布在 repository 的 `Save` / `Update{Aggregate}` 内部完成**——repository 是 adapter,合法 import 自己 BC 的 `domain` + `ports/module/translate.go`。app/command 只调 `repo.Save(ctx, aggregate)` 或 `repo.Update{Aggregate}(ctx, id, actor, updateFn)`,完全不感知 Integration Event。

详见 [`cross-context.md`](./cross-context.md) §3.2。

---

### 7.5 commit 前 drain domain events

**反例**

```go
events := aggregate.PullEvents() // 反例:事务内清空事件
for _, event := range events {
    pending = append(pending, module.Translate(event))
}
if err := tx.Commit(); err != nil {
    return err // 业务写可能未提交,但内存事件已经丢失
}
```

**正例**:repository / UoW 只能用 `PeekEvents()` 做事务内翻译与序列化校验;`ClearEvents()` 只能在 commit 成功且所有 publish 成功之后调用。新代码不提供 `PullEvents()`;历史代码如已有,标 `Deprecated`,且 repository / UoW 禁止调用。

详见 [`cross-context.md`](./cross-context.md) §3.2。

---

### 7.6 通过 `context.Context` 传递 `*sql.Tx`

**反例**

```go
func (u *UnitOfWork) Run(ctx context.Context, fn func(context.Context) error) error {
    tx, _ := u.db.BeginTx(ctx, nil)
    ctx = context.WithValue(ctx, txKey{}, tx) // 反例:ctx 承载事务句柄
    return fn(ctx)
}

func (r *OrderRepository) Save(ctx context.Context, o *order.Order) error {
    tx := TxFromContext(ctx)
    return saveOrder(ctx, tx, o)
}
```

**正例**:`context.Context` 只传取消、超时、trace。手动事务使用 `RunInTx(ctx, func(ctx, repos) error { ... })`;`repos` 中的 repository 实例内部持有同一个 `*sql.Tx`,方法签名仍是 `Save(ctx, aggregate)` / `Update{Aggregate}(ctx, id, actor, updateFn)`。`updateFn` 可以接收 ctx,但不得从 ctx 取事务。

详见 [`cross-context.md`](./cross-context.md) §4.2。

---

### 7.7 `RunInTx` 闭包内误用普通 repository

**反例**

```go
func (h Handler) Handle(ctx context.Context, cmd Command) error {
    return h.uow.RunInTx(ctx, func(txCtx context.Context, repos Repositories) error {
        return h.orders.Save(txCtx, cmd.Order) // 反例:用了 handler 字段上的普通 repository
    })
}
```

**正例**:`RunInTx` 闭包内所有 DB 操作都只使用参数 `repos` 提供的 tx-bound repository:

```go
func (h Handler) Handle(ctx context.Context, cmd Command) error {
    return h.uow.RunInTx(ctx, func(txCtx context.Context, repos Repositories) error {
        return repos.Orders.Save(txCtx, cmd.Order)
    })
}
```

---

### 7.8 app 层手写 `FindByID` + mutate + `Save`

**反例**

```go
func (h CancelOrderHandler) Handle(ctx context.Context, cmd CancelOrder) error {
    o, err := h.orders.FindByID(ctx, cmd.OrderID)
    if err != nil {
        return err
    }
    if err := o.Cancel(cmd.User); err != nil {
        return err
    }
    return h.orders.Save(ctx, o) // 反例:app 层无法保证 SELECT FOR UPDATE 与事务边界
}
```

**正例**:修改既有聚合时使用 repository closure,由 adapter 在事务内加载并锁定聚合:

```go
func (h CancelOrderHandler) Handle(ctx context.Context, cmd CancelOrder) error {
    return h.orders.UpdateOrder(ctx, cmd.OrderID, cmd.User, func(ctx context.Context, o *order.Order) error {
        return o.Cancel(cmd.User)
    })
}
```

---

### 7.9 把 commit outcome unknown 当普通 error 重试

**反例**

```go
err := repo.Save(ctx, order)
if err != nil {
    retry(err) // 反例:Commit 返回 error 时业务写可能已经提交,盲目重试可能重复写
}
```

**正例**:`CommitOutcomeUnknownError` 必须用 `errors.As` 识别;上层按业务 key 查询确认提交结果后,再决定补偿 / 重放,不得自动重试。

详见 [`cross-context.md`](./cross-context.md) §3.2。

---

### 7.10 把 post-commit publish failure 当业务未写入

**反例**

```go
err := h.app.Commands.PlaceOrder.Handle(ctx, cmd)
if err != nil {
    w.WriteHeader(http.StatusInternalServerError) // 反例:DB 可能已写入,客户端重试会重复下单
    return
}
```

**正例**:`PostCommitPublishError{Committed: true}` 表示业务写已提交但事件发布失败。ports/http、client retry、middleware 必须把它和"业务未写入"区分开,不能自动重放非幂等 command。

详见 [`cross-context.md`](./cross-context.md) §3.2。

---

### 7.11 跨进程 / 跨 BC 用 `errors.Is(err, otherBCSentinel)`

**反例**

```go
// order/adapters/shopclient/products_service.go
if errors.Is(err, shopdomain.ErrProductNotFound) { ... }  // 反例
```

**正例**:shop 的 module contract 自己 export 一份 transport 层的错误(或 error code),shopclient 翻译为 order 域错误。

## 8. 命名

### 8.1 包名是技术名词

**反例**:`package utils`、`package common`、`package helpers`、`package misc`。

**正例**:按业务概念命名。

---

### 8.2 接口加 `I` 前缀

**反例**:`IOrderRepository`、`IShopModule`。

**正例**:`Repository`、`ShopModule`。

---

### 8.3 类型加 `Entity` / `DTO` / `VO` 后缀(domain 类型)

**反例**:`OrderEntity`、`OrderVO`(domain 内)。

**正例**:domain 类型用业务名:`Order`、`OrderItem`、`Money`。**`DTO` 后缀只用于 app/query 和 ports/module 的传输对象**。

---

### 8.4 Handler 方法不叫 `Handle`

**反例**:`(h Handler) Execute(ctx, cmd) error`、`(h Handler) Run(...)`。

**正例**:统一 `Handle`,搜索友好,工具链一致。

## 9. Composition Root(BC module + bootstrap)

### 9.1 业务代码 import bootstrap

**反例**

```go
// internal/order/app/command/place_order.go
import "myshop/internal/bootstrap"  // 反例
```

**正例**:bootstrap 是叶子节点,只能被 `cmd/main.go` import。业务代码绝不反向依赖。

---

### 9.2 bootstrap 包含业务装配细节

**反例**:`bootstrap/server.go` 内出现 `command.PlaceOrderHandler{Orders: mysql.NewOrderRepository(...)}` 这种 BC 内部装配。

**正例**:bootstrap 只做"进程级编排"——打开技术底座、调各 BC 的 `NewModule(Deps)`、做跨 BC wiring、挂 router、启动。BC 内部装配由 `internal/{bc}/module.go` 负责。

---

### 9.3 bootstrap 直接 import BC 内部包

**反例**

```go
// internal/bootstrap/server.go
import (
    ordermysql "myshop/internal/order/adapters/mysql"   // 反例
    ordercmd   "myshop/internal/order/app/command"      // 反例
    orderhttp  "myshop/internal/order/ports/http"       // 反例
)
```

**正例**:bootstrap **只** import 各 BC 的根包(`order`、`shop`、`payments`)和 `platform/`。

```go
import (
    "myshop/internal/order"
    "myshop/internal/payments"
    "myshop/internal/shop"
)

orderMod, _ := order.NewModule(order.Deps{ ... })
```

---

### 9.4 BC 内部装配散落在 bootstrap

**反例**:`bootstrap/order_module.go` 持有 order BC 的装配代码,`internal/order/` 内没有 `module.go`。

**正例**:每个 BC 拥有 `internal/{bc}/module.go`,定义 `Deps`、`Module`、`NewModule(Deps) (*Module, error)`。bootstrap 只调这个工厂。

理由:

- BC 模块完整内聚,拆服务时整个 BC 目录可整体迁移
- 装配变更只影响本 BC,bootstrap 不动
- BC 测试可直接 `order.NewModule(testDeps)`,无需复制 bootstrap 逻辑

---

### 9.5 `{bc}/module.go` 打开 DB / Redis / HTTP server

**反例**

```go
// internal/order/module.go
func NewModule(cfg config.Config) (*Module, error) {
    db, _ := sql.Open("mysql", cfg.MySQL.DSN)  // 反例
    ...
}
```

**正例**:`Deps` 接收已打开的 `*sql.DB`、`eventbus.Bus` 等技术底座实例。打开它们是 bootstrap 的事。

---

### 9.6 `{bc}/module.go` 跨 BC import 对方任何包(包括 `ports/module/`)

**反例**

```go
// internal/order/module.go
import (
    shopapp "myshop/internal/shop/app"           // 反例
    "myshop/internal/shop/adapters/mysql"        // 反例
    shopmod "myshop/internal/shop/ports/module"  // 反例:即便是契约层也不允许
)

type Deps struct {
    ShopModule shopmod.ShopModule // 反例:直接持有对方类型
}
```

**正例**:跨 BC 依赖通过 `Deps.Products command.ProductsService`(本 BC 自己定义的接口)注入。`module.go` **不 import 任何**其他 BC 的包,包括对方的 `ports/module/`。

**原因**:Go `internal/` 边界规则下,等到本 BC 被剥离到独立仓库时,跨仓库的 `internal/.../ports/module/` 物理上将不可 import,这一行会让"剥离零变更"的承诺破产。

---

### 9.7 跨 BC 依赖隐式注入(不在 `Deps` 显式声明)

**反例**:`{bc}/module.go` 内部直接 `New` 一个其他 BC 的 client,隐藏了 BC 依赖关系。

**正例**:**所有跨 BC 依赖必须出现在 `Deps` 结构体字段**,这是 BC 之间依赖关系的 single source of truth。grep `Deps` 一眼看清依赖图。

---

### 9.8 `{bc}/module.go` 内部 `new` 跨 BC adapter

**反例**

```go
// internal/order/module.go
import "myshop/internal/order/adapters/shopclient"  // 反例:模块内部触发

func NewModule(deps Deps) (*Module, error) {
    products := shopclient.NewProductsService(deps.ShopModule) // 反例
    ...
}
```

`shopclient` 依赖 shop 的 `ports/module`,在 module.go 内部 new 等于把 shop 的 import 传染过来。

**正例**:跨 BC adapter 由 bootstrap 实例化并通过 `Deps` 注入:

```go
// bootstrap/server.go
products := shopclient.NewProductsService(shopMod.PortsModule)
orderMod, _ := order.NewModule(order.Deps{Products: products, ...})
```

`module.go` 既不 import `shopclient`,也不 import `shopmod`。

---

### 9.9 把 `*sql.DB` 一类基础设施句柄抽成接口

**反例**

```go
type DBHandle interface { ... } // 反例:为了"接口化而接口化"
type Deps struct {
    DB DBHandle
}
```

**正例**:基础设施技术句柄(`*sql.DB`、`*redis.Client`、`*http.Client`)**直接用具体类型**。这些不是"业务依赖",是平台提供的标准句柄,生命周期由 bootstrap 持有。

接口化只用在三类场景:横切能力(Clock / IDGen / Tracer)、平台抽象(eventbus)、跨 BC 依赖。详见 [ADR-0007 §Deps 设计准则](./adr/0007-bc-module-factory.md)。

## 10. 测试

### 10.1 domain test 起数据库

**反例**:`domain/order/order_test.go` 内 testcontainers MySQL。

**正例**:domain 是纯逻辑,测试零外部依赖。

---

### 10.2 app test 用 `sqlmock`

**反例**:app handler 测试 mock 了 SQL。

**正例**:app 通过接口依赖 adapter,测试 mock 接口,不应该感知 SQL。

---

### 10.3 集成测试不用 build tag,默认跑

**反例**:`go test ./...` 直接连数据库。

**正例**:集成测加 `//go:build integration`,CI 分阶段。

---

### 10.4 直接 `time.Now()` / `uuid.New()`

**反例**:domain / app 内直接调,测试无法重复。

**正例**:注入 `Clock` / `IDGen` 接口。

## 11. 杂项

### 11.1 出现 `pkg/`

**反例**:仓库根有 `pkg/` 目录。

**正例**:所有代码在 `internal/`,避免给外部 import 留口子。

---

### 11.2 一个 BC 有多个 `app.Application`

**反例**:`order/app/v1/app.go` + `order/app/v2/app.go`。

**正例**:每个 BC 一个 `app.Application`。HTTP 版本演进通过 `ports/http/v1`、`ports/http/v2` 调同一个 app 解决。

---

### 11.3 跨 BC 直接共享数据库表

**反例**:order BC 和 shop BC 共用 `products` 表,各自的 ORM 读写。

**正例**:每个 BC 拥有自己的表,跨 BC 数据通过 Module Contract / Integration Event 流转。即便物理上是同一个 MySQL 实例,逻辑上也要隔离。

---

### 11.4 通过私有 export 字段绕过封装

**反例**

```go
//nolint:unused
// 仅用于测试
func (o *Order) SetStatusForTest(s Status) { o.status = s }
```

**正例**:测试通过业务方法构造场景(`order.Place()` → `order.Confirm()` → `order.Cancel()`),或用包内 Builder。
