# 项目架构

棋境是一套 Vue 3 前端与 Go 后端组成的模块化单体。当前重构目标是让业务代码只依赖
稳定边界，页面和 HTTP Handler 不再承担跨领域职责，同时保持现有路由、接口和交互行为。

## 总体调用链

```text
Browser
  |
  | REST + WebSocket/SSE
  v
apps/web
  View
    -> Store / Composable
    -> Domain API module
    -> Common HTTP / stream client
  |
  v
cmd/api
  -> transport/httpapi
  -> application services
  -> repository contracts
  -> domain rules / engine
```

## 前端架构

```text
apps/web/src
├─ api
│  ├─ client.ts       通用 REST 客户端、错误模型
│  ├─ stream.ts       对局 WebSocket/SSE 通道
│  ├─ contracts.ts    后端 DTO 与接口契约
│  ├─ matches.ts      对局、权威合法着法 API
│  ├─ records.ts      棋谱 API
│  ├─ learning.ts     学习任务和版本 API
│  ├─ analysis.ts     复盘任务和结果 API
│  └─ system.ts       难度、引擎健康等系统 API
├─ components         可复用界面组件
├─ composables        可复用交互和动画状态
├─ stores             Pinia 页面状态与业务流程编排
├─ utils              无副作用的棋盘、日期、结果计算
├─ views              路由页面与展示组合
├─ router             路由定义
├─ types              前端领域类型
└─ styles             全局设计令牌和样式
```

### 前端依赖规则

1. View 可以依赖 Store、Composable、Utils 和 API 查询模块。
2. Store 通过领域 API 模块访问后端，不直接拼接 URL、幂等请求头或 multipart 数据。
3. `api/client.ts` 只处理通用 HTTP 行为，不包含具体业务路径。
4. `api/contracts.ts` 描述后端 DTO；棋盘交互类型放在 `types/xiangqi.ts`。
5. 可重复、无副作用的展示计算放在 `utils`，避免多个页面各自解释胜负和日期。
6. 实时对局状态以服务端快照和事件版本为准，页面不能自行推导权威 FEN。
7. 生产页面不得注入 Demo、Mock 或静态业务兜底；接口失败必须展示明确错误状态。

### 对局状态流

```text
MatchView
  -> match Store
  -> matches API / legal moves API / match stream
  -> applySnapshot / version guard / move deduplication / cursor recovery
  -> XiangqiBoard
```

`match` Store 保留以下一致性约束：

- 忽略低于当前版本的旧事件或快照；
- REST 快照和实时事件按 `ply` 去重；
- 合法落点只读取服务端 `Position.LegalMoves()` 的查询结果；
- 合法着法按对局 ID、版本和起点缓存，版本变化后立即失效；
- `fenAfter` 决定棋盘局面和行棋方；
- WebSocket 重连携带最后成功消费的 `eventId`；
- 事件版本跳跃时停止应用事件，并重新获取完整快照；
- 请求失败恢复交互状态，并通知棋盘播放拒绝落子反馈；
- 棋盘动画只消费状态变化，不参与合法性判断。

## 后端架构

```text
cmd/api
  composition root
  |
  v
internal/transport/httpapi
  server.go       服务组装与路由
  matches.go      对局 Handler、实时流
  records.go      棋谱 Handler、上传解析
  learning.go     学习 Handler
  analysis.go     复盘 Handler
  system.go       健康、难度、许可证
  response.go     JSON、错误、中间件、SSE
  websocket.go    WebSocket 协议
  |
  v
internal/game | records | learning | analysis
  application services
  |
  v
internal/domain/xiangqi | internal/engine
  rules and search
```

### 后端依赖规则

1. `cmd/api` 是依赖组装入口，负责选择实现并注入服务。
2. HTTP Handler 只做协议转换、输入解码、状态码和错误映射。
3. Application Service 负责编排用例，不依赖具体 HTTP 类型。
4. `game.Service` 依赖 `game.Repository` 接口，不依赖内存仓库实现。
5. 中国象棋合法性、轮次、终局和 FEN 由 `domain/xiangqi` 统一判定。
6. 引擎通过 `engine.Engine` 接口接入，业务层不绑定具体搜索实现。

### 服务端权威性

前端提交的当前轮次、FEN、结果和合法性均不可信。落子接口只接受 ICCS 着法和
`expectedMatchVersion`，服务端从保存的权威 FEN 重新解析并验证。

AI 搜索结果写回前会再次验证版本、轮次和合法性；迟到结果直接丢弃。创建对局、落子、
悔棋、认输和求和请求使用幂等键，避免重试造成重复状态变更。

## 持久化现状

当前 `cmd/api` 注入 `game.MemoryRepository`，因此：

- 对局数据随 API 进程重启而丢失；
- readiness 的 `authoritativeStore` 会返回实际仓库名称 `memory`；
- `XIANGQI_DATA_MODE=mysql` 当前会连接数据库并执行 migration，但尚未注入 MySQL
  对局仓库；
- `migrations/0001_initial.sql` 已提供 MySQL 8 数据模型基线；
- Redis 和独立 worker 队列适配器尚未完成，不能作为权威数据来源。

后续实现 MySQL 对局仓库时，应新增独立 adapter 并实现 `game.Repository`，然后只在
`cmd/api` 更换注入，不修改 `game.Service` 或 HTTP Handler。

## 主要请求路径

### 创建和进行对局

```text
NewGameView
  -> match Store
  -> api/matches.createMatch
  -> POST /api/v1/matches
  -> httpapi.matches
  -> game.Service
  -> game.Repository + xiangqi rules + engine
```

### 查询合法落点

```text
XiangqiBoard
  -> match Store versioned cache
  -> GET /api/v1/matches/{id}/legal-moves?from=a3
  -> game.Service.LegalMoves
  -> domain/xiangqi.Position.LegalMoves
```

### 棋谱学习

```text
RecordsView -> records Store -> records API -> records.Service
LearningView -> learning Store -> learning API -> learning.Service
```

### 复盘分析

```text
AnalysisView
  -> analysis Store
  -> analysis API
  -> analysis.Service
  -> game.Service + engine.Engine
```

## 扩展约定

- 新增后端领域接口时，先在对应的 `httpapi/<domain>.go` 增加 Handler，再扩展服务用例。
- 新增前端接口时，先加入对应领域 API 文件，再由 Store 或 View 调用。
- 不在 View 中复制胜负、状态或时间计算；优先复用或扩展 `utils`。
- 不把数据库、请求对象或前端 DTO 引入 `domain/xiangqi`。
- 大范围样式拆分应单独进行，并配合页面截图和响应式回归验证。

## 验证命令

```powershell
go test ./...
go vet ./...
go build ./cmd/api ./cmd/worker ./cmd/tools/perft ./cmd/tools/enginebench

npm.cmd run typecheck:web
npm.cmd run lint:web
npm.cmd run test:web
npm.cmd run build:web
```
