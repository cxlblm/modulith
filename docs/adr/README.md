# Architecture Decision Records (ADR)

ADR 是**为什么这么做**的记录。架构文档(architecture.md 等)说**怎么做**;ADR 解释**当时为什么做这个决策、考虑了哪些替代方案、放弃了什么**。

## 索引

| 编号 | 标题 | 状态 |
|---|---|---|
| [0001](./0001-layering.md) | 采用 DDD Lite 四层分层(domain / app / adapters / ports) | Accepted |
| [0002](./0002-module-contract.md) | 跨 BC 通信采用 Module Contract 模式 | Accepted |
| [0003](./0003-repository-vs-services.md) | Repository 接口归 domain,services 接口归 app | Accepted |
| [0004](./0004-cqrs-readmodel.md) | 显式 CQRS,query 走独立 ReadModel | Accepted |
| [0005](./0005-events-internal-vs-integration.md) | 区分 Domain Event 与 Integration Event | Accepted |
| [0006](./0006-transaction-and-outbox.md) | 事务边界与事件发布 | Accepted |
| [0007](./0007-bc-module-factory.md) | BC 自带 Module 工厂,bootstrap 只做进程级编排 | Accepted |
| [0008](./0008-pricing-bc-lock-pricing.md) | Pricing BC 负责下单锁价与优惠分摊 | Accepted |

## 何时写 ADR

下列情况**必须**写 ADR(在 PR 中附 ADR 文件):

- 新增、修改、推翻分层规则或 import 规则
- 新增跨 BC 通信方式(超出 [cross-context.md](../cross-context.md) 第 1 节列举的三种)
- 修改 Module Contract 接口语义(非新增字段的变更)
- 选择新的核心基础设施(数据库、消息中间件、RPC 框架)
- `shared/` 或 `platform/` 新增组件
- 任何"破例"行为

下列情况**不需要**写 ADR:

- 业务功能的常规迭代
- 已有 ADR 范围内的实现细节
- bug 修复

## 流程

1. 复制 [`template.md`](./template.md) 到 `00XX-<short-title>.md`,编号顺延
2. 状态先写 `Proposed`
3. PR 中讨论,共识后改为 `Accepted`
4. 被后续 ADR 推翻时,改为 `Superseded by ADR-XXXX`,**不要删除文件**

## 命名

- `00XX-<short-kebab-title>.md`
- 编号一旦分配,**永远不复用**(被推翻的 ADR 也保留编号)
