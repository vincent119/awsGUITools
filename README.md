# AWS TUI - 終端 GUI 巡檢與管理 AWS 資源

以 Go 開發的終端 GUI 工具，參考 k9s 操作邏輯，提供直觀的 AWS 資源瀏覽與管理體驗。

## 功能特色

- **資源瀏覽**：EC2、RDS、S3、Lambda 清單與詳情
- **關聯檢視**：Security Groups、IAM Role、EBS、Subnet Group 等
- **監控整合**：CloudWatch Metrics（CPU、連線數等）與 Logs
- **基本操作**：Start/Stop/Reboot（EC2/RDS）、Test Invoke（Lambda）
- **標籤管理**：新增、刪除、修改資源標籤
- **多帳號/區域**：快速切換 AWS Profile 與 Region
- **主題支援**：Dark、Light、High-Contrast

## 快速開始

```bash
# 建置
go build -o aws-tui ./cmd/aws-tui

# 執行
./aws-tui

# 指定設定檔
./aws-tui --config configs/config.yaml
```

## 快捷鍵

| 按鍵 | 功能 |
| ------ | ------ |
| `1`-`4` | 切換資源類型（EC2/RDS/S3/Lambda） |
| `/` | 搜尋 |
| `Enter` | 進入詳情 |
| `g` | 重新整理 |
| `p` | 切換 Profile |
| `r` | 切換 Region |
| `t` | 切換主題 |
| `a` | 操作面板 |
| `T` | 標籤編輯器 |
| `?` | 說明 |
| `q` | 離開 |

## 設定

參考 `configs/config.example.yaml`：

```yaml
profile: default
region: ap-northeast-1
theme: dark
page_size: 50
timeout: 15s
```

## IAM 權限

### 唯讀（基本瀏覽）

```json
{
  "Effect": "Allow",
  "Action": [
    "ec2:Describe*",
    "rds:Describe*",
    "s3:ListAllMyBuckets",
    "s3:GetBucket*",
    "lambda:List*",
    "lambda:GetFunction",
    "cloudwatch:GetMetricData",
    "logs:FilterLogEvents"
  ],
  "Resource": "*"
}
```

### 操作功能

```json
{
  "Effect": "Allow",
  "Action": [
    "ec2:StartInstances",
    "ec2:StopInstances",
    "ec2:RebootInstances",
    "rds:StartDBInstance",
    "rds:StopDBInstance",
    "rds:RebootDBInstance",
    "lambda:InvokeFunction"
  ],
  "Resource": "*"
}
```

### 標籤管理

```json
{
  "Effect": "Allow",
  "Action": [
    "ec2:CreateTags",
    "ec2:DeleteTags",
    "rds:AddTagsToResource",
    "rds:RemoveTagsFromResource",
    "s3:PutBucketTagging",
    "s3:DeleteBucketTagging",
    "lambda:TagResource",
    "lambda:UntagResource"
  ],
  "Resource": "*"
}
```

## 開發

```bash
# 安裝依賴
go mod tidy

# 執行測試
make test

# 靜態分析
make lint

# 建置
make build
```

## 專案結構

```bash
cmd/aws-tui/          # CLI 進入點
internal/
  app/                # 應用生命週期與設定
  aws/                # AWS SDK 封裝（session、clients、repo）
  models/             # 資料模型
  ops/                # 資源操作（start/stop/reboot）
  service/            # 業務邏輯層
  tags/               # 標籤管理
  theme/              # 主題管理
  ui/                 # tview UI 元件
configs/              # 設定範例
specs/                # 規格與任務追蹤
tests/                # 測試
```

## 授權

MIT License
