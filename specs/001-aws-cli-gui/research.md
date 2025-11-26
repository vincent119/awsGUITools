# Research: AWS CLI GUI（k9s 風格 TUI）

## 技術與 UX 研究

- 介面框架：tview（基於 tcell），適合列表/彈窗/多頁面切換；可自訂 Styles 以支援主題。  
- 操作模式：參照 k9s  
  - 清單頁：方向鍵/`j`/`k` 移動，`/` 搜尋，`Enter` 進入詳情，`Esc` 返回  
  - 全域：`p` 切換 profile，`r` 切換 region，`t` 切換主題，`?` 顯示快捷鍵  
- 主題：提供 dark/light/高對比三種內建 JSON；以 `theme.Manager` 套用至 tview Styles 與自訂元件。

## AWS SDK v2 使用策略

- 認證：沿用 AWS CLI profiles（`~/.aws/config`、`~/.aws/credentials`），以 shared config loader 建立 `aws.Config`。  
- Region：以 `WithRegion()` 切換；切換時刷新 client 與快取。  
- 節流/重試：採 SDK 內建重試並加入 context deadline；對清單類 API 進行分頁與平行度限制。  
- 快取：以 in-memory LRU（後續）避免重複查詢；MVP 可先無快取，確保正確性。  
- 關聯查詢：  
  - EC2：透過 `DescribeInstances`、`DescribeSecurityGroups`、`DescribeIamInstanceProfileAssociations`、`DescribeVolumes`  
  - RDS：`DescribeDBInstances`，並補抓 `DBSubnetGroup`、`DBParameterGroups`、`VpcSecurityGroups`  
  - S3：`GetBucketVersioning`、`GetBucketEncryption`、`GetBucketPolicy`（如有）、`GetBucketLifecycleConfiguration`（如有）  
  - Lambda：`GetFunction`（含配置與環境變數），必要時補抓 event source mapping  
  - CloudWatch：`GetMetricData`；CloudWatch Logs：`FilterLogEvents`/`GetLogEvents`

## Rate Limit 與效能

- 並發控制：以 `errgroup` + 限制 semaphore；避免過多同時呼叫。  
- 分頁：所有清單 API 必須分頁；UI 採增量載入避免一次性阻塞。  
- 逾時：同步單呼叫預設 5s；跨頁彙整 1s～2s 內完成（中小帳號），超時則顯示部分資料與提示。  
- 指標圖：MVP 以彙總數值/小圖表（sparklines）呈現，後續再優化取樣與粒度。

## 鍵盤快捷鍵（初稿）

- 全域：`p` 切換 profile、`r` 切換 region、`t` 切換主題、`/` 搜尋、`?` 說明、`q` 退出  
- 清單：方向鍵/`j`/`k` 移動、`Enter` 詳情、`f` 篩選、`o` 排序  
- 詳情：`←` 返回、`m` 指標、`l` 日誌、`g` 關聯、`a` 操作、`x` 標籤

## 風險與備援

- 權限不足：清楚顯示 AccessDenied 與對應 ARN/Action，提供最小權限建議。  
- 大量帳號/跨帳號：初版不做 org 聚合；以單 profile 快速切換。  
- Logs 不存在：顯示不可用提示與啟用指南。  

