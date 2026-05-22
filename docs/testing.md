# 测试组织

本文档定义各层测试的边界、工具、运行方式。**测试代码与生产代码共享同一套架构纪律。**

> 前置阅读:[`architecture.md`](./architecture.md)、[`cross-context.md`](./cross-context.md)

## 1. 测试金字塔

```text
              ┌─────────────────┐
              │  e2e / cross-BC │  少量、跨进程边界
              ├─────────────────┤
              │  ports test     │  中量、HTTP / 事件层
              ├─────────────────┤
              │  app / adapters │  较多、用例级
              ├─────────────────┤
              │  domain unit    │  最多、纯逻辑
              └─────────────────┘
```

| 层 | 形态 | 工具 | 数量级 |
|---|---|---|---|
| `domain/` | 纯单测,无 mock | `testing` + `testify` | 最多 |
| `app/command/`、`app/query/` | 单测,mock 接口 | `testify/mock` 或手写 fake | 较多 |
| `adapters/{persistence}/` | 集成测,真数据库 | `testcontainers-go` | 较多 |
| `adapters/{vendor}/` | 集成测,真服务 or 录制响应 | `httptest` / VCR | 较少 |
| `adapters/{bc}client/` | 单测,mock 对方 `ports/module/` 接口 | mock | 中量 |
| `ports/http/` | HTTP 测试,mock app | `httptest` | 中量 |
| `ports/eventsub/` | 单测,mock app | 直接调 handler | 中量 |
| `ports/module/` | 集成测,真 app + 内存/容器 db | testcontainers | 较少 |
| 跨 BC 端到端 | 黑盒,启动 bootstrap | testcontainers + HTTP | 少量 |

## 2. 文件位置

测试文件**与被测代码同包同目录**,Go 标准做法。例外:

- 跨 BC 端到端测试放在 `internal/{bc}/e2e_test.go` 或仓库根的 `tests/e2e/`,**专门用 build tag 隔离**
- 集成测试用 build tag 区分(下节)

## 3. Build Tag 与运行策略

```go
//go:build integration
// +build integration

package mysql_test
```

| 类型 | build tag | 默认执行 | 命令 |
|---|---|---|---|
| 单元测试 | 无 | 是 | `go test ./...` |
| 集成测试 | `integration` | 否 | `go test -tags=integration ./...` |
| 端到端 | `e2e` | 否 | `go test -tags=e2e ./...` |

CI 流水线:`unit → integration → e2e`,失败立即中止。

## 4. 各层测试细则

### 4.1 Domain 层

- **纯单测,零外部依赖**,不需要 mock
- **表驱动**为主
- 覆盖:不变量校验、状态机转换、业务方法返回的 domain event

```go
func TestOrder_Place_HappyPath(t *testing.T) {
    tests := []struct {
        name  string
        items []order.OrderItem
        want  order.Status
    }{
        {"single item", []order.OrderItem{...}, order.StatusPlaced},
        ...
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            o, err := order.NewOrder("user-1", tc.items)
            require.NoError(t, err)
            require.Equal(t, tc.want, o.Status())
        })
    }
}
```

**禁止**:domain 测试中出现 `*sql.DB`、`*http.Request`、`gomock` 等。

### 4.2 App 层

- 用例级单测,**mock 所有接口**(Repository、ReadModel、services.go 中定义的外部能力)
- 不依赖真实数据库,不依赖 HTTP server
- 覆盖:用例编排、错误传递、跨聚合协同

```go
func TestPlaceOrder_StockShortage_ReturnsError(t *testing.T) {
    repo := newFakeOrderRepo()
    products := newFakeProductsService()
    products.SetStock("p-1", 0)

    h := command.PlaceOrderHandler{
        Orders:   repo,
        Products: products,
    }

    err := h.Handle(ctx, command.PlaceOrder{
        UserID: "u-1",
        Items:  []command.PlaceOrderItem{{ProductID: "p-1", Qty: 1}},
    })

    require.ErrorIs(t, err, command.ErrInsufficientStock)
    require.Empty(t, repo.Saved())
}
```

**Mock 选型**:

- 简单接口:**手写 fake**(更可读,无生成代码维护成本)
- 接口方法多 / 校验调用次数 / 校验参数:用 `testify/mock` 或 `mockery`

推荐:**优先手写 fake**,只在确实需要复杂校验时上 mock 框架。

#### 事务 typed error 覆盖

如果 command 调用的 repository / UnitOfWork 可能返回 `platform/txerr` 中的事务语义错误,app 层必须覆盖这些路径:

| error | app 层预期 |
|---|---|
| `txerr.CommitOutcomeUnknownError` | 不自动重试;按业务 key 查询确认提交结果,再决定补偿 / 重放 |
| `txerr.PostCommitPublishError{Committed: true}` | 不当作业务未写入;向上保留 `Committed=true` 语义,供 ports / retry middleware 区分 |

ports/http 或 retry middleware 也必须有单测证明:commit outcome unknown 不触发盲目重试;post-commit publish failure 不被包装成普通 5xx 后由客户端自动重放同一非幂等 command。

### 4.3 Adapters 层

#### `adapters/{persistence}/`

- **真数据库,testcontainers**
- 每个测试用例独立 schema 或 transaction rollback
- 覆盖:CRUD、并发、事务、错误翻译
- Repository / UoW 事务必须覆盖:
  - 普通 `Save` 自动开启事务;commit 前失败 rollback,不 publish,不 `ClearEvents`
  - `Update{Aggregate}` 在同一事务内加载并锁定聚合、执行 `updateFn(ctx, aggregate)`、保存并 commit 后 publish
  - `Update{Aggregate}` 的 `updateFn` 返回 error 时 rollback,不 publish,不 `ClearEvents`
  - `RunInTx` 闭包返回 error 时 rollback,不 flush pending events
  - `RunInTx` commit 成功后统一 publish pending events,全部成功后统一 `ClearEvents`
  - `Commit` 返回 error 时返回 `txerr.CommitOutcomeUnknownError`,且不 publish
  - post-commit publish 失败时返回 `txerr.PostCommitPublishError{Committed: true}`

code review / 静态检查项:app 层接口不暴露 `*sql.Tx`;repository 不通过 `context.Context` 读取事务。

```go
//go:build integration

func TestOrderRepository_SaveAndFind(t *testing.T) {
    db := setupMySQL(t) // testcontainers
    repo := mysql.NewOrderRepository(db)

    o, _ := order.NewOrder(...)
    require.NoError(t, repo.Save(ctx, o))

    got, err := repo.FindByID(ctx, o.ID())
    require.NoError(t, err)
    require.Equal(t, o.ID(), got.ID())
}
```

`setupMySQL(t)` 抽到 `internal/platform/mysql/testing.go` 或 `internal/{bc}/adapters/mysql/testutil_test.go`。

#### `adapters/{vendor}/`

- 推荐用 `httptest` 起一个模拟服务器,而不是真打第三方
- 或用 VCR(`go-vcr`)录制 + 回放真实响应
- 真打第三方仅在 nightly / smoke 测试中

#### `adapters/{bc}client/`

- 单测即可,**mock 对方 `ports/module/` 接口**
- 覆盖:DTO 转换、错误翻译

```go
func TestProductsService_GetProduct_NotFound(t *testing.T) {
    shop := newMockShopModule()
    shop.On("GetProduct", mock.Anything, mock.Anything).Return(
        shopmod.ProductDTO{}, shopmod.ErrProductNotFound,
    )

    svc := shopclient.NewProductsService(shop)
    _, err := svc.GetProduct(ctx, "p-1")
    require.ErrorIs(t, err, command.ErrProductNotFound) // 翻译为 order 域错误
}
```

### 4.4 Ports 层

#### `ports/http/`

- `httptest` + **mock app**
- 覆盖:路由、参数解析、响应码、错误响应格式

```go
func TestPlaceOrderHandler_BadRequest(t *testing.T) {
    app := newMockApp()
    h := orderhttp.NewOrderHandler(app)

    req := httptest.NewRequest("POST", "/orders", strings.NewReader(`{}`))
    rec := httptest.NewRecorder()

    h.PlaceOrder(rec, req)

    require.Equal(t, http.StatusBadRequest, rec.Code)
}
```

#### `ports/eventsub/`

- 构造 `eventbus.Envelope`(含 `event_id` + 序列化 payload),调 handler,断言 app 被正确调用
- 覆盖:**envelope decode、DTO → command 翻译、委托 app**(handler 层职责边界)

**eventsub 单测矩阵(每个 handler 必须覆盖)** — 只验证协议层,**不**用 mock 断言"业务幂等":

| 场景 | 预期行为 | 证明什么 |
|---|---|---|
| 正常 envelope | decode 成功,app handler 被调用一次 | decode + 委托 |
| Decode 失败 | 返回 error,app handler 不被调用 | 错误不穿透到 app |
| 重复投递同一 envelope | app handler **再次被调用**(调用次数 = 投递次数) | handler 会重复委托;幂等不在此层保证 |

```go
func TestOrderPlacedHandler_HappyPath(t *testing.T) {
    app := newMockInitializePaymentHandler()
    h := eventsub.NewOrderPlacedHandler(app)

    // 本期 event_id 是投递尝试 ID;测试中固定值便于断言与日志关联。
    env := eventbus.NewEnvelope(t,
        eventbus.EventID("evt-001"),
        eventbus.EventType("order.OrderPlacedV1"),
        ordermod.OrderPlacedV1{OrderID: "o-1", Total: 100},
    )

    require.NoError(t, h.Handle(ctx, env))
    require.Equal(t, 1, app.HandleCalls)
    require.Equal(t, "o-1", app.LastOrderID)
}

func TestOrderPlacedHandler_DecodeFailed(t *testing.T) {
    app := newMockInitializePaymentHandler()
    h := eventsub.NewOrderPlacedHandler(app)

    // 直接构造 raw Envelope,避免 NewEnvelope 对 []byte 再 marshal 成合法 JSON
    env := eventbus.Envelope{
        EventID:   "evt-002",
        EventType: "order.OrderPlacedV1",
        Payload:   []byte(`{invalid json`),
    }

    err := h.Handle(ctx, env)
    require.Error(t, err)
    require.Equal(t, 0, app.HandleCalls)
}

func TestOrderPlacedHandler_RepeatedEnvelope_CallsAppAgain(t *testing.T) {
    app := newMockInitializePaymentHandler()
    h := eventsub.NewOrderPlacedHandler(app)

    env := eventbus.NewEnvelope(t,
        eventbus.EventID("evt-003"),
        eventbus.EventType("order.OrderPlacedV1"),
        ordermod.OrderPlacedV1{OrderID: "o-1", Total: 100},
    )

    require.NoError(t, h.Handle(ctx, env))
    require.NoError(t, h.Handle(ctx, env))
    require.Equal(t, 2, app.HandleCalls, "重复 envelope 会重复委托 app")
}
```

> **不要**直接构造 DTO 调 handler——这绕开了 envelope,无法覆盖 handler 的实际入口签名。
> **不要**在 eventsub 单测里用 mock 的 `IdempotentByOrderID` 之类标志断言"无副作用"——那只能证明 mock 行为,形成假覆盖。

#### 业务 command 幂等(必须在 app / persistence 层验证)

真正的幂等(唯一约束、upsert、状态检查、物化视图单调 version 更新)必须在以下层测试:

| 层 | 测什么 | 工具 |
|---|---|---|
| `app/command/` | 同一 command 入参调用两次,无副作用(可 mock repo 记录写入次数) | 单测 + fake repo |
| `adapters/{persistence}/` | 唯一约束冲突、upsert、`WHERE version < ?` 等真实 SQL 语义 | `integration` tag + testcontainers |

```go
//go:build integration

func TestInitializePayment_IdempotentByOrderID(t *testing.T) {
    db := setupMySQL(t)
    h := setupInitializePaymentHandler(t, db)

    cmd := command.InitializePayment{OrderID: "o-1", Amount: 100}
    require.NoError(t, h.Handle(ctx, cmd))
    require.NoError(t, h.Handle(ctx, cmd)) // 模拟 at-least-once 重复 command

    count, err := countPaymentsForOrder(ctx, db, "o-1")
    require.NoError(t, err)
    require.Equal(t, 1, count)
}
```

#### `ports/module/`

- **集成测**:真 app + testcontainers db
- 这是契约层,必须保证发布给其他 BC 的能力真的能用

```go
//go:build integration

func TestShopModule_GetProduct(t *testing.T) {
    app := setupShopApp(t)
    m := module.NewShopModule(app)

    // 通过 app/command 写入数据
    require.NoError(t, app.Commands.CreateProduct.Handle(ctx, ...))

    dto, err := m.GetProduct(ctx, "p-1")
    require.NoError(t, err)
    require.Equal(t, "p-1", string(dto.ID))
}
```

### 4.5 单 BC 集成测试

得益于 `internal/{bc}/module.go` 工厂,可以**直接装配一个 BC**做端到端测试,不依赖 bootstrap。

```go
//go:build integration

func TestOrder_PlaceOrder_E2E(t *testing.T) {
    db := setupMySQL(t)
    eb := eventbus.NewInMemory()

    // Deps.Products 是 order 自己定义的接口(command.ProductsService),
    // 测试里直接用 fake 满足,无需启动 shop。
    products := newFakeProductsService(t)
    products.PutProduct("p-1", 100)

    mod, err := order.NewModule(order.Deps{
        DB:       db,
        EventBus: eb,
        Logger:   slog.Default(),
        Clock:    clock.Fixed(time.Now()),
        IDGen:    idgen.NewSequential(),
        Products: products, // 跨 BC 接口 fake
    })
    require.NoError(t, err)

    err = mod.App.Commands.PlaceOrder.Handle(ctx, command.PlaceOrder{ ... })
    require.NoError(t, err)
}
```

这是 module 工厂模式的最大测试收益:**单 BC 可以独立起一个完整运行单元**,跨 BC 依赖用 fake 满足。

### 4.6 跨 BC 端到端

- 启动完整 bootstrap(可仅启用相关 BC)
- 通过 HTTP 调用,断言 integration event 被正确流转
- 数量少而精,只覆盖关键 happy path

```go
//go:build e2e

func TestE2E_OrderPlaced_PaymentInitialized(t *testing.T) {
    sys := bootstrap.StartForTest(t, bootstrap.Config{
        EnabledBCs: []string{"order", "payments", "shop"},
        // ... 测试用 config
    })
    defer sys.Stop()

    resp := sys.HTTP.POST("/orders", ...)
    require.Equal(t, 201, resp.Code)

    require.Eventually(t, func() bool {
        return sys.Payments.HasPaymentForOrder(orderID)
    }, 3*time.Second, 50*time.Millisecond)
}
```

`bootstrap.StartForTest` 是测试专用入口,内部仍然走 `{bc}.NewModule(Deps)` 的标准装配路径。

## 5. Fixture 与 Builder

- 大量重复构造领域对象时用 Builder 模式
- Builder 放在被测包同目录的 `*_test.go` 中(包内)或 `testdata/` 中

```go
// order_test_builder.go(test 包内)
func orderBuilder() *OrderBuilder { return &OrderBuilder{userID: "u-1", items: defaultItems()} }
func (b *OrderBuilder) WithStatus(s Status) *OrderBuilder { b.status = s; return b }
func (b *OrderBuilder) Build(t *testing.T) *Order {
    o, err := NewOrder(b.userID, b.items)
    require.NoError(t, err)
    return o
}
```

## 6. 并发与时间

- 时间依赖:domain / app 接收 `Clock interface { Now() time.Time }`,测试注入 `FakeClock`
- 随机依赖:同上,`IDGen interface { Next() string }`
- **禁止**直接调用 `time.Now()` / `uuid.New()`,把"现在"和"随机"显式化,测试才可重复

## 7. 覆盖率

- domain:**> 90%**
- app:**> 80%**
- adapters:**> 70%**(集成测覆盖)
- ports:**> 70%**
- 不强求绝对值,但 PR 不允许覆盖率下降

CI 输出 `go test -cover` 摘要,放在 PR comment。

## 8. CI 编排

```yaml
jobs:
  lint:
    - go vet
    - golangci-lint run
    - import-boundary check    # 依赖方向校验

  unit:
    - go test ./...

  integration:
    - 启动 testcontainers
    - go test -tags=integration ./...

  e2e:
    - go test -tags=e2e ./tests/e2e/...
```

## 9. 反例

| 反例 | 正解 |
|---|---|
| domain test 起 MySQL | domain 是纯逻辑,不需要 |
| app test 用 sqlmock 模拟 SQL | 用 fake repository,SQL 是 adapter 的事 |
| ports/http test 真打数据库 | mock app |
| 集成测试不用 build tag,默认跑 | 用 `integration` tag 隔离 |
| 直接 `time.Now()` 难以测试 | 注入 Clock |
| 一个测试函数测 N 个场景没有 subtest | 表驱动 + `t.Run` |
