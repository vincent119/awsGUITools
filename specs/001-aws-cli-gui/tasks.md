# Tasks: AWS CLI GUI（k9s 風格 TUI）

**Input**: Design documents from `/specs/001-aws-cli-gui/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: 憲章要求核心路徑具單元/整合測試，以下明確標註。
**Organization**: 依 User Story（US1~US4）與 Phase 分組，保證每個故事可獨立交付與測試。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可並行（不同檔案、無相依）
- **[Story]**: 對應 user story（US1~US4）
- 描述內含實際檔案路徑

## Phase 1: Setup（Shared Infrastructure）

**Purpose**: 建立專案骨架、CLI 入口與基本工具

- [x] T001 建立 `cmd/aws-tui/main.go`，使用 cobra 初始化 root command、版本旗標
- [x] T002 建立 `internal/app/app.go`，負責設定注入（config、AWS session、UI 啟動）並加入 `Makefile` 目標（lint/test/build）
- [x] T003 [P] 設定 `go.mod` 依賴（cobra、tview、aws-sdk-go-v2 模組），執行 `go mod tidy`
- [x] T004 [P] 建立 `internal/app/config/config.go` 與 `configs/config.example.yaml`，支援 profile、region、page size、timeout、theme

## Phase 2: Foundational（Blocking Prerequisites）

**Purpose**: 核心基礎建設，完成前不得開始 User Stories

- [x] T010 建立 `internal/aws/session/loader.go`：使用 AWS CLI profiles/regions 生成 `aws.Config`，並提供快取/切換機制
- [x] T011 [P] 建立 `internal/aws/clients/factory.go`：集中產生 ec2/rds/s3/lambda/cloudwatch/cloudwatchlogs Client，加入 context deadline 與重試設定
- [x] T012 [P] 實作 `internal/theme/theme.go` 與 `internal/theme/themes/*.json`，支援 dark/light/高對比載入與 Runtime 切換
- [x] T013 [P] 建立 `internal/ui/root.go`（tview Application、頁面容器、快捷鍵註冊），整合 config/theme
- [x] T014 建立 `internal/app/state/state.go`：管理目前 profile/region/filters、事件廣播
- [x] T015 [P] 建立 `internal/observability/logger.go` 與 metrics stub，確保 AWS 呼叫具延遲/錯誤統計
- [x] T016 建立 `internal/search/filter.go`：提供前綴/子字串/模糊比對 API，供清單頁使用
- [x] T017 建立 `tests/integration/aws/mock_clients_test.go`（使用 aws-sdk-go-v2 smithy stubs）供後續測試共用

---

## Phase 3: User Story 1 - 清單/詳情/關聯（MVP，P1） 🎯

**Goal**: 可在單一 Region/Profile 內瀏覽 EC2/RDS/S3/Lambda 清單、搜尋、進入詳情並顯示關聯
**Independent Test**: 使用者可以列出四種資源、搜尋、查看詳情/關聯，無需監控或操作功能

### Tests for User Story 1

- [x] T101 [P] [US1] 建立 `tests/unit/search/filter_test.go`，覆蓋前綴/子字串/模糊情境
- [ ] T102 [P] [US1] 建立 `tests/integration/aws/ec2_repo_test.go`（使用 mock）驗證分頁與關聯組合

### Implementation for User Story 1

- [x] T110 [US1] 實作 `internal/aws/repo/ec2_repo.go`：DescribeInstances + SG/IAM/EBS 關聯，支援分頁與 context timeout
- [x] T111 [US1] 實作 `internal/aws/repo/rds_repo.go`：DescribeDBInstances + SubnetGroup/ParameterGroup/SG
- [x] T112 [US1] 實作 `internal/aws/repo/s3_repo.go`：列出 buckets + versioning/encryption/policy/lifecycle（可用則抓）
- [x] T113 [US1] 實作 `internal/aws/repo/lambda_repo.go`：ListFunctions + GetFunction 詳情、環境變數、觸發來源
- [x] T114 [US1] 建立 `internal/models/*`（EC2Instance、RDSInstance、S3Bucket、LambdaFunction）與 ViewModel 轉換
- [x] T115 [US1] 建立 `internal/ui/list/list_page.go`：可切換資源類型、支援 `/` 搜尋、排序、分頁載入
- [x] T116 [US1] 建立 `internal/ui/detail/detail_page.go`：呈現資源基本資訊 + 關聯（Cards/Tab）
- [x] T117 [US1] 建立 `internal/ui/widgets/status_bar.go`：顯示 profile/region/theme/搜尋狀態
- [x] T118 [US1] 建立 `internal/ui/keymap/keymap.go`：整理快捷鍵並提供 `?` 說明視窗
- [x] T119 [US1] 接線 profile/region 切換：`internal/ui` modal，與 state 互動
- [x] T120 [US1] 整合 theme 切換按鍵 `t`，即時套用至 tview Styles
- [x] T121 [US1] 更新 quickstart.md 記錄鍵位與操作流程

**Checkpoint**: 完成後即為 MVP，可交付示範/內部狗食，後續增量不影響基本巡檢流程

---

## Phase 4: User Story 2 - CloudWatch 指標與日誌（P2）

**Goal**: 在詳情頁檢視主要 metrics 與最近 logs，時間區間可調
**Independent Test**: 單一資源詳情即可檢視指標/日誌，與操作/標籤無關

### Tests for User Story 2

- [x] T201 [P] [US2] `tests/integration/aws/metrics_fetcher_test.go`：模擬 CloudWatch GetMetricData 分頁/粒度
- [x] T202 [P] [US2] `tests/integration/aws/logs_fetcher_test.go`：模擬 CloudWatch Logs FilterLogEvents 多頁

### Implementation for User Story 2

- [x] T210 [US2] 建立 `internal/aws/metrics/templates.go`：對 EC2/RDS/Lambda/S3 定義 KPI 與查詢範本，支援自訂時間區間
- [x] T211 [US2] 建立 `internal/aws/logs/loggroup.go`：根據資源推導 log group（例如 `/aws/lambda/<fn>`）並提供分頁 API
- [x] T212 [US2] 在 `internal/ui/detail/detail_page.go` 加入 metrics/logs tabs，繪製文字 sparklines/統計摘要
- [x] T213 [US2] 建立 `internal/ui/widgets/time_range_picker.go` 供使用者調整查詢區間
- [x] T214 [US2] 確保 context deadline/退避策略應用於 CloudWatch 呼叫並於 UI 顯示載入/錯誤狀態

---

## Phase 5: User Story 3 - 基本操作（P2）

**Goal**: 對 EC2（Start/Stop/Reboot）、RDS（Start/Stop）、Lambda（Test Invoke）提供安全操作
**Independent Test**: 每項操作有二次確認/可選 Dry-Run，且會回報結果

### Tests for User Story 3

- [x] T301 [US3] `tests/unit/ops/confirm_dialog_test.go`：確認流程/文字/快捷鍵
- [x] T302 [US3] `tests/integration/aws/ec2_ops_test.go`：mock 驗證狀態輪詢與錯誤處理

### Implementation for User Story 3

- [x] T310 [US3] 建立 `internal/ops/ec2_ops.go`（Start/Stop/Reboot）：整合 dry-run 與進度輪詢
- [x] T311 [US3] 建立 `internal/ops/rds_ops.go`（Start/Stop）：判斷可停止條件並提示限制
- [x] T312 [US3] 建立 `internal/ops/lambda_ops.go`（Test Invoke）：允許輸入 payload、顯示結果/統計
- [x] T313 [US3] 建立 `internal/ui/modals/confirm_modal.go`，統一顯示操作確認/結果/錯誤
- [x] T314 [US3] 在 `detail_page` 內掛上 `a` 操作面板與狀態更新回饋

---

## Phase 6: User Story 4 - 標籤管理（P2）

**Goal**: 在詳情頁檢視與 CRUD 標籤，支援批次、衝突檢查與審計訊息
**Independent Test**: 單一資源可新增/刪除/更新標籤並立即回饋

### Tests for User Story 4

- [x] T401 [US4] `tests/unit/tags/validator_test.go`：key/value 驗證與衝突處理
- [x] T402 [US4] `tests/integration/aws/tags_repo_test.go`：驗證批次更新與錯誤分流

### Implementation for User Story 4

- [x] T410 [US4] 建立 `internal/tags/repository.go`：封裝 Create/Update/Delete，多資源共用
- [x] T411 [US4] 建立 `internal/ui/modals/tags_editor.go`：顯示現有標籤、允許批次新增/刪除/修改
- [x] T412 [US4] 整合標籤變更後的 UI refresh 與通知（status bar/Toast）
- [x] T413 [US4] 於 quickstart/README 補充標籤管理步驟與 IAM 權限需求

---

## Phase N: Polish & Cross-Cutting Concerns

- [x] T901 [P] 更新 `quickstart.md`、`docs/UX-flow.md`、`README.md`（若存在）以反映快捷鍵、主題、profiles/regions
- [x] T902 [P] 整體效能調校：加入 LRU 快取、調整預設分頁大小、記錄查詢延遲（必要時）
- [x] T903 [P] 安全掃描：檢查 logs 無敏感資訊、確保 config 加密/忽略
- [x] T904 執行 `make lint && make test && make build`，確保最終交付符合憲章門檻
- [ ] T905 建立 Demo 錄影或 README GIF，示範基本巡檢與切換操作

---

## Dependencies & Execution Order

### Phase Dependencies

- Setup（Phase 1）：無依賴，可立即啟動
- Foundational（Phase 2）：依賴 Setup；完成前不得開始任何 user story
- User Story 1（Phase 3）：依賴 Phase 2；完成後即達 MVP，可獨立交付
- User Stories 2~4（Phases 4~6）：皆依賴 Phase 3，彼此可視資源平行，但需避免同檔衝突
- Polish（Phase N）：所有目標故事完成後再進行

### Parallel Opportunities

- Setup 與 Foundational 中標註 [P] 任務可並行
- User Story 1 中，repo/model/UI 可依文件協調並行（注意同檔案）
- User Story 2~4 各自模組相對獨立，可指派不同工程師
- 測試任務（build tag mock）可與實作交錯進行，但需確保依賴檔案已完成

---

## Implementation Strategy（MVP → Incremental）

1. 完成 Phase 1 + Phase 2（骨架、session、UI root、搜尋、obsv）
2. Phase 3（US1）達成可 demo 的 MVP（清單/詳情/關聯/搜尋/切換）
3. Phase 4（US2）加入監控整合
4. Phase 5（US3）提供操作能力（需額外權限與確認）
5. Phase 6（US4）完善標籤治理
6. Phase N 針對文件、效能、安全進行收尾
