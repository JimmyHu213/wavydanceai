# ComfyUI Playground — 执行计划

> 目标：让 wavydanceai 客户在管理后台直接玩多模态 AI 工作流（text → image / video / audio / etc.），底层走自部署 ComfyUI + 你 gateway 的统一计费。
>
> 创建日期：2026-06-06
> 关联：`docs/BACKEND_GAP_VS_NEW_API.md` §6 多模态/异步任务

---

## 业务上下文（必读）

本项目是 **白标 SaaS** —— 每个客户 = 独立 wavydanceai 部署 + 独立 ComfyUI 实例 + 独立 GPU。客户付钱：你给他们交付整套部署。

**法律基线**（重要）：
- ComfyUI 是 **GPL-3** —— 必须独立部署、独立进程、独立源码，**绝不 fork 进我们的 repo**
- 我们的代码通过 HTTP API 调用 ComfyUI = mere aggregation，不传染我们的 license
- **永远不要 import / vendor / bundle ComfyUI 源码**
- 自己想加新节点 = 写**独立的 custom_node 包**（MIT/Apache），通过 ComfyUI 插件机制加载

**模型 license 注意**（v0 必做）：
- FLUX.1-dev 非商用，要 FLUX.1-pro 才能商用
- SD 3 月收入 <$1M 才可商用
- Hunyuan Video 用户数 <1 亿才商用
- SD-XL / FLUX.1-schnell / Wan 2.1 / Mochi → 商用安全
- 给客户部署前必须 review 每个模型的 license

---

## 计费模型（已决策）

**统一 `User.Quota`**，**不另开 credit 系统**：

- 沿用现有 P0 充值通道（Stripe / E-Pay / Crypto）—— 用户充值进 `User.Quota`
- ComfyUI 工作流里每个节点的外部 API 调用都路由到 wavydanceai gateway
- gateway 按现有 `ModelRatio` 表扣 quota（跟 LLM relay 完全同一套机制）
- 本地 GPU 节点（SD/Flux/Wan）按 GPU-second × ratio 扣 quota
- 前端 UI 上叫 "Credits" / "余额"（i18n 字符串），后端字段保持 `quota`

**用户不绑定任何 upstream API key** —— 所有 key 在 admin 的 channel 配置里，对用户透明。

---

## 三阶段路线

### v0 — 评估期（research，无代码改动）

**目标**：亲手跑通 ComfyUI，识别 v1 要复刻的节点 + 锁定商用安全的模型清单 + 评估 GPU 成本。

**任务清单**：
1. 部署一个独立 ComfyUI instance（[决策：see 开放问题 Q1]）
2. 装 5-10 个常见 custom_nodes，尝试：
   - text → image（SD-XL、Flux Schnell）
   - text → video（Wan 2.1、Mochi）
   - image → video（同上）
   - LLM 节点（Claude / OpenAI / Anthropic 官方插件）
3. 记录每个节点的 input/output schema（v1 复刻参考）
4. 测每种生成的实际 GPU 时间 + 成本估算（GPU $/hour × seconds = 单价基准）
5. 审计：每个模型 license + 每个 custom_node license

**Deliverables**：
- `docs/COMFY_V0_FINDINGS.md`，至少包含：
  - 部署文档（用什么 GPU、装了什么、踩了什么坑）
  - 实测的节点清单（input/output/cost）
  - 商用安全模型白名单（with license 链接）
  - 商用安全 custom_node 白名单
  - GPU 成本表（每种生成 ~$X / sec）
  - 建议的 v1 节点列表（≤10 个）

**Gate to v1**：
- [ ] ≥3 种模型走通完整生成流程
- [ ] 至少一个 image + 一个 video 模型确认商用 license
- [ ] GPU 成本基线测出（可以反推定价）
- [ ] 决定 v1 是 iframe 还是 React Flow（基于使用体验判断）

**v0 不做**：
- 任何 wavydanceai 代码改动
- 任何 custom_node 开发
- 任何用户面 UI
- 生产部署

---

### v1 — MVP 集成（要写代码）

**目标**：客户能在 wavydanceai 管理后台用 playground，调 ComfyUI，从统一 quota 扣费。

**Architecture**：

```
[ Browser ]
    │  /console/playground (iframe OR 自写 React Flow)
    ▼
[ ComfyUI UI (GPL, 独立服务) ]
    │  /prompt + WS /ws
    ▼
[ ComfyUI Backend ]
    │  custom node 调外部 API 时 →
    ▼
[ wavydanceai gateway (Go) ]
    │  - 验证用户 session
    │  - 按 ModelRatio 扣 User.Quota
    │  - 路由到 channel 配置的 upstream
    ▼
[ Anthropic / Bytedance / xAI / 本地 GPU ]
```

**子项目**：

1. **`wavydanceai-comfy-nodes`**（独立 repo, MIT/Apache）
   - 5-10 个核心节点（按 v0 findings 选）
   - 每个节点：调 wavydanceai gateway 的 endpoint，传递用户 token
   - 文档：怎么装到 ComfyUI custom_nodes 目录

2. **wavydanceai backend**（本 repo）
   - 新 channel adapter（如果有新 upstream provider）
   - 新模型加入 `ModelRatio` 表
   - 新 endpoint（如果 v0 发现现有 `/v1/chat/completions` 不够用）
   - **不**新增 quota 系统、**不**新增 credit 概念

3. **wavydanceai frontend**（本 repo）
   - `/console/playground` 页面
   - 路径决定：
     - **iframe 方案**：一个 iframe + 自动登录 ComfyUI 的 SSO bridge
     - **React Flow 方案**：从 ComfyUI `/object_info` 拉节点定义，前端渲染
   - i18n 字符串："Credits balance" / "余额" 等

**Gate to v2**：
- [ ] ≥10 个客户跑过 playground 至少一次
- [ ] 计费准确率 100%（spot check 10 单）
- [ ] WS 进度推送稳定（24 小时不挂）
- [ ] 至少 1 个客户主动要求"我能自定义工作流吗"（再上 React Flow 才有用户驱动）

**v1 不做**：
- 工作流模板市场
- 多用户协作画布
- 自定义节点编辑器
- 计费层架构改造

---

### v2 — Brand-owned canvas（later）

**触发条件**：v1 跑稳 + 用户表达"想要自己拼工作流" + 你想完全控制 brand。

**核心**：用 React Flow 重写画布，彻底替换 iframe。

**预期工作**：
- React Flow 集成 `web/wavy/`
- 节点 UI 一对一对照 ComfyUI 节点（自动从 `/object_info` 生成）
- 工作流 JSON 跟 ComfyUI 原生格式兼容（用户可以导入导出）
- 工作流模板存 DB，按用户/分组共享

**完全不在本计划范围**，只是把它写下来防止 v1 阶段做出"以后没法升级"的决定。

---

## 开放问题（v0 启动前要决策）

### Q1: v0 ComfyUI 跑在哪？

| 选项 | 成本 | 速度 | 适合 |
|---|---|---|---|
| 本地 Mac M-series + MPS | $0 | 慢，video 几乎跑不动 | 只测 image |
| RunPod / Vast.ai GPU（按小时 ~$0.5-2/h）| 用了才付 | 快 | 短期评估 |
| Cloudflare GPU / Modal | recurring | 中 | 评估 + 顺便测 production hosting |
| 自买 server（4090 ~$2k） | capex | 快 | 长期客户 demo 用 |

**建议 v0 用 RunPod / Vast.ai 按小时**（成本最低、速度够、不留资产）

### Q2: v0 阶段评估哪些模型？

候选优先级（按"使用频率 × 商用安全度"）：
- [ ] SD-XL（image, 商用安全）
- [ ] FLUX.1-schnell（image, Apache 2.0）
- [ ] Wan 2.1（video, Apache 2.0）
- [ ] Mochi 1（video, Apache 2.0）
- [ ] Hunyuan Video（video, ⚠️ 用户数限制）
- [ ] FLUX.1-dev（image, ⚠️ 非商用 → 仅评估不部署）
- [ ] SD 3 medium（image, ⚠️ 收入限制）

[决策：圈定 v0 跑哪些]

### Q3: v1 UI 选 iframe 还是 React Flow？

**v0 结束前不要定**，让实际使用体验告诉你：
- iframe：1-2 周上线，brand 损失
- React Flow：4-6 周，完全 brand

### Q4: 谁来跑 v0？

- [ ] 你自己跑（最了解客户期望）
- [ ] 我帮你出部署脚本 + 测试 checklist，你执行

### Q5: v0 时间盒？

定一个上限（比如 2 周），防止"评估"无限期拖。到期不管多少 findings 都进入 v1 决策。

---

## 风险

| 风险 | 缓解 |
|---|---|
| **客户的 ComfyUI 部署需要 GPU 机器** —— 你的白标 SaaS 第一次卖硬件资源 | v0 时同步评估 GPU 供应链（RunPod B2B / Vast.ai partner / 自有 cluster）。v1 上线时 pricing 模型要含 GPU 成本 |
| **ComfyUI 版本升级可能破坏 custom_nodes** | wavydanceai-comfy-nodes 在 CI 里跑 ComfyUI 最新 main 测试。版本号写入文档 |
| **iframe SSO 跨域复杂** | v0 时验证 cookie domain / postMessage 方案。失败就走 React Flow |
| **模型 license 后期变更**（例如 Flux 突然改条款） | 每次客户部署前重审。维护一个 dated changelog |
| **GPU 滥用 / DDoS** | 沿用现有 quota 系统（已有 rate limit + per-user quota）+ ComfyUI 队列限制 |
| **GPL 边界被律师质疑** | docs/COMFY_LEGAL_FAQ.md 写清"ComfyUI 是独立服务、源码可向上游获取、我方代码不 fork"。客户合同附 |

---

## 不在本计划范围

- 订阅制 / 套餐（沿用现有 quota，不加 subscription）
- 工作流模板市场
- 多人协作画布
- 移动端 app
- ComfyUI fork / 修改

---

> 文档维护：v0 跑完后更新本文档"v0 → v1 gate" checkbox 状态。
> v1 启动时把本文档转成 GSD phase（`/gsd:plan-phase`）。
