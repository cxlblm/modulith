# ADR-0005: 区分 Domain Event 与 Integration Event

- **状态**:Accepted
- **日期**:2026-05-20
- **决策者**:架构组

## 背景

事件机制在两个维度都有用:

- **BC 内部**:聚合产生事件,app 层据此触发同 BC 内的副作用
- **跨 BC**:order 完成 → payments 创建支付

如果两类事件用同一份定义、走同一个总线,会出现:

- domain 内部字段被跨 BC 暴露
- domain event 名字 / 字段一改,所有订阅方破坏
- 跨 BC 契约的版本管理无处安放

需要在概念与目录上分开。

## 决策

**两类事件,两个目录**:

| 维度 | Domain Event | Integration Event |
|---|---|---|
| 目录 | `domain/{aggregate}/events.go` | `ports/module/events.go` |
| 受众 | 本 BC 内部 | 其他 BC |
| 稳定性 | 可随 domain 演进 | **稳定契约**,字段变更需 deprecation |
| 字段 | 可含聚合内部细节 | 字段精简,只含跨 BC 必要信息 |
| 命名 | 过去式:`OrderPlaced` | 过去式 + 版本:`OrderPlacedV1` |
| 产生 | 聚合业务方法返回 | repository `Save` / `Update{Aggregate}` 内基于 domain event 翻译生成 |
| 投递 | 事务内 `PeekEvents()` + `Translate`(不 drain) | commit 后 publish,成功后 `ClearEvents()` |

### 翻译职责

- 聚合业务方法记录 `domain.OrderPlaced`(BC 内部事件),由 `PeekEvents()` / `ClearEvents()` 访问(见 [`../cross-context.md`](../cross-context.md) §3.2)
- **`app/command` 完全不感知 Integration Event**,只调 `repo.Save(ctx, aggregate)` 或 `repo.Update{Aggregate}(ctx, id, actor, updateFn)`
- **翻译执行点是 `adapters/{persistence}/{aggregate}_repository.go`**(repository 是 adapter,合法 import 自己 BC 的 `domain` + `ports/module`):
  1. **事务内**写聚合表(`Update{Aggregate}` 先 `SELECT ... FOR UPDATE` 加载并锁定聚合,再执行 update closure)
  2. **事务内**`PeekEvents()`(不 drain) → `Translate` + 序列化校验
  3. **commit 成功后**生成本次投递尝试的 `event_id` → eventbus 发布
  4. **全部 publish 成功后**`ClearEvents()`

> **事件生命周期与重试**:commit **前**失败,事务回滚,aggregate 仍持有事件,**可**用同一实例重试 `Save`,或重试整个 `Update{Aggregate}` command。`Commit` 返回 error 时结果未知,不 publish,上层需按业务 key 查询确认。commit **成功后**任意失败(含 publish)或 **`ClearEvents` 后**,**禁止**复用同一 aggregate 实例再次 `Save`,也禁止复用 update closure 内拿到的 aggregate 指针。若需可靠投递,引入 Outbox(事务内写 outbox,commit 后 dispatcher 发布)。UoW 多聚合规则见 [`../cross-context.md`](../cross-context.md) §4.2。
- **翻译规则的定义放在 `ports/module/translate.go`**(契约层的所有权)
- 其他 BC 的 `ports/eventsub/` 接收 `eventbus.Envelope`(含 event_id + payload),decode 为 Integration Event DTO 并委托已幂等的 app/command 处理(幂等策略见 [`../cross-context.md`](../cross-context.md) §3.5)

#### 为什么 app/command 不做翻译

`app` 不允许 import `ports/`(分层约束,详见 [`../architecture.md`](../architecture.md) §5)。若把翻译逻辑放在 app,会破坏分层。把翻译下沉到 repository:

- repository 是 adapter,合法 import `ports/module`
- 翻译与发布自然跟随 `Save` / `Update{Aggregate}` 动作,调用方无需额外步骤
- app 保持 transport-agnostic,新增 Integration Event 时 app 完全无感

详见 [`../cross-context.md`](../cross-context.md) §3。

## 替代方案

### 方案 A:统一一种事件,所有人共享

- **优点**:零模板
- **缺点**:domain 内部字段被外部依赖,演进瘫痪
- **放弃原因**:破坏 BC 边界

### 方案 B:跨 BC 直接用 Domain Event,只是放在跨 BC 总线

- **优点**:省一次翻译
- **缺点**:domain event 没有版本,字段名一改全网炸;且必然带上聚合内部细节
- **放弃原因**:Integration Event 是契约,Domain Event 是实现细节,概念不同

### 方案 C:跨 BC 只走 RPC,不用事件

- **优点**:逻辑直观,链路短
- **缺点**:同步耦合,故障传播;复杂工作流(如下单 → 支付 → 发货)同步链路过长
- **放弃原因**:某些场景必须异步(性能、解耦、可重试)

## 后果

### 正向影响

- domain 演进自由,只要 `ports/module/translate.go` 跟上
- Integration Event 是稳定契约,可以放心做版本演进
- 拆服务时,Integration Event 直接对应 Kafka 消息,改的是发布通道而非定义
- app/command 永远只调 repository 的 `Save` / `Update{Aggregate}` 等聚合持久化方法,不感知事件 transport,新增 Integration Event 时 app 零变更

### 负向影响 / 代价

- 同一业务事实在 BC 内外有两份定义,需要在 `ports/module/translate.go` 集中维护翻译
- Integration Event 版本管理需要纪律(deprecation 流程)
- 跨 BC 事件订阅方需保证 app/command 业务幂等(见 ADR-0006、cross-context §3.5)

### Integration Event 版本演进规则

- **新增字段**:直接加,旧订阅方忽略,无需新版本
- **删字段 / 改语义 / 改字段类型**:发新版本(`V2`),保留 `V1` 直到所有订阅方迁移完成
- 不允许"原地修改"已发布的事件契约
- 发布方负责 V1 → V2 的并行发布期

### 后续需要做的事

- [ ] 提供 Integration Event 的脚手架与 deprecation 流程文档
- [ ] 演进到外部 broker / 多实例时:评估是否引入 `processed_events` 幂等占位表(见 ADR-0006)
- [ ] 演进到外部 broker 时,评估是否引入 Outbox 保证原子性

## 相关文档

- [`../cross-context.md`](../cross-context.md) §3
- [`../anti-patterns.md`](../anti-patterns.md) §2.5、§2.6
- [`../naming.md`](../naming.md) §7
- [ADR-0002 Module Contract](./0002-module-contract.md)
- [ADR-0006 事务边界与事件发布](./0006-transaction-and-outbox.md)
