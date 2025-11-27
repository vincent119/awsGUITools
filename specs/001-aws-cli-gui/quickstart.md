# Quickstart: AWS CLI GUI（Go/cobra/tview/AWS SDK v2）

## 前置需求

- Go 1.22+  
- 已設定 AWS CLI 認證與 Profiles（`aws configure --profile <name>`）  
- 有效的 AWS 權限（唯讀起步，操作功能需額外授權）  

## 安裝與執行（未來）

```bash
# 建置
go build -o aws-tui ./cmd/aws-tui

# 執行（預設載入 AWS 預設 profile 與 region）
./aws-tui
```

## 主要操作與快捷鍵

- `1`/`2`/`3`/`4`：切換資源（EC2、RDS、S3、Lambda）
- `/`：焦點至搜尋列並輸入關鍵字（Name/Tag/ID）
- `Enter`：套用搜尋或於清單進入詳情
- `g`：重新整理目前資源清單
- `p`：切換 AWS Profile
- `r`：切換 AWS Region
- `t`：切換主題（dark → light → high-contrast）
- `a`：開啟操作面板（Start/Stop/Reboot 等）
- `T`：開啟標籤編輯器（新增/刪除/修改標籤）
- `?`：顯示快捷鍵說明；`q`：離開應用

## 標籤管理

1. 選取資源後按 `T` 開啟標籤編輯器
2. 現有標籤會顯示在上方列表，按 `d` 可刪除
3. 在下方表單輸入 Key/Value 後按「新增」
4. 按「儲存」提交變更，或按「取消」/`Esc` 放棄

### 標籤限制

- Key 最長 128 字元，Value 最長 256 字元
- 每個資源最多 50 個標籤
- Key 不可使用 `aws:` 前綴（AWS 保留）

### IAM 權限需求

標籤管理需要以下權限：

- EC2：`ec2:CreateTags`、`ec2:DeleteTags`
- RDS：`rds:AddTagsToResource`、`rds:RemoveTagsFromResource`
- S3：`s3:PutBucketTagging`、`s3:DeleteBucketTagging`
- Lambda：`lambda:TagResource`、`lambda:UntagResource`

## 設定

- `configs/config.example.yaml` 可作為起點：設定 profile、region、主題、查詢 page size、timeout 等。  
- 主題檔位於 `internal/theme/themes/*.json`，可依需求新增。  

## 故障排除

- AccessDenied：檢查目前 profile 與必要 IAM 權限；工具將顯示缺少的 Action 與建議。  
- 逾時或節流：調整分頁大小/區間或稍後重試；工具內建退避重試與提示。  

