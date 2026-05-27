# 0008. Pricing BC 负责下单锁价与优惠分摊

状态: Accepted

## 背景

下单流程需要计算商品原始价、行级分摊价和参与的优惠活动。价格和优惠规则会独立演进,并且未来可能被订单外的结算、预览或售后场景复用。

如果把算费逻辑放入 order domain,order 聚合将需要感知商品价格、优惠活动和跨 BC 查询,破坏 domain 只依赖内层业务规则的约束。如果只把结果散落在 order app 中,价格快照的一致性也缺少明确边界。

## 决策

新增独立 `pricing` BC 作为算费中心。

- Pricing BC 通过 Module Contract 暴露 `CalculateOrderPricing`,接收用户和商品行,返回锁价快照。
- Order BC 在 `app/command.PlaceOrder` 中调用自己定义的 `PricingService` 小接口,由 `order/adapters/pricingclient` 适配 `pricing/ports/module`。
- Order domain 不调用 Pricing,只接收 Pricing 返回的价格快照,校验订单自身的不变量并保存不可变快照。
- v1 只支持订单级满减活动,命中一个最优活动,优惠按商品行原始小计比例分摊。

## 后果

- 下单保存后的订单价格不受后续商品价格或活动变化影响。
- Pricing 规则可以独立演进,不会污染 Order 聚合。
- Order app command 仍然负责用例编排:用户资格、地址、算费、预占库存、保存订单。
- 如果后续需要叠加活动、优惠券占用或活动名额扣减,需要新增 ADR 明确事务和一致性语义。
