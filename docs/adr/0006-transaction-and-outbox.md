# ADR-0006: 事务边界与事件发布

- **状态**:Accepted
- **日期**:2026-05-20
- **决策者**:架构组

## 背景

事务一致性在 modular monolith 中有三种典型场景:

1. 单 BC 内、单聚合写
2. 单 BC 内、多聚合写
3. 跨 BC 协作

跨 BC 协作不能用 2PC / 分布式事务(性能、运维、未来拆服务的可移植性都不允许)。需要明确各场景的事务边界与事件发布机制。

## 决策

### 1. 单聚合写

- 事务边界在 `adapters/{persistence}/{aggregate}_repository.go` 内。新建聚合或调用方已持有聚合实例时使用 `Save(ctx, aggregate)`;修改既有聚合时优先提供 `Update{Aggregate}(ctx, id, actor, updateFn)` closure,由 repository 在事务内加载并锁定聚合
- `updateFn` 签名包含 `ctx context.Context`,用于传递取消、超时、trace;**不得**通过 ctx 携带 `*sql.Tx`,也不应在事务闭包内做慢外部 IO
- **事务内**:写聚合表 → `aggregate.PeekEvents()`(不 drain) → `ports/module.Translate(...)` + payload 序列化校验(失败则回滚;aggregate 仍持有事件,可重试)
- **commit 成功后**:生成本次投递尝试的 `event_id` → `eventbus.Publish` → 全部成功后 `aggregate.ClearEvents()`
- **失败语义**:commit 前失败可重试同一 `Save`,或重试整个 `Update{Aggregate}` command;`Commit` 返回 error 时结果未知,不 publish,上层按业务 key 查询确认;**commit 成功后**任意 `Save` / `Update{Aggregate}` error(含 publish 失败)或 **`ClearEvents` 后**均禁止复用同一 aggregate,也不得复用 update closure 内拿到的 aggregate 指针。post-commit publish 失败必须返回 typed error(如 `PostCommitPublishError{Committed: true}`),不能被自动重试误判为业务未写入。详见 [`../cross-context.md`](../cross-context.md) §3.2
- **app/command Handler 不感知事务,也不感知 Integration Event**(关键:这避免了 app → ports 的反向 import)
- Integration Event 翻译位置选 repository 而非 app/command 的原因见 [ADR-0005](./0005-events-internal-vs-integration.md) 与 [`../cross-context.md`](../cross-context.md) §3.2

### 2. 单 BC 内多聚合写

- **首选**:重新审视聚合边界。如果两个聚合**总是**被一起修改,它们可能本来就该是同一个聚合
- 仍需多聚合写时:在本 BC 的 `app/command/services.go` 引入 **BC-local** `UnitOfWork` 接口和 repository bundle

```go
type UnitOfWork interface {
    RunInTx(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type Repositories struct {
    Orders order.Repository
    // 只放本 BC 内需要共享同一事务的 repository
}
```

- `app` 层只看到上述非泛型接口;不要暴露 `*sql.Tx`,也不要通过 `context.Context` 传事务
- `internal/platform/sqltx` 提供泛型事务模板,类型参数是本 BC 的 `Repositories`;`adapters/{persistence}/uow.go` 只负责把同一个 `*sql.Tx` 绑定到本 BC 的 repository 实例
- tx-bound repository 不得在子 `Save` / `Update{Aggregate}` 内 commit / publish / `ClearEvents`;`PeekEvents` + `Translate` 结果追加到 UoW 级 pending collector,由 `RunInTx` 在外层 **commit 成功后**统一 publish,再按 aggregate 去重后 `ClearEvents`。若 UoW `Commit` 返回 error,不 publish,上层按业务 key 查询确认。详见 [`../cross-context.md`](../cross-context.md) §4.2
- UoW 内修改既有聚合时仍优先调用 tx-bound `Update{Aggregate}`;它复用外层 `*sql.Tx` 和 pending collector,不自己开启嵌套事务。**不支持嵌套 `RunInTx`**;需要共享事务的写操作必须放进同一个顶层 `RunInTx` 闭包

### 3. 跨 BC 协作

- **绝不**跨 BC 事务,**绝不**分布式事务
- 以 **Integration Event** 解耦跨 BC 协作;**当前为 best-effort 投递,不承诺可靠最终一致性**(可靠保证需引入 Outbox 或等效机制):
  - 本 BC 内:单聚合 `Save` / `Update{Aggregate}` commit 成功后,或 UoW commit 成功后,统一 best-effort 发布 integration event 到 eventbus
  - 跨 BC:订阅方在 `ports/eventsub/` 接收并幂等处理
- 当前单进程 in-memory eventbus 下,commit 后在同一进程同步 publish,属于 **best-effort 投递**——commit 后 publish 任意失败(进程崩溃、eventbus 已关闭、handler panic 等)均可能丢事件,本期可接受但**不等于可靠投递保证**(Translate / 序列化失败在 commit 前,会回滚)

### 3.1 订阅方幂等

- 任何 integration event 订阅方**必须幂等**
- **本期**:幂等由 **app/command 的业务唯一键**保证(如支付以 `order_id` 唯一约束);`ports/eventsub/` handler 仅 decode → 委托 app,**不**引入 `processed_events` 表
- 发布方仍生成 **envelope.event_id**,但本期语义是**投递尝试 ID**,仅用于日志追踪与排障;未来 broker dedup 所需的稳定 `event_id` 必须在 outbox 中持久化生成。详见 [`../cross-context.md`](../cross-context.md) §3.3、§3.5

**code review 检查项**:每个 eventsub handler 所调用的 app command,必须能指出其业务幂等键及实现方式。

**演进**:接入外部 broker(at-least-once 投递)、多实例并发消费、或需要统一 poison message 停损时,先由 Outbox 持久化生成稳定 `event_id`,再在 `ports/eventsub/` 引入 `processed_events` 幂等占位表(以该稳定 `event_id` 为 dedup 键),与业务幂等形成双层防线——届时另写 ADR 或补全 cross-context §3.5。

## 替代方案

### 方案 A:2PC / XA 分布式事务

- **优点**:语义最强
- **缺点**:运维复杂、可用性差、无法跨异构存储、未来拆服务时也走不动
- **放弃原因**:在 modular monolith 阶段不值得,在 microservices 阶段更不可行

### 方案 B:Outbox 模式(事务内写 outbox 表 + 异步 dispatcher 发布)

- **优点**:保证"业务写 + 事件发布意向"的原子性,即使进程崩溃也不丢事件
- **缺点**:每个 BC 一张 outbox 表 + dispatcher 进程/协程;投递有延迟(取决于轮询周期)
- **本期不采用的原因**:单进程 in-memory eventbus 下,事务提交后直接 publish 属于 best-effort 投递(commit 后任意失败均可能丢事件),当前阶段可接受这一风险窗口;Outbox 引入的复杂度不值得。**演进到外部 broker 时再引入**

### 方案 C:Saga 编排所有跨 BC 协作

- **优点**:对复杂工作流表现力强
- **缺点**:Saga 引擎、补偿动作、状态管理都是额外复杂度
- **放弃原因**:本期默认用 Integration Event,Saga 仅在确实需要长事务编排时引入,届时新写 ADR

## 后果

### 正向影响

- 单聚合写极简,大多数业务用例落在这一类
- 跨 BC 通过 Integration Event 解耦
- 发布侧实现简单,无需 outbox 表和 dispatcher 进程;订阅侧本期仅需保证 app/command 业务幂等,无需 `processed_events` 表

### 负向影响 / 代价

- 订阅方需要做幂等
- commit 后 publish 任意失败(进程崩溃、eventbus 已关闭、handler panic 等)均可能丢事件;这是 **best-effort** 而非可靠投递,需要 Outbox 或等效机制才能升级为 at-least-once 保证
- 演进到外部 broker 时需要引入 Outbox 或等效机制

### 后续需要做的事

- [ ] `platform/eventbus` 抽象出"进程内 fan-out → 未来可换 Kafka"的接口
- [ ] 演进到外部 broker 时:评估并引入 Outbox 模式(新 ADR)
- [ ] 演进到外部 broker / 多实例时:评估并引入 `processed_events` 幂等占位表(新 ADR 或补全 cross-context §3.5)

## 相关文档

- [`../cross-context.md`](../cross-context.md) §3、§4
- [`../anti-patterns.md`](../anti-patterns.md) §7
- [ADR-0005 Domain Event vs Integration Event](./0005-events-internal-vs-integration.md)
