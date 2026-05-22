# ADR-0003: Repository 接口归 domain,services 接口归 app

- **状态**:Accepted
- **日期**:2026-05-20
- **决策者**:架构组

## 背景

业务层会依赖两类外部能力:

1. **聚合的持久化**:order 的存取
2. **跨 BC 与第三方能力**:查商品、调支付网关、发短信

两类接口在 Go 中都用 interface 表达,但**归属层应当不同**。如果统一放在 app 层,domain 就会需要 app 帮它做持久化(语义不通);如果统一放在 domain,domain 就要知道"支付网关"这种纯应用编排概念。

需要明确两类接口的层归属。

## 决策

- **Repository 接口** → 放在 `domain/{aggregate}/repository.go`
- **外部能力接口**(其他 BC、第三方服务、跨切能力如时钟、ID 生成器) → 放在 `app/command/services.go`(或 `app/query/services.go`)

### 判断标准

| 接口语义 | 归属 |
|---|---|
| "这个聚合怎么落地/恢复" | domain Repository |
| "用例编排时,需要协调谁来做事" | app services |

## 替代方案

### 方案 A:Repository 接口放在 app

- **优点**:接口与使用方同层,Go 社区常见
- **缺点**:domain 失去对持久化契约的所有权,聚合内部结构变化时,接口由 app 维护、domain 维护、还是两边商量? 容易扯皮
- **放弃原因**:Repository 是聚合契约,所有权应属 domain

### 方案 B:全部接口都放在 domain

- **优点**:domain 自描述完整
- **缺点**:domain 出现 `PaymentGateway`、`EmailSender` 等纯应用编排概念,违反 domain 纯净性
- **放弃原因**:破坏分层语义

### 方案 C:两类接口都放在调用方就近(handler 旁)

- **优点**:就近定义,可读性好
- **缺点**:Repository 在每个 handler 重复;且当一个聚合的 Repository 被多个 handler 共享时归属不清
- **放弃原因**:接口应该有明确的"所有者"

## 后果

### 正向影响

- domain 自洽:聚合 + 持久化契约在同一目录
- app 自洽:用例编排的依赖契约就近可见
- 拆微服务时,把 `services.go` 中某个接口的实现从"进程内 adapter" 换成"远程 client"即可,domain 不动

### 负向影响 / 代价

- 新人需要记住"两类接口两个位置",code review 需要主动检查
- 偶尔会争论某个能力算"领域持久化"还是"外部能力"(如 audit log)

### 命名与放置规则

- **Repository**:`domain/{aggregate}/repository.go`,接口名就叫 `Repository`,通过包名区分(`order.Repository`、`product.Repository`)
- **外部能力**:`app/command/services.go`,接口名按调用方视角命名(`ProductsService`、`PaymentGateway`),**不**叫 `Repository`

### 后续需要做的事

- [ ] 在 [`../anti-patterns.md`](../anti-patterns.md) 中记录"Repository 接口在 adapter 层定义"是反例
- [ ] code review checklist 加入"接口位置是否符合 ADR-0003"

## 相关文档

- [`../architecture.md`](../architecture.md) §4.3、§4.4
- [`../naming.md`](../naming.md) §4
- [ADR-0001 分层](./0001-layering.md)
