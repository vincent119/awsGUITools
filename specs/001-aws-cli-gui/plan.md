# Implementation Plan: AWS CLI GUI（類 k9s 的終端互動體驗）

**Branch**: `[001-aws-cli-gui]` | **Date**: 2025-11-26 | **Spec**: `specs/001-aws-cli-gui/spec.md`
**Input**: Feature specification from `/specs/001-aws-cli-gui/spec.md`

## Summary

本計畫將以 Go 開發一個終端 GUI 的 AWS 資源巡檢與操作工具，操作邏輯參照 k9s：以鍵盤為主、清單 → 詳情 → 操作的切換體驗。  
第一階段（MVP）聚焦「可瀏覽清單、進入詳情、查看關聯、快速搜尋」，第二階段擴充 CloudWatch 指標/日誌、基本操作（EC2/RDS/Lambda）、與標籤管理。  
技術選擇：Go 1.22+、cobra（CLI）、tview（UI）、AWS SDK v2（API 呼叫）。支援主題切換、多帳號（AWS CLI profiles）、多區域切換、與快速資源搜尋。

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**:

- CLI：`github.com/spf13/cobra`
- TUI：`github.com/rivo/tview`（底層 `github.com/gdamore/tcell/v2`）
- AWS：`github.com/aws/aws-sdk-go-v2` 及其對應 service 模組（ec2、rds、s3、lambda、cloudwatch、cloudwatchlogs）
- 搜尋（可選）：內建前綴/子字串/模糊比對，後續再評估第三方  
  **Storage**: 無後端 DB；快取以記憶體為主（可考慮本機快取檔案，但 MVP 先不做）  
  **Testing**: `go test`（table-driven）；整合測試以 aws-sdk-go-v2 mock/smithy stubs 與 build tags 控制  
  **Target Platform**: macOS/Linux/Windows 終端環境  
  **Project Type**: 單一 Go 應用（tui）  
  **Performance Goals**:
- 清單查詢（中小帳號）：p95 < 1s
- 單資源詳情：p95 < 250ms（單 API 呼叫；必要時並行查詢但受限於節流）  
  **Constraints**:
- CLI/終端互動需要良好鍵盤體驗與即時回應
- 資料量大時，必須分頁與增量載入；對 CloudWatch 以區間與取樣率控制  
  **Scale/Scope**: 預期支援數百～數千資源等級，避免一次抓全與無界記憶體成長

## Constitution Check

必須通過下列門檻後方可合併：

- 高品質/可維護：`gofmt -s`、`goimports`、`go vet`、linter 全通過；小範圍精準修改。
- 測試先行與可測試性：核心路徑具單元測試；整合測試以 mock/stub 驗證 AWS 邊界。
- MVP 優先：先完成「清單/詳情/關聯/搜尋」的可用切片；CloudWatch、操作、標籤第二波。
- 使用者體驗一致性：統一錯誤模型與文案（正體中文）；鍵盤快捷鍵一致；明確回饋。
- 效能與可觀測性：關鍵 API 加入延遲與錯誤率度量；context deadline 與退避重試。
- 安全：不輸出敏感憑證；最小權限；避免將機密寫入檔案或日誌。

## Project Structure

### Documentation (this feature)

```text
specs/001-aws-cli-gui/
├── plan.md              # 本檔案
├── research.md          # 技術與 UX 研究、架構決策
├── data-model.md        # 資料模型與關聯
├── quickstart.md        # 安裝/設定/操作指南與快捷鍵
└── contracts/           # （可選）格式定義/快捷鍵表/主題格式說明
```

### Source Code (repository root)

```text
cmd/aws-tui/main.go              # 進入點（DI/初始化/啟動）

internal/app/
├── app.go                       # 應用啟動、DI、生命週期
├── config/                      # 設定（profile/region/page size/timeouts）
└── state/                       # 全域狀態（目前 profile/region/filters）

internal/ui/                     # tview 畫面與導覽
├── root.go                      # 主頁面與路由（清單/詳情）
├── list/                        # 清單頁（EC2/RDS/S3/Lambda）
├── detail/                      # 詳情頁（與關聯）
├── widgets/                     # 通用元件（狀態列、搜尋列、通知）
└── keymap/                      # 鍵盤快捷鍵與說明

internal/theme/
├── theme.go                     # 主題模型與套用
└── themes/                      # 內建主題 JSON（dark/light/高對比）

internal/aws/
├── session/                     # 使用 AWS CLI profiles 與 region 建立 Config/Client
├── clients/                     # 具體 service client（ec2/rds/s3/lambda/cw/cwl）
├── repo/                        # 查詢介面（清單/詳情/關聯/分頁/快取）
├── metrics/                     # CloudWatch metrics 查詢
└── logs/                        # CloudWatch logs 查詢（分頁）

internal/models/                 # 領域模型（EC2Instance/RDSInstance/...）
internal/search/                 # 搜尋（name/tag/id/arn，前綴/子字串/模糊）
internal/ops/                    # 基本操作（EC2 start/stop/reboot、RDS start/stop、Lambda invoke）
internal/tags/                   # 標籤 CRUD 與批次
internal/observability/          # 指標/追蹤/日誌（封裝）

tests/
├── unit/                        # table-driven 單元測試
└── integration/                 # aws-sdk mock/smithy stubs（build tags 控制）
```

**Structure Decision**: 採單一 Go 應用；以 `internal/` 為主，避免過早暴露 `pkg/` API。模組邊界依資源類型與功能切分，符合 MVP 與可維護性。

## Complexity Tracking

目前無超出憲章的複雜度；如未來引入離線快取/圖視覺化，再評估合理性。
