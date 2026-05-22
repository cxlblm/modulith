# 架构文档

本目录是项目架构的权威说明。新人 onboarding 按下面顺序阅读即可。

## 阅读顺序

1. [`architecture.md`](./architecture.md) — 核心架构:分层、目录结构、各层职责、import 规则
2. [`cross-context.md`](./cross-context.md) — 跨 Bounded Context 通信:Module Contract、Integration Event
3. [`naming.md`](./naming.md) — 命名规约(包、文件、类型、接口、错误)
4. [`testing.md`](./testing.md) — 测试组织与各层测试策略
5. [`anti-patterns.md`](./anti-patterns.md) — 反例集合,code review 自检清单
6. [`adr/`](./adr/README.md) — 架构决策记录(ADR),解释"为什么这么做"

## 一句话定位

本项目是一个 **Go Modular Monolith**,采用:

- **DDD Lite** 分层(domain / app / adapters / ports / bootstrap)
- **Module Contract** 模式做跨 BC 通信(参考 [Three Dots Labs](https://threedots.tech/post/microservices-or-monolith-its-detail/))
- **CQRS 显式分离**(command / query 在 app 层分两个子包)
- **演进友好**:任何一个 Bounded Context 都能在不改动 app / domain 层的前提下被拆为独立服务

## 文档纪律

- 架构决策一律走 [ADR](./adr/README.md),不在群里口头拍板
- 修改本目录的 PR 需要至少一位架构 owner approve
- 反例(anti-patterns)只增不删:历史上踩过的坑要被记下来
