# 命名规约

本文档定义包名、文件名、类型、接口、错误、方法的命名规则。规则违反在 code review 时是 blocking。

> 前置阅读:[`architecture.md`](./architecture.md)、[`cross-context.md`](./cross-context.md)

## 1. 包名

### 1.1 通用规则

- **小写、无下划线、无连字符**
- **单数**(`order` 而非 `orders`、`product` 而非 `products`)
- **业务名词**,不是技术名词
- **不允许**:`utils`、`helpers`、`common`、`misc`、`tools`

### 1.2 各层包名

| 路径 | 包名 |
|---|---|
| `internal/bootstrap/` | `bootstrap` |
| `internal/{bc}/`(根,含 `module.go`) | `{bc}`(如 `order`、`shop`、`payments`) |
| `internal/{bc}/domain/{aggregate}/` | `{aggregate}`(如 `order`、`product`) |
| `internal/{bc}/app/` | `app` |
| `internal/{bc}/app/command/` | `command` |
| `internal/{bc}/app/query/` | `query` |
| `internal/{bc}/ports/http/` | `http`(可用别名避开标准库冲突) |
| `internal/{bc}/ports/ws/` | `ws` |
| `internal/{bc}/ports/stream/` | `stream` |
| `internal/{bc}/ports/eventsub/` | `eventsub` |
| `internal/{bc}/ports/module/` | `module` |
| `internal/{bc}/adapters/mysql/` | `mysql` |
| `internal/{bc}/adapters/redis/` | `redis` |
| `internal/{bc}/adapters/{other_bc}client/` | `{other_bc}client`(如 `shopclient`) |
| `internal/{bc}/adapters/{vendor}/` | `{vendor}`(如 `stripe`、`s3`) |
| `internal/shared/bizerr/` | `bizerr` |
| `internal/platform/{component}/` | `{component}` |

> BC 根目录(如 `internal/order/`)既是目录也是 Go 包,包名就是 BC 名(`package order`),里面只有 `module.go`(含 `Deps`、`Module`、`NewModule`)。**不**在 BC 根放业务代码,业务全部在子包里。

### 1.3 容易混淆的命名

- `internal/{bc}/ports/http/` 包名 `http` 会与标准库冲突,**import 时用别名**:
  ```go
  import (
      "net/http"

      orderhttp "myshop/internal/order/ports/http"
  )
  ```
  内部代码自己用 `http`(本包名)是 OK 的。

- 跨 BC adapter 必须是 `{other_bc}client/`,**不能**用对方 BC 名:
  ```text
  正确:order/adapters/shopclient/
  错误:order/adapters/shop/
  ```
  原因:`order/adapters/shop/` 容易被误读为"shop BC 内嵌在 order 下"。

- BC 自己的 module 工厂(`internal/{bc}/module.go`,包名 `{bc}`)与对外契约 `internal/{bc}/ports/module/`(包名 `module`)**不是同一个东西**:

  | 路径 | 包名 | 含义 |
  |---|---|---|
  | `internal/order/module.go` | `order` | BC 内部装配代码 |
  | `internal/order/ports/module/` | `module` | BC 对其他 BC 暴露的进程内契约 |

  import 时给后者别名以避免混淆:
  ```go
  import (
      "myshop/internal/order"                            // BC 根包
      ordermod "myshop/internal/order/ports/module"      // 对外契约
  )
  ```

## 2. 文件名

- **snake_case**:`place_order.go`、`order_repository.go`
- **一文件一用例 / 一聚合 / 一概念**
- 测试文件:`{name}_test.go`、`{name}_integration_test.go`

### 2.1 各类文件命名约定

| 文件用途 | 文件名 |
|---|---|
| Aggregate Root 主文件 | `{aggregate}.go`(如 `order.go`) |
| Repository 接口 | `repository.go` |
| Domain Event | `events.go` |
| typed errors | `errors.go` |
| 一个写用例 | `{verb_noun}.go`(如 `place_order.go`) |
| 一个读用例 | `get_{noun}.go` / `list_{noun}.go` / `search_{noun}.go` |
| Query DTO | `types.go` |
| Query ReadModel 接口 | `read_model.go` |
| 外部能力接口 | `services.go` |
| Application 容器 | `app.go` |
| Module Contract 接口 + 实现 | `contract.go` |
| Module Transport DTO | `types.go` |
| Integration Event | `events.go` |
| HTTP Handler | `{noun}_handler.go` |
| HTTP Router | `router.go` |
| BC 内部装配工厂 | `internal/{bc}/module.go` |
| 进程级编排 | `internal/bootstrap/server.go` |
| Repository 实现 | `{aggregate}_repository.go` |
| ReadModel 实现 | `{aggregate}_read_model.go` |
| UnitOfWork 实现 | `uow.go` |
| 事件发布辅助 | `event_publisher.go` |
| 跨 BC adapter | `{capability}_service.go`(如 `products_service.go`) |

## 3. 类型命名

### 3.1 聚合根、Entity、Value Object

- **PascalCase**,简洁、单数:`Order`、`OrderItem`、`Money`
- **不带 `Domain` / `Entity` / `VO` 前后缀**
- 内部字段全部小写不导出,通过方法访问

```go
type Order struct {
    id     OrderID
    status Status
    items  []OrderItem
}

func (o *Order) ID() OrderID { return o.id }
```

### 3.2 构造函数

| 用途 | 命名 |
|---|---|
| 全量校验创建新聚合 | `NewOrder(...)` |
| 从持久化恢复(绕过校验) | `Rehydrate(...)` |
| 从外部 BC DTO 恢复(罕见) | `UnmarshalOrderFromExternal(...)` |

`Rehydrate` **只允许 `adapters/{persistence}/` 调用**。当同一个包内确实有多个需要恢复的类型且 Go 无法重载函数名时,再用类型名消歧,如 `user.RehydrateAddress(...)`。

### 3.3 Command / Query 类型

- Command 结构体:动词短语,PascalCase:`PlaceOrder`、`CancelOrder`
- Command Handler:`PlaceOrderHandler`
- Query 结构体:`GetOrder`、`ListOrders`
- Query Handler:`GetOrderHandler`

每个 Handler 暴露唯一方法 `Handle(ctx, cmd) error`(或 `Handle(ctx, q) (Result, error)`),不要叫 `Execute`、`Run`、`Do`。

```go
type PlaceOrder struct {
    UserID  string
    Items   []PlaceOrderItem
}

type PlaceOrderHandler struct {
    orders   order.Repository
    products ProductsService
    uow      UnitOfWork
}

func (h PlaceOrderHandler) Handle(ctx context.Context, cmd PlaceOrder) error { ... }
```

### 3.4 DTO

- Query DTO:`OrderDTO`、`OrderListItemDTO`(放在 `app/query/types.go`)
- Module Transport DTO:`ProductDTO`、`OrderSnapshotDTO`(放在 `ports/module/types.go`)
- **统一用 `DTO` 后缀**,与 domain 类型区分,搜索友好

### 3.5 ID 类型

每个聚合一个 typed ID:

```go
type OrderID string

func NewOrderID() OrderID { return OrderID(uuid.NewString()) }
func (id OrderID) String() string { return string(id) }
```

避免 `string` 满天飞、误传 UserID 给 OrderID。

## 4. 接口命名

### 4.1 通用规则

- **接口在调用方定义**(Go 的 idiom)
- 单方法接口用动词 + `er` 后缀:`Reader`、`Writer`、`Closer`(标准库风格)
- 多方法接口用名词:`Repository`、`ReadModel`、`ShopModule`
- **不**用 `I` 前缀(如 `IRepository`)

### 4.2 各层接口命名

| 接口位置 | 命名形态 | 例子 |
|---|---|---|
| `domain/{aggregate}/repository.go` | `Repository` | `order.Repository` |
| `app/query/read_model.go` | `ReadModel` 或 `{Aggregate}ReadModel` | `OrderReadModel` |
| `app/command/services.go` | 调用方视角的能力名 | `ProductsService`、`PaymentGateway` |
| `ports/module/contract.go` | `{Bc}Module` | `ShopModule`、`PaymentsModule` |

### 4.3 接口粒度

- **services.go 中的接口要小**:只声明本 BC 真正用到的方法
- **module contract 接口可以大**:汇总本 BC 对外的所有能力
- 在 adapter 处做适配,Go 隐式实现无需 type assertion

## 5. 错误命名

### 5.1 Sentinel Error

- 包级变量,`Err` 前缀,PascalCase:`ErrOrderNotFound`、`ErrInsufficientStock`
- 放在 `errors.go`

```go
var (
    ErrOrderNotFound      = errors.New("order not found")
    ErrInvalidOrderStatus = errors.New("invalid order status")
)
```

### 5.2 Typed Error

需要携带上下文时用结构体 + `Error()` + 可选 `Unwrap()`:

```go
type StockShortageError struct {
    ProductID string
    Required  int
    Available int
}

func (e *StockShortageError) Error() string { ... }
```

### 5.3 错误翻译

- adapter 必须把基础设施错误翻译为领域错误
- 跨 BC adapter 必须把对方域错误翻译为本 BC 域错误
- **禁止** `errors.Is(err, gorm.ErrRecordNotFound)` 出现在 domain / app 层

## 6. 方法命名

### 6.1 Aggregate Root 方法

- 业务动词:`order.Place()`、`order.Cancel()`、`product.AdjustStock(n)`
- **不**用 `Set` / `Update` 系列,这些是数据语义不是业务语义
- 返回 domain event(可选):`(event order.OrderPlaced, err error)`

### 6.2 Repository 方法

- `Save(ctx, entity)`:保存新建聚合,或保存调用方已经通过合法路径持有的聚合实例;不要把它当作既有聚合的通用 upsert 入口
- `Update{Aggregate}(ctx, id, actor, updateFn)`:既有聚合的读锁-修改-保存 closure,如 `UpdateOrder(ctx, orderID, actor, updateFn)`;`actor` 是领域上下文参数占位,可落成 `User`、`Operator` 等具体类型
- `FindByID(ctx, id)`:查询单个
- `FindBy{Criteria}(ctx, ...)`:按某条件查询
- **避免** `Get`、裸 `Update`、`Delete` 这种数据库语义;`Update{Aggregate}` 只用于封装单聚合事务 closure,不是 CRUD DAO 方法

### 6.3 Service 方法

- 业务动词:`PaymentGateway.Charge(...)`、`ProductsService.GetProduct(...)`
- 第一参数永远是 `context.Context`

### 6.4 Handler 方法

- 统一叫 `Handle`,不用 `Execute` / `Run` / `Do`

### 6.5 UnitOfWork 与事务辅助

- UnitOfWork 方法:`RunInTx(ctx, fn)`,不用 `Run` / `DoInTx`
- Handler 字段名:`uow UnitOfWork`
- tx-bound repository 构造函数:`New{Aggregate}RepositoryWithTx`(如 `NewOrderRepositoryWithTx`)
- adapter 如需额外依赖,依赖 struct 统一叫 `Deps`;构造函数可写成 `New{Aggregate}Repository(db *sql.DB, deps Deps)` 或 `NewUnitOfWork(db *sql.DB, deps Deps)`
- 平台事务辅助类型:`PendingCollector`、`PendingPublish`

## 7. Integration Event 命名

- 过去式 + 版本后缀:`OrderPlacedV1`、`PaymentConfirmedV1`、`ProductPriceChangedV1`
- 版本是契约的一部分,**绝不省略**
- 字段命名 PascalCase,与 JSON 字段对齐(用 struct tag)

```go
type OrderPlacedV1 struct {
    OrderID string `json:"order_id"`
    UserID  string `json:"user_id"`
    Total   int64  `json:"total"`
    PlacedAt time.Time `json:"placed_at"`
}
```

## 8. 测试命名

- 文件:`{name}_test.go`、`{name}_integration_test.go`
- 函数:`TestXxx_{场景}_{期望}`,如 `TestPlaceOrder_StockShortage_ReturnsError`
- 表驱动测试用 subtest:`t.Run(tc.name, func(t *testing.T){...})`

详见 [`testing.md`](./testing.md)。

## 9. 常见违规速查

| 违规 | 正确做法 |
|---|---|
| `package utils` | 按职责命名 |
| `IOrderRepository` | `Repository` |
| `OrderEntity` | `Order` |
| `order_dto.go` | `types.go` |
| `(o *Order) UpdateStatus(s Status)` | `(o *Order) Confirm()` / `(o *Order) Cancel()` 等业务动词 |
| `handler.Execute(...)` | `handler.Handle(...)` |
| `OrderPlaced` 作为跨 BC 事件 | `OrderPlacedV1` |
| `order/adapters/shop/` | `order/adapters/shopclient/` |
| `bootstrap/order_module.go`(BC 装配放 bootstrap) | `internal/order/module.go`(BC 装配放 BC 自己) |
