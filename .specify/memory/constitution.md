<!--
Sync Impact Report
- Version change: (new) → 1.0.0
- Modified principles: N/A（新文件）
- Added sections: Core Principles、標準與門檻、開發流程與品質關卡、Governance
- Removed sections: N/A
- Templates requiring updates:
  - .specify/templates/plan-template.md ✅ 已對齊（含 Constitution Check 欄位）
  - .specify/templates/spec-template.md ✅ 已對齊（MVP、可獨立測試的 User Story）
  - .specify/templates/tasks-template.md ✅ 已對齊（分故事、MVP、可獨立測試）
- Follow-up TODOs: 無
-->

# CK123GoGo Constitution

## Core Principles

### I. 高品質與可維護性（Non‑Negotiable）

- MUST 遵循既有專案風格與目錄規範；保持小範圍精準修改，避免無關重排。
- MUST 通過 `gofmt -s`、`goimports`、`go vet` 與靜態分析；移除未使用程式碼。
- MUST 採用具語意的命名與清楚的錯誤包裝（`fmt.Errorf("context: %w", err)`），避免吞錯。
- MUST 嚴格資源管理（`Close()`/context）；禁止 goroutine 洩漏。
- SHOULD 僅為「非顯而易見的設計意圖」撰寫精煉註解，避免註解噪音。
- MUST 使用 Conventional Commits，維持可追溯歷史與自動化釋出。

### II. 測試先行與可測試性

- SHOULD 優先採 TDD（至少核心路徑）；若非 TDD，提交前 MUST 補齊測試。
- MUST 使用 table‑driven 單元測試；涉及 DB/gRPC/HTTP 的功能需有整合測試（可用 build tag 控制）。
- MUST 讓每個 User Story 可「獨立測試與示範」；測試不可互相依賴。
- SHOULD 維持穩定測試：去除 flakiness、固定時鐘/隨機、隔離外部資源（mock/fake）。
- SHOULD 以 CI 執行 `make test`（或等價）作為品質關卡；失敗不得合併。

### III. MVP 優先、拒絕過度設計（YAGNI/KISS）

- MUST 先交付最小可用產品（MVP），以「可上線切片」迭代擴展。
- MUST 避免預先抽象；出現第三次重複再抽象（Rule of Three）。
- SHOULD 偏好簡單直接的資料流與直連組合，暫緩層層間接與過度泛型化。
- SHOULD 使用小 PR、明確驗收條件與回滾策略，降低風險。

### IV. 使用者體驗一致性（API/CLI/文件）

- MUST 統一 API 行為：錯誤模型、分頁與排序參數、Idempotency、JSON 命名風格。
- MUST 錯誤訊息可讀且可追蹤（含 `trace_id`/`req_id`）；不得洩漏敏感資訊。
- MUST 文件、PR、Issue、規格與對外文字「一律使用正體中文」；專有名詞保留英文字形。
- SHOULD 提供清楚範例與 Quickstart，降低學習成本。

### V. 效能與可觀測性（Budget‑Driven）

- SHOULD 設定預設效能預算（可依服務調整）：
  - 同步 API：p95 < 250ms、p99 < 1s；錯誤率 < 1%
  - 後台作業：受限於 SLA 與隊列深度，需明確上限與退讓策略
- MUST 避免 N+1、缺失索引與無界查詢；對大量結果需分頁或限制。
- MUST 以指標/追蹤/日誌觀測熱點：延遲、錯誤率、DB/HTTP 用量、連線池等待。
- MUST 設定 context deadline/timeout、限制併發與具備 backpressure；避免無界併發。

## 標準與門檻（Standards & Gates）

- 風格與靜態檢查：`gofmt -s`、`goimports`、`go vet`、linter 全通過方可合併。
- 測試關卡：單元測試與關鍵整合測試必須通過；核心行為需具失敗後的明確訊號。
- 效能關卡：熱路徑變更需提供基本基準或延遲量測；若超出預算，需提出優化或分段交付。
- 安全關卡：禁止字串拼接 SQL；必用 prepared statements；機密不進版控；輸入長度/格式驗證。
- 文件關卡：更新 README/設計決策（必要時 ADR）；PR 需勾選 Constitution Check。

## 開發流程與品質關卡（Workflow & Quality Gates）

- 小步提交、易審閱：建議 PR < 400 行變更且聚焦單一目標。
- 提交訊息：Conventional Commits；破壞性變更需 `!` 或 `BREAKING CHANGE` 並附遷移。
- CI 需求：至少執行 `make lint`、`make test`、`make build`；任一失敗即阻擋合併。
- 版本政策：SemVer；以行為相容性決定 MAJOR/MINOR/PATCH。
- 變更管理：新增依賴需說明用途與風險；涉及架構調整需提出風險、回滾與替代方案。

## Governance

- 優先序：本憲章高於其他工作慣例與偏好；衝突時以本憲章為準。
- 修訂流程：以 PR 提案，說明動機、影響範圍、風險與回滾；需至少一名守門人審核通過。
- 合規審查：PR 模板中「Constitution Check」必填；審查者應對不合規處要求修正或給出明確豁免理由。
- 版本控管：依本檔尾端版本行維護；每次修訂需更新版本與日期並在檔案頂端同步報告。

**Version**: 1.0.0 | **Ratified**: 2025-11-26 | **Last Amended**: 2025-11-26
