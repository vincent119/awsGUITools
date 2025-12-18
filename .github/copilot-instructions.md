# AWS TUI - Copilot 指令

## 專案概覽

AWS TUI 是一個用 Go 開發的終端 GUI 工具，參考 k9s 操作邏輯，提供直觀的 AWS 資源瀏覽與管理體驗。使用 **tview** 作為 TUI 框架，**AWS SDK v2** 與 AWS 服務互動。

### 核心架構

```
cmd/aws-tui/         # CLI 入口，使用 cobra，組裝 App
internal/app/        # 應用生命週期管理，協調各層元件
  ├── config/        # 組態載入 (YAML)
  └── state/         # 全域狀態 (profile/region/theme/language)
internal/aws/        # AWS SDK v2 封裝層
  ├── clients/       # Factory 模式，依 profile/region 產生各服務 client
  ├── repo/          # Repository 層，封裝 AWS API 呼叫與模型轉換
  ├── logs/          # CloudWatch Logs fetcher
  ├── metrics/       # CloudWatch Metrics fetcher + 模板
  └── session/       # AWS config/credentials 載入邏輯
internal/service/    # 業務服務層
  └── resource/      # 統一資源查詢介面，含快取、搜尋、狀態管理
internal/ui/         # tview 畫面管理
  ├── root.go        # 主應用入口，協調所有 UI 元件與事件
  ├── list/          # 資源清單頁 (Table)
  ├── detail/        # 資源詳情頁 (Tabs: Info/Metrics/Logs)
  ├── modals/        # 互動視窗 (profile/region 選擇器、標籤編輯器、確認對話框)
  ├── widgets/       # 可重用元件 (StatusBar, TimeRangePicker)
  └── keymap/        # 鍵盤快捷鍵定義與說明
internal/i18n/       # 國際化 (en, zh-TW)，使用嵌入式 JSON
internal/theme/      # 主題管理 (dark/light/high-contrast)，動態套用至 tview
```

### 資料流向

1. **啟動**: `main.go` → `app.New()` → 載入 config → 初始化 session/factory/service → 建立 UI
2. **列表查詢**: UI event → `resource.Service.ListItems()` → repo → AWS SDK → 轉換為 `models.ListItem`
3. **詳情檢視**: UI 選取項目 → `resource.Service.GetDetail()` → 從快取或 repo 取得 → `models.DetailView`
4. **操作執行**: UI 觸發 → `resource.Service.DoOperation()` → ops 層 (ec2_ops/rds_ops/lambda_ops) → AWS API
5. **切換 Profile/Region**: UI modal → 更新 `state.Store` → UI 重新載入資料

## 關鍵設計決策

### 1. Repository 與 Service 分層

- **Repository** (`internal/aws/repo/`): 專注於 AWS SDK 呼叫與 AWS 型別 ↔ 內部 models 轉換
- **Service** (`internal/service/resource/`): 提供統一介面給 UI，處理快取、搜尋過濾、狀態管理
- 理由: 隔離 AWS SDK 依賴，方便測試與 mock；Service 層可橫跨多個 repo 聚合資料

### 2. Factory 模式產生 AWS Clients

`clients.Factory` 根據 profile/region 動態產生各服務 client，避免在各處直接操作 `aws.Config`。所有 AWS 操作都透過 Factory 取得 client，確保認證與區域設定一致。

### 3. State Store 管理全域狀態

`state.Store` 使用 `sync.RWMutex` 保護，供 UI 與 service 讀寫 profile/region/theme/language/filter。切換 profile/region 時更新 Store，UI 監聽變化重新載入。

### 4. tview 的事件驅動模式

- 所有 UI 互動透過 `SetInputCapture()` 攔截鍵盤事件
- 使用 `app.QueueUpdateDraw()` 確保 UI 更新在主執行緒
- Modal 使用 `pages.ShowPage()` / `HidePage()` 切換
- 避免在 goroutine 直接操作 tview 元件，改用 `QueueUpdateDraw()` callback

### 5. i18n 與 Theme 動態切換

- i18n: `i18n.T("key")` 從嵌入式 JSON 讀取，支援 en/zh-TW
- Theme: `theme.Manager.Apply()` 動態更新 tview 的顏色方案，不需重啟

## 開發流程

### 建置與執行

```bash
make build    # 編譯到 bin/aws-tui
make run      # 建置並執行
make test     # 執行所有測試
make lint     # gofmt + goimports + go vet
```

### 測試策略

- **單元測試** (`tests/unit/`): 邏輯函式 (search/filter, i18n, tags validator)
- **整合測試** (`tests/integration/`): AWS repo/ops 層，使用 mock clients
- 命名: `<package>_test.go`，使用 table-driven tests
- Mock: 手寫 mock 結構實作 AWS SDK 介面，避免外部依賴

### 新增資源類型

1. 在 `internal/models/` 定義模型 (例如 `DynamoDBTable`)
2. 在 `internal/aws/repo/` 新增 repository (例如 `dynamodb_repo.go`)
3. 在 `internal/aws/clients/factory.go` 新增 client 方法
4. 在 `resource.Service` 擴充 `Kind` 與 `ListItems()`/`GetDetail()` 邏輯
5. 更新 UI keymap 與 root.go 的資源切換邏輯

### 新增操作功能

1. 在 `internal/ops/` 新增或擴充 ops 檔案 (例如 `ec2_ops.go`)
2. 在 `resource.Service` 新增 `Do<Operation>()` 方法
3. 在 `internal/ui/modals/confirm_modal.go` 使用確認對話框
4. 更新 keymap 與 UI event handler

## 編碼規範重點

### Go 風格 (參考 `.cursorrules` 與 `.cursor/rules/go-style-and-behavior.mdc`)

- **錯誤處理**: 立即檢查 `err`，使用 `fmt.Errorf("context: %w", err)` 包裝
- **context 傳遞**: 所有 I/O 操作第一個參數為 `ctx context.Context`，不在 struct 保存
- **併發**: goroutine 需有明確結束條件 (WaitGroup/errgroup/ctx.Done)；io.Closer 必須 Close
- **JSON 與結構**: 對外模型加 `json` tag；解碼外部輸入時用 `DisallowUnknownFields()`
- **imports 分組**: 標準庫 → 第三方 → 專案內

### 專案特定模式

- **Repository 方法**: 命名 `List<Resource>()`, `Describe<Resource>()`, `Get<Resource>Detail()`
- **Service 方法**: `ListItems()`, `GetDetail()`, `DoOperation(kind, op, id)`
- **UI 元件**: 大寫開頭導出，建構函式 `New<Component>()`，使用 `tview.Primitive`
- **i18n 鍵值**: `<category>.<key>` (例如 `list.title.ec2`, `error.network.timeout`)
- **測試**: table-driven，使用 `t.Run(tt.name, ...)`，mock 手寫而非套件

### 命名慣例

- Package: 全小寫，單數 (例如 `client`, `model`, `repo`)
- 檔案: 小寫 + 底線 (例如 `ec2_repo.go`, `list_page.go`)
- 型別: 大寫駝峰 (例如 `EC2Instance`, `ListItem`)
- 私有欄位/函式: 小寫駝峰 (例如 `currentKind`, `initLayout()`)

## 常見陷阱

1. **tview UI 更新**: 不在 goroutine 直接修改元件，改用 `app.QueueUpdateDraw(func() {...})`
2. **AWS client 重用**: 不快取 client，每次操作透過 Factory 取得 (profile/region 可能變化)
3. **Context timeout**: 每個 AWS 操作設定合理 timeout (預設 15s)，使用 `context.WithTimeout()`
4. **State 讀寫**: 使用 Store 的方法 (含鎖)，不直接存取內部欄位
5. **i18n fallback**: 若 key 不存在，`i18n.T()` 回傳 key 本身，不 panic

## 除錯技巧

- 使用 `internal/observability/logger.go` 記錄結構化日誌 (slog)
- 啟用 AWS SDK debug: 設定 `AWS_SDK_LOG_LEVEL=debug`
- tview debug: 使用 `app.EnableMouse(true)` 方便互動測試
- 檢查 `~/.aws/credentials` 與 `~/.aws/config` 確認 profile 設定

## 參考資源

- [tview 官方文件](https://github.com/rivo/tview)
- [AWS SDK Go v2](https://aws.github.io/aws-sdk-go-v2/docs/)
- [Uber Go Style Guide](https://github.com/uber-go/guide)
- 專案規格: `specs/001-aws-cli-gui/spec.md`
