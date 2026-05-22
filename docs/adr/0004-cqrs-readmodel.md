# ADR-0004: 显式 CQRS,query 走独立 ReadModel

- **状态**:Accepted
- **日期**:2026-05-20
- **决策者**:架构组

## 背景

读侧与写侧的关注点本质不同:

- 写侧关心**业务规则**(不变量、状态机、事件)
- 读侧关心**展示与查询**(投影、join、聚合、分页)

如果读写复用同一个聚合,会出现:

- query 因为要展示多张表的数据,把不属于聚合的字段塞进聚合
- 聚合为了序列化给前端,丢失对内部字段的封装
- 性能优化(读库、缓存、物化视图)被聚合不变量绑架

需要在工程层面把读写显式分开。

## 决策

在 `app/` 层显式区分 command 与 query:

- `app/command/`:写用例,依赖 domain Repository 接口
- `app/query/`:读用例,依赖独立的 ReadModel 接口
- ReadModel 接口定义在 `app/query/read_model.go`,实现在 `adapters/{persistence}/{aggregate}_read_model.go`
- Query 返回 `app/query/types.go` 中定义的 DTO,**不返回 domain 对象**

`app/app.go` 用 `Application{Commands, Queries}` 容器把两侧聚合起来。

### 选项:何时不用独立 ReadModel

- 极简单的"按 ID 拿单聚合"查询,可以暂时复用 Repository
- 但**返回值仍必须是 DTO,不能直接吐出 domain 对象**
- 一旦该查询需要 join 或字段裁剪,立即拆出独立 ReadModel

## 替代方案

### 方案 A:不区分 command/query,handler 平铺

- **优点**:最简单
- **缺点**:读写关注点混在一起,query 需求会污染 domain
- **放弃原因**:中等以上规模项目反复证明会失控

### 方案 B:Event Sourcing + CQRS 重型版

- **优点**:事件溯源 + 投影,极强的演进能力
- **缺点**:复杂度爆炸,运维成本高,与团队水位不匹配
- **放弃原因**:本期过度设计

### 方案 C:Query 直接走 ORM 拿聚合再转 DTO

- **优点**:省一份 ReadModel 接口
- **缺点**:聚合 hydrate 成本(关联加载、事件回放)在读路径上无意义;聚合字段变化会破坏 query
- **放弃原因**:读路径与聚合生命周期解耦才能优化

## 后果

### 正向影响

- 读路径可独立优化:读库、缓存、物化视图、独立索引
- 写路径专注业务规则,不为展示让步
- 拆微服务时,读侧甚至可以走独立的 query 服务

### 负向影响 / 代价

- 接口数量 ~翻倍(每个聚合一对 Repository + ReadModel)
- 简单 CRUD 也要写两套
- 团队需要克制"用 Repository 顺手查一下"的冲动

### 跨 BC Query 的处理

跨 BC 的复合查询(如订单列表带商品名)默认走**本地物化视图**:

- order BC 在本地存一份精简的 product 物化视图
- 通过订阅 shop 的 Integration Event 更新物化视图
- query handler 在 order 自己的库上 join

不推荐 query handler 跨 BC 同步调用 module contract 做组合(除非数据量小且容忍延迟)。详见 [`../cross-context.md`](../cross-context.md) §5。

### 后续需要做的事

- [ ] 在 BC 脚手架中默认生成 `app/query/read_model.go` 与 `adapters/{persistence}/{aggregate}_read_model.go`
- [ ] 提供示例:跨 BC 物化视图的实现样板

## 相关文档

- [`../architecture.md`](../architecture.md) §4.4
- [`../cross-context.md`](../cross-context.md) §5
- [`../anti-patterns.md`](../anti-patterns.md) §5
- [ADR-0001 分层](./0001-layering.md)
