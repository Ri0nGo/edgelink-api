# 历史数据停止写入：审查与排查记录

审查日期：2026-09-08。结论基于当前工作区源码和提供的日志，未连接生产 Redis/MySQL，尚未确认生产中缺失的具体属性。

## 已确认的失败链路

`mysql_persistence.go` 读取 `HGET device:<device_id>:data <property_key>`。旧实现遇到某个字段返回 `redis.Nil` 时直接返回，丢弃同一批次的其他数据（默认每批 100 个设备属性）。最终结果为空时，`BatchSave` 向 `slice.SplitChunk` 传入分批大小 0，触发 `split size out of slice length`，没有执行 MySQL 插入。

任务每分钟第 2 秒执行，写入时间戳为当前分钟减一分钟。17:08:02 的失败对应 17:07:00 的历史记录；10:13:02 的失败与最后记录停在 10:11:00 相符。

前面的 `SELECT * FROM product` 是设备列表用于补充产品名称的查询。`GetDeviceList` 只读数据库和 Redis 状态，不更新持久化属性内存。HTTP 200 和这两条错误相邻不能证明存在因果关系。

## 本次已修复与新增日志

- 缺失字段或非法数值只跳过当前属性；同批和其他批次的有效值继续保存。
- Redis 实际读取错误携带设备和属性信息返回；调用方仍保存已读到的有效数据。
- 没有配置/没有可读数据时跳过保存，避免空列表分批。
- MySQL 插入失败向上传递，不再记录为整轮成功；成功批次记录行数和时间戳。
- 缺失项记录 `device_id`、`property_id`、`property_key`、`redis_key`、`sample_ts`。
- 每轮记录配置属性数、已保存行数、跳过属性数、读取错误状态和耗时。
- 配置通知记录发布/接收事件、订阅者数量、发布实例、事件时间以及旧/新属性 key。
- 启动日志记录 Redis 地址和 DB；服务层日志记录新增设备、物模型变更、持久化开关等来源。
- 通知附带可选 `publisher_id`、`publisher_db` 字段，兼容旧事件；收到不同 DB 的通知会告警。目前仅记录，不改变通知频道或拦截事件。

## 仍需处理或确认的配置同步问题

以下属于源码中发现的触发路径，不代表已证实是本次生产故障的起因。本次未重构这些配置同步行为。

| 路径 | 当前行为与影响 |
| --- | --- |
| `DeviceSvc.CreateDevice` | 所有模型属性默认开启持久化并立即通知；首次上报前 Redis 可能没有字段。 |
| `ThingModelSvc.CreateThingModelProp` | 新属性自动关联使用模型的设备并开启持久化，未上报字段会缺失。 |
| `ThingModelSvc.UpdateThingModelProp` | 将请求中的 key 通知给模型下所有设备，没有筛选属性关联及 `persistent`；已关闭持久化的属性也可能被重新加入内存。 |
| `ThingModelSvc.UpdateThingModel` | 整体编辑模型更新数据库属性，但不通知持久化内存，可能继续读取旧 key。 |
| `DeviceSvc.DeleteDevice` | 删除数据库记录但不通知；持久化器也未订阅设备事件。即使补发事件，还需补齐订阅，并正确携带/解析设备 ID（现有发布器未设置顶层 `Event.DeviceId`）。 |
| `DeviceSvc.DeleteDeviceProps` | 服务层删除属性关联不通知内存；当前该删除路由未注册。 |
| `DeviceRepo.UpdateDevicePropRef` | 使用 GORM 结构体 `Updates`，`Persistent=false` 是零值，不能依赖该写法关闭数据库中的开关；可能出现内存已移除、重启后又加载的现象。 |
| `device.event` 固定频道 | 多套程序共用同一 Redis 实例时，配置通知可能串入其他环境。Redis Pub/Sub 不按 DB 编号隔离，即使数据分别存在 DB 1、DB 2，也能收到同频道通知。是否存在多套程序需要核对部署。 |

Pub/Sub 行为参考：[Redis 官方文档：Database & Scoping](https://redis.io/docs/latest/develop/pubsub/#database--scoping)。新增 DB 告警需要发布端也运行含元数据的新版本；没有告警不等于排除旧发布端或同 DB 的其他环境。

## 部署后的定位步骤

1. 查看启动日志 `redis connect success`，确认地址和 `redis_db`；记录 `redis subcribe success` 的频道。
2. 等待下一分钟，搜索 `persistence redis field missing`，直接获取缺失设备 ID、属性 ID、属性 key。字段缺失日志按每轮输出，配置不修正会继续出现。
3. 向前查找同设备/属性的 `persistence config property upserted`，对照 `old_property_key` 与 `property_key`；结合 `device config event received` 和服务层 `config_source` 查明发布实例及变更来源。
4. 查询数据库期望持久化的全部属性（不是只查设备 4、5、7）：

```sql
SELECT d.id AS device_id, d.`key` AS device_key,
       p.id AS property_id, p.`key` AS property_key,
       r.persistent
FROM device_property_ref r
JOIN device d ON d.id = r.device_id
JOIN thing_model_property p ON p.id = r.property_id
WHERE r.persistent = 1
ORDER BY d.id, p.id;
```

5. 在相同 Redis 实例、相同 DB 中，对日志指向的字段执行只读检查。例如（ID、字段需替换为实际日志中的值）：

```redis
EXISTS device:4:data
HEXISTS device:4:data t
HGET device:4:data t
HKEYS device:4:data
```

6. 如果内存日志存在数据库查询结果之外的属性，优先核对未过滤持久化状态的属性更新、删除后内存残留、其他实例发来的通知。如果数据库和内存一致但 Redis 缺字段，核对设备实际上报字段及首次上报时间。
7. 通过 `persistence mysql batch saved`、`persistence cycle completed` 确认保存行数，并核对数据库最新时间戳。`batch save data failed` 表示至少一批失败，不能认为整轮全部回滚。

不要往缺失字段中补写虚假值。当前 Redis 保存的是最新值，本次修复不会自动补回中断期间的真实历史数据。

## 验证范围

持久化回归测试覆盖不同位置的缺失属性、跨批部分失败、空数据、非法数值、插入错误向上传递、部分读取结果仍保存以及配置旧/新 key 日志。通知测试通过 Redis 客户端 hook 验证实际发布事件的元数据，不访问网络。现有依赖本机 Redis/MQTT 的集成测试不作为本次无外部服务验证的一部分。
