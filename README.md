# Industrial Agent Memory

> 面向工业现场的长期记忆服务。把设备、产线、工艺、告警和处置结果沉淀成可召回、可审计、可合并的工程经验。  
> A long-term memory service for industrial agents. It turns equipment history, process context, alarms, and action outcomes into recallable, auditable, and maintainable operational experience.

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Status](https://img.shields.io/badge/status-early--stage-orange?style=flat-square)](#路线图--roadmap)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](#license)

工业智能体不能每次都从零开始。  
Industrial agents should not start from scratch every time.

同一台冲压机上个月刚因为轴承润滑不足出现过高频振动；某条产线在夜班切换物料后容易出现温度波动；某位维修工程师处理伺服报警时总会先检查接线端子。这些信息通常散落在维修记录、班组交接、聊天记录和个人经验里，很难被下一次诊断稳定使用。

A press machine may have shown high-frequency vibration last month due to poor bearing lubrication. A line may become unstable during night-shift material changeovers. A maintenance engineer may always check terminal wiring first when handling servo alarms. These details often live in work orders, shift handovers, chat records, or personal experience, and are rarely reused reliably in the next diagnosis.

`Industrial Agent Memory` 关注的不是聊天上下文，而是工业现场的经验资产。  
`Industrial Agent Memory` is not a chat-history store. It is designed for operational experience from the shop floor:

- 哪台设备发生过什么问题。 / Which equipment had which issue.
- 当时有哪些信号、告警、参数和工况。 / Which signals, alarms, parameters, and operating conditions were present.
- 采取了什么处置动作。 / Which action was taken.
- 结果是否有效。 / Whether the action worked.
- 下次遇到相似问题时应该优先召回哪些经验。 / Which experience should be recalled first when a similar case appears.

## 适合做什么 / Use Cases

这个项目可以作为工业智能体平台里的独立记忆层，也可以单独嵌入设备诊断、质量分析、能耗优化、巡检助手等系统。

This project can be used as the memory layer of an industrial agent platform, or embedded directly into equipment diagnosis, quality analysis, energy optimization, and inspection assistant systems.

典型场景：

- 设备诊断：召回同设备、同告警码、同振动/温度模式下的历史处置经验。 / Equipment diagnosis: recall historical actions from the same equipment, alarm code, vibration pattern, or temperature pattern.
- 维修助手：记录某类故障的有效维修动作，并在相似事件中优先提示。 / Maintenance assistant: record effective repair actions and surface them when a similar failure occurs.
- 工艺分析：沉淀参数调整、物料切换、班次变化对质量和节拍的影响。 / Process analysis: preserve the impact of parameter changes, material changeovers, and shift patterns.
- 告警治理：记录重复告警的真实原因，减少每次都重新排查。 / Alarm governance: record the real causes of repeated alarms instead of restarting investigation every time.
- 人员协同：保留工程师偏好、职责范围和审批习惯，但不把它们混进设备记忆。 / Team collaboration: keep engineer preferences, responsibilities, and approval habits separate from equipment memory.

## 不做什么 / Non-Goals

为了让边界清楚，当前内核刻意不做这些事。  
To keep the boundary clear, the core intentionally does not:

- 不内置大模型调用。 / Call LLMs by itself.
- 不绑定某个向量数据库。 / Bind to a specific vector database.
- 不替代 CMMS、MES、SCADA 或时序数据库。 / Replace CMMS, MES, SCADA, or time-series databases.
- 不把所有对话历史都塞进“记忆”。 / Treat all conversation history as memory.
- 不直接下发控制指令。 / Send control commands to field devices.

它只负责一件事：把经过筛选的工业经验变成结构化记忆，并在需要时可靠召回。  
It does one thing: convert selected industrial experience into structured memory and recall it reliably when needed.

## 核心设计 / Core Design

```text
Observation
  -> validate
  -> Memory
  -> Store
  -> Recall / Outcome / Consolidation / Forget
```

### 记忆不是一段文本 / Memory Is Not Just Text

每条记忆都包含明确的工业上下文。  
Each memory carries explicit industrial context:

```text
Memory
├── Scope         设备 / 产线 / 工艺 / 告警 / 用户 / 站点
│                 equipment / line / process / alarm / user / site
├── Kind          故障 / 维修 / 参数 / 规程 / 偏好 / 观察 / 决策
│                 incident / maintenance / parameter / procedure / preference / observation / decision
├── Content       可读描述 / human-readable description
├── Tags          主题标签 / topic tags
├── Signals       告警码、物料、班次、工况等结构化信号
│                 structured signals such as alarm code, material, shift, operating condition
├── Evidence      来源、记录编号、观察时间 / source, record reference, observed time
├── Confidence    置信度 / confidence
├── Outcome       是否被验证有效 / whether the memory was verified by outcome
├── Importance    重要性 / importance
└── TTL           可选过期时间 / optional expiration
```

### 召回不只看相似文本 / Recall Is More Than Text Similarity

工业现场里，“同设备”“同告警码”“同工艺段”往往比句子相似度更重要。当前召回评分综合了：

On the shop floor, "same equipment", "same alarm code", and "same process segment" are often more important than sentence similarity. The current recall score combines:

- Scope 匹配：同设备、同产线、同工艺对象优先。 / Scope match: same equipment, line, or process object is prioritized.
- Kind 匹配：诊断时优先找故障和维修记忆。 / Kind match: diagnosis can prefer failure and maintenance memories.
- Tags 匹配：如 `vibration`、`bearing`、`temperature`。 / Tags match: for example `vibration`, `bearing`, `temperature`.
- Signals 匹配：如 `alarm_code=vib-high`、`material=PA66`。 / Signals match: for example `alarm_code=vib-high`, `material=PA66`.
- Text 匹配：标题和内容的轻量词项重合。 / Text match: lightweight token overlap from title and content.
- Outcome 加权：被验证有效的记忆会更靠前。 / Outcome weight: verified-effective memories rank higher.
- Recency 衰减：旧记忆不会立刻失效，但权重会下降。 / Recency decay: old memories do not disappear immediately, but their weight decreases.

### 记忆需要被维护 / Memory Needs Maintenance

记忆层不能只增不减。项目内置了几个维护动作。  
A memory layer should not grow forever without maintenance. The project includes:

- `MarkOutcome`：标记某条记忆是否有效，影响后续排序。 / Mark whether a memory worked and influence future ranking.
- `ForgetExpired`：清理临时观察、班次备注等短期信息。 / Remove temporary observations, shift notes, and short-lived facts.
- `ConsolidationCandidates`：找出可能重复或高度相似的记忆，交给人工或后台任务合并。 / Find likely duplicate or highly similar memories for human review or background consolidation.

## 快速开始 / Quick Start

```bash
git clone https://github.com/fde-ai/industrial-agent-memory.git
cd industrial-agent-memory
go test ./...
go run ./cmd/iamemory
```

## 作为 Go 库使用 / Usage as a Go Library

```go
package main

import (
	"context"
	"fmt"

	"github.com/fde-ai/industrial-agent-memory/memory"
)

func main() {
	ctx := context.Background()
	service := memory.NewService(memory.NewInMemoryStore())

	item, _ := service.Remember(ctx, memory.Observation{
		Kind:    memory.KindKnownFailure,
		Scope:   memory.Scope{Type: memory.ScopeEquipment, ID: "press-01"},
		Title:   "振动升高伴随温度上升",
		Content: "优先检查轴承润滑、模具偏载和最近一次换模记录。",
		Tags:    []string{"vibration", "temperature", "bearing"},
		Signals: map[string]string{
			"alarm_code": "vib-high",
		},
		Confidence: memory.ConfidenceHigh,
		Importance: 0.82,
	})

	_, _ = service.MarkOutcome(ctx, item.ID, memory.OutcomeEffective)

	results, _ := service.Recall(ctx, memory.RecallQuery{
		Scope: memory.Scope{Type: memory.ScopeEquipment, ID: "press-01"},
		Text:  "vibration temperature bearing",
		Tags:  []string{"vibration"},
		Signals: map[string]string{
			"alarm_code": "vib-high",
		},
		Limit: 5,
	})

	for _, result := range results {
		fmt.Println(result.Score, result.Memory.Title, result.Reason)
	}
}
```

## 存储方式 / Storage

当前提供两种实现。  
Two implementations are currently included:

- `InMemoryStore`：适合单元测试、演示和嵌入式试验。 / Suitable for unit tests, demos, and embedded experiments.
- `JSONLStore`：适合本地开发、边缘网关轻量部署和可读审计。 / Suitable for local development, lightweight edge deployment, and readable audit trails.

后续计划增加。  
Planned storage and integration options:

- PostgreSQL / pgvector。
- Qdrant。
- Milvus。
- Mem0 adapter。
- OpenTelemetry trace。

## 和工业智能体平台的关系 / Platform Integration

推荐放在智能体运行链路中的这个位置。  
Recommended position in an industrial agent runtime:

```text
Alarm / WorkOrder / Operator Note
  -> Tool Gateway
  -> Agent Runtime
  -> Industrial Agent Memory
       ├── Recall before reasoning
       ├── Remember after action
       ├── Mark outcome after verification
       └── Consolidate during maintenance window
```

一个设备诊断智能体可以这样使用它。  
An equipment diagnosis agent can use it like this:

1. 接到告警后，先用设备 ID、告警码、关键指标召回历史经验。 / After receiving an alarm, recall historical experience using equipment ID, alarm code, and key signals.
2. 把召回结果作为诊断上下文，而不是直接相信模型猜测。 / Use recalled memories as diagnosis context instead of relying on model guesses alone.
3. 生成处置建议后，记录本次诊断过程。 / After generating an action plan, record the diagnosis process.
4. 工单关闭时，根据维修结果标记记忆是否有效。 / When the work order is closed, mark whether the memory was effective.
5. 定期合并重复经验，清理过期观察。 / Periodically merge duplicate memories and clean up expired observations.

## 项目结构 / Project Structure

```text
industrial-agent-memory/
├── cmd/iamemory/        最小 CLI 示例 / minimal CLI example
├── memory/
│   ├── types.go         领域模型 / domain model
│   ├── service.go       写入、召回、标记、遗忘、合并候选
│   │                     remember, recall, outcome marking, forgetting, consolidation candidates
│   ├── scoring.go       召回评分与重复判断 / recall scoring and duplicate detection
│   ├── store.go         Store 接口与内存实现 / Store interface and in-memory implementation
│   ├── jsonl_store.go   JSONL 持久化实现 / JSONL persistence
│   └── service_test.go  核心行为测试 / core behavior tests
└── go.mod
```

## 路线图 / Roadmap

短期 / Short term:

- 增加 HTTP API 服务。 / Add an HTTP API service.
- 增加 PostgreSQL 存储。 / Add PostgreSQL storage.
- 增加更细的 Scope 层级：工厂、车间、产线、工位、设备、部件。 / Add finer-grained scope hierarchy: site, workshop, line, station, equipment, component.
- 增加记忆合并接口，而不只是合并候选。 / Add a memory merge API, not only merge candidates.
- 增加记忆导入导出工具。 / Add memory import and export tools.

中期 / Mid term:

- 接入 pgvector / Qdrant / Milvus。 / Integrate pgvector / Qdrant / Milvus.
- 支持 Mem0 adapter。 / Support a Mem0 adapter.
- 支持 LangGraph 工具封装。 / Support LangGraph tool wrappers.
- 增加设备故障码、工艺参数、SOP 步骤的专用索引。 / Add dedicated indexes for equipment fault codes, process parameters, and SOP steps.
- 增加多租户和权限边界。 / Add multi-tenant and permission boundaries.

长期 / Long term:

- 支持边缘侧离线记忆与中心侧同步。 / Support offline edge memory and central synchronization.
- 支持记忆质量评分和人工复核队列。 / Support memory quality scoring and human review queues.
- 支持从工单、报警、巡检记录中抽取候选记忆。 / Extract candidate memories from work orders, alarms, and inspection records.
- 支持与工业知识图谱关联。 / Link memories with industrial knowledge graphs.

## 设计原则 / Design Principles

- 记忆要有对象：没有设备、产线、工艺或人的归属，就很难被正确使用。 / Memory needs an object: without equipment, line, process, or user ownership, it is hard to use correctly.
- 记忆要有来源：每条高价值记忆都应该能追溯到记录、工单或观察。 / Memory needs evidence: valuable memories should trace back to records, work orders, or observations.
- 记忆要有结果：有效和无效的经验都重要，但权重不能一样。 / Memory needs outcomes: effective and ineffective experience both matter, but should not carry the same weight.
- 记忆要能过期：班次备注、临时旁路、短期异常不应该永久影响判断。 / Memory needs expiration: shift notes, temporary bypasses, and short-term anomalies should not affect decisions forever.
- 记忆要能被合并：重复经验会降低召回质量。 / Memory needs consolidation: duplicate experience lowers recall quality.
- 记忆不等于知识库：手册、SOP、规范属于知识库；现场反复验证过的经验属于记忆。 / Memory is not a knowledge base: manuals, SOPs, and standards belong to knowledge; field-tested experience belongs to memory.

## License

Apache-2.0
