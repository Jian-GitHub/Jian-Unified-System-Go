# Jian Unified System - Go

<div align="center">
  <a href="https://deepwiki.com/Jian-GitHub/Jian-Unified-System-Go">
    <img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki"/>
  </a>
</div>

## ✨ Prototype
> 🎨 [Excalidraw Concept Sketch](https://excalidraw.com/#json=RgtulBzQifhCgScE2CES-,p-0YAriCu_qE712OWTSWvQ)

---

<p>🚧 Work in Progress...</p>

## 工程文档

项目包含 Apollo 统一账户服务、JQuantum 量子计算任务服务，以及共享的 jus-core / jus-hermes 组件。Go 管理接口、认证、数据与任务执行，Python 客户端负责 Qiskit 电路编码与提交。

完整导航见 **[工程文档目录](docs/README.md)**。文档基于当前工作区整理，区分实现现状、验证结果与待确认事项；按维护者要求跳过 QuEST 第三方 C++ 源码。

新对话从 **[项目交接摘要](docs/context/README.md)** 开始；长期架构背景见 [DDD 讨论记录](docs/discussions/2026-09-05-project-and-ddd.md)。协作与记录更新约定见 [AGENTS.md](AGENTS.md)。

- [架构与流程](docs/architecture.md) · [接口协议](docs/interfaces.md) · [数据设计](docs/data-model.md)
- [配置参考](docs/configuration.md) · [开发与构建](docs/development.md) · [部署与 auth.yaml](docs/operations.md)
- [测试与验收](docs/quality.md) · [问题清单](docs/review.md) · [源码导航](docs/source-map.md)
- [可编辑 Excalidraw 架构图](docs/diagrams/system-architecture.excalidraw)
- [Apollo 并行重构](docs/apollo-refactoring.md)：新实现位于 `apollo-api-ddd` / `apollo-rpc-ddd`，旧服务保持原状。

## Star History

<a href="https://www.star-history.com/#Jian-GitHub/Jian-Unified-System-Go&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=Jian-GitHub/Jian-Unified-System-Go&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=Jian-GitHub/Jian-Unified-System-Go&type=Date" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=Jian-GitHub/Jian-Unified-System-Go&type=Date" />
 </picture>
</a>

## Stats

![Alt](https://repobeats.axiom.co/api/embed/62d3047f9991186059fe32a41b899a343daa6ccc.svg "Repobeats analytics image")
