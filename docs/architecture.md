# 后端架构

当前实现是一套 **Go 标准库优先的模块化单体 MVP**：

```text
HTTP / WebSocket / SSE
        ↓
game / records / learning / analysis
        ↓
xiangqi rules + builtin alpha-beta engine
        ↓
memory repositories (development)
```

## 已实现

- `internal/domain/xiangqi`：坐标、FEN、规则、合法着、将军/终局、稳定 Zobrist 哈希；
- `internal/engine/builtin`：可取消、限时、限深、限节点的迭代加深 Alpha-Beta；
- `internal/game`：权威对局、乐观版本、幂等、AI 异步写回、悔棋取消、事件历史；
- `internal/records`：受大小限制的坐标棋谱/项目 JSON 导入、逐着验证和去重；
- `internal/learning`：从真实导入棋谱生成不可变局面着法统计版本；
- `internal/analysis`：异步、有限预算的候选着复盘；
- `internal/transport/httpapi`：REST、WebSocket，同时提供 SSE 降级；
- `migrations`：MySQL 8 生产数据模型基线。

## 权威性

前端提交的当前轮次、FEN、结果和合法性都不可信。落子只接受 ICCS 字符串和
`expectedMatchVersion`，服务端从保存的 FEN 重新解析并验证。AI 结果写回前再次验证
版本、轮次和合法性；迟到结果直接丢弃。

## 当前持久化边界

默认 `XIANGQI_DATA_MODE=memory`，便于无需基础设施直接联调。进程重启会丢失数据，
因此当前 readiness 会明确报告内存模式。`migrations/0001_initial.sql` 已定义生产表、
唯一约束和查询索引，但 MySQL/Redis 适配器尚未冒充为已完成。

Redis 只允许承载可重建任务和短期协调，不能成为对局唯一事实来源。

## Worker

内存模式的异步任务由 API 进程内受控 goroutine 执行，`cmd/worker` 提供独立进程和
健康端点，等待生产持久化队列适配器接入。这样不会假装两个独立进程可以共享内存队列。

