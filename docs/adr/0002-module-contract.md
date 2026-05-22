# ADR-0002: 跨 BC 通信采用 Module Contract 模式

- **状态**:Accepted
- **日期**:2026-05-20
- **决策者**:架构组

## 背景

在 Modular Monolith 中,Bounded Context 之间会有同步调用需求(如 order 创建时查询 product)。如果允许 BC 互相直接 import,modularity 名存实亡:

- A 改 B 的内部字段会破坏 A
- 拆服务时所有调用点都要重写
- 测试无法独立

需要一种**进程内调用**机制,既要解耦,又不付出"全部跨进程"的运行时开销。

## 决策

每个 BC 通过 `ports/module/` 包对外暴露**进程内调用契约**:

- `contract.go`:接口定义(如 `ShopModule`)+ 实现(内部调用本 BC 的 `app.Application`)
- `types.go`:transport DTO,与 domain entity 解耦
- `events.go`:对外发布的 Integration Event 定义

调用方在自己的 `app/command/services.go` 声明**最小接口**(只用到的方法),`adapters/{bc}client/` 同时持有发布方的大接口、实现调用方的小接口,在边界处做适配与错误翻译。

**强制约束**:**任何 BC 只允许 import 其他 BC 的 `ports/module/` 包**,不允许 import 对方的 domain / app / adapters。

详见 [`../cross-context.md`](../cross-context.md)。

## 替代方案

### 方案 A:全部走进程内 HTTP 自调

- **优点**:接口规约用 OpenAPI 描述,拆服务零改动
- **缺点**:进程内 HTTP 性能差、可观测性割裂、序列化/反序列化无意义开销
- **放弃原因**:为未来可能的拆服务,在 99% 的时间里付出 100% 的成本,不划算

### 方案 B:BC 之间直接 import,约定不 import 对方 domain 内部

- **优点**:零模板代码
- **缺点**:约定无法被工具强制,迟早被破坏;无 transport DTO 边界,domain 字段一改就传染
- **放弃原因**:架构纪律必须可被 CI 强制

### 方案 C:走 `internal/api/contracts/` 集中放契约,各 BC 都 import 这里

- **优点**:契约集中可见
- **缺点**:契约定义脱离 BC,后期维护权混乱;违反"BC 拥有自己的对外契约"
- **放弃原因**:契约的所有权应属于发布方 BC

### 方案 D:Three Dots Labs 原版 `interfaces/private/intraprocess/`

- **优点**:已被工业验证
- **缺点**:命名与本项目的 `ports/` 体系不一致(原版用 `interfaces/`)
- **采纳但改名**:本质相同,但放在 `ports/module/`,与 `ports/http`、`ports/stream` 体系一致——所有 ports 都是"把外部事件翻译为 app 调用"

## 后果

### 正向影响

- BC 边界由代码强制,而非靠人自觉
- 拆微服务时,只需把 `adapters/{bc}client/` 从"调 module 接口"换成"HTTP/gRPC 调用",app 层零变更
- 各 BC 的对外契约由发布方拥有,演进权清晰

### 负向影响 / 代价

- 同一用例的调用路径加长:HTTP handler 调 app 是一份代码,module contract 调 app 又是一份代码
- 跨 BC 调用需要在 adapter 处写 DTO 翻译
- 对接口稳定性要求高,演进需要 deprecation 流程

### 后续需要做的事

- [ ] 在 `import-boundary` 规则中加入"BC 之间只允许 import `ports/module/`"
- [ ] 准备 BC 脚手架,自动生成 `ports/module/` 三件套
- [ ] 制定 Module Contract 接口的 deprecation 流程文档

## 相关文档

- [`../cross-context.md`](../cross-context.md)
- [Three Dots Labs: Microservices or Monolith its detail?](https://threedots.tech/post/microservices-or-monolith-its-detail/)
- [ADR-0001 分层](./0001-layering.md)
- [ADR-0005 Domain Event vs Integration Event](./0005-events-internal-vs-integration.md)
