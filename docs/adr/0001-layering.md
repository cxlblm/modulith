# ADR-0001: 采用 DDD Lite 四层分层

- **状态**:Accepted
- **日期**:2026-05-20
- **决策者**:架构组

## 背景

项目是一个 Go Modular Monolith,需要在以下约束之间取得平衡:

- 业务复杂度持续增长,纯 MVC / 三层架构难以约束业务逻辑泄漏
- 团队不全是 DDD 重度用户,需要避免完整 DDD 的概念过载(Domain Service / Specification / Anti-corruption Layer 全套)
- 单体起步,但**必须保留未来拆为微服务的可能性**
- 需要架构纪律可被 linter / CI 强制执行

## 决策

采用 **DDD Lite 四层分层**:

```
ports / adapters  → app  → domain
```

- **domain**:聚合、Entity、VO、Repository 接口、Domain Event、领域错误。零外部依赖。
- **app**:用例编排(command / query),依赖 domain 与"外部能力接口"。CQRS 显式分离。
- **ports**:入站端口(HTTP / WS / Stream / EventSub / Module),把外部事件翻译为 app 调用。
- **adapters**:出站适配器,实现 domain Repository 与 app services 接口。

**铁规则:内层永远不 import 外层。**

详见 [`../architecture.md`](../architecture.md)。

## 替代方案

### 方案 A:Standard Go Project Layout(`pkg/`、`internal/`、`api/`)

- **优点**:社区熟悉,新人 onboarding 成本低
- **缺点**:只规定了目录,没规定层内职责与依赖方向,业务复杂后纪律会瓦解
- **放弃原因**:不解决核心问题(业务逻辑泄漏、跨模块依赖混乱)

### 方案 B:完整 DDD(含 Application Service / Domain Service / Anti-corruption Layer / Specification / Saga 等)

- **优点**:理论完备,大型领域覆盖度高
- **缺点**:概念多,新人学习曲线陡;过度抽象在中等规模项目变成负担
- **放弃原因**:对当前业务复杂度过度设计

### 方案 C:Clean Architecture / Hexagonal(entity / usecase / interface / framework)

- **优点**:与 DDD Lite 思想一致
- **缺点**:Uncle Bob 原版的命名(`entities`、`use_cases`、`interface_adapters`)在 Go 社区不主流,且没规定跨 BC 边界
- **放弃原因**:DDD Lite 命名(`domain` / `app`)更贴近业务语言,且 Three Dots Labs 等 Go 社区团队已有成熟实践

## 后果

### 正向影响

- 业务逻辑集中在 domain 与 app,可独立测试
- 跨 BC 边界清晰(配合 [ADR-0002](./0002-module-contract.md))
- 拆微服务时 domain / app 无需改动

### 负向影响 / 代价

- 短期写法变重:每个用例都要走 command → handler → repository,比一把梭复杂
- 新人需要先读 [`../architecture.md`](../architecture.md) 和 [`../anti-patterns.md`](../anti-patterns.md) 才能上手
- 简单 CRUD 也要走完整流程,有一定模板代码

### 后续需要做的事

- [ ] 引入 `go-arch-lint` 或等价工具,在 CI 中强制 import 规则
- [ ] 新人 onboarding doc 中加入 [`../architecture.md`](../architecture.md)、[`../anti-patterns.md`](../anti-patterns.md) 阅读环节
- [ ] 提供 BC 脚手架生成器(可选)

## 相关文档

- [`../architecture.md`](../architecture.md)
- [`../anti-patterns.md`](../anti-patterns.md)
- [ADR-0002 Module Contract](./0002-module-contract.md)
