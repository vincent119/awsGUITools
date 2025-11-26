# Feature Specification: AWS CLI GUI（類 k9s 的終端互動體驗）

**Feature Branch**: `[001-aws-cli-gui]`  
**Created**: 2025-11-26  
**Status**: Draft  
**Input**: User description: "/speckit.specify 建立一個 CLI GUI 工具，參照 k9s 的操作邏輯，查詢 AWS 資源（EC2、RDS、S3、Lambda），顯示詳情與關聯，整合 CloudWatch 指標/日誌，提供基本操作（啟動/停止/重啟）與標籤管理（新增/刪除/修改）。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 瀏覽資源清單與詳情（核心 MVP） (Priority: P1)

使用者可以在終端 GUI 中快速切換資源類型（EC2、RDS、S3、Lambda），即時列出資源清單，支援搜尋/篩選/排序，並可進入單一資源頁面查看詳細資訊與關聯（例如：EC2 ↔ SG/IAM Role/EBS）。

**Why this priority**: 這是最小可用價值切片（MVP）的核心：沒有查詢與詳情/關聯，就無法完成後續操作或監控整合。

**Independent Test**:  
- 在具備有效 AWS 認證與選定 Region 的情況下，能獨立完成：  
  1) 列出 EC2/RDS/S3/Lambda 任一類型清單  
  2) 點選單一資源查看基本欄位與主要關聯  
  3) 搜尋/篩選/排序對清單生效

**Acceptance Scenarios**:
1. Given 已設定 AWS 認證與 Region，When 開啟工具並切換至 EC2，Then 顯示 EC2 清單且可用關鍵字篩選出指定實例。  
2. Given EC2 清單頁，When 進入某 EC2 詳情，Then 顯示 InstanceId、Name、狀態、型號、AZ、PrivateIP、PublicIP、VPC、Subnet、SecurityGroups、IAM Role、EBS Volumes。  
3. Given RDS 清單頁，When 進入某 RDS 詳情，Then 顯示 DBInstanceIdentifier、Engine、版本、Multi‑AZ、Subnet Group、Parameter Group、Security Group、存取端點。  
4. Given S3 清單頁，When 進入某 Bucket 詳情，Then 顯示區域、版本設定、加密設定、Bucket Policy/Lifecycle。  
5. Given Lambda 清單頁，When 進入某函式詳情，Then 顯示 Runtime、記憶體/逾時、別名/版本、IAM Role、環境變數、觸發來源（可用時）。  
6. Given 清單頁，When 使用搜尋/篩選條件，Then 清單在 1 秒內回應並更新結果（中小型帳號資料量）。  

---

### User Story 2 - 監控資訊整合（CloudWatch 指標與日誌） (Priority: P2)

使用者可以在資源詳情頁直接查看關鍵 CloudWatch 指標（Metrics）與最近日誌（Logs），支援時間區間/維度選擇與快速跳轉。

**Why this priority**: 整合監控可在單一介面完成巡檢/診斷，提升效率。

**Independent Test**:  
- 在任一資源詳情頁，能獨立查詢並呈現主要指標/日誌（例如 EC2 CPUUtilization、RDS CPU/FreeableMemory、Lambda Errors/Duration、S3 請求指標與存取日誌若啟用）。

**Acceptance Scenarios**:
1. Given EC2 詳情頁，When 選擇過去 1 小時區間，Then 顯示 CPUUtilization、NetworkIn/Out、StatusCheckFailed 指標圖表（文字小圖/統計）。  
2. Given Lambda 詳情頁，When 查看日誌，Then 顯示最近 N 筆 log events，支援上/下頁。  
3. Given RDS 詳情頁，When 切換時間區間，Then 指標以 <1.5 秒完成刷新（資料可用時）。  

---

### User Story 3 - 基本操作（啟動/停止/重啟等） (Priority: P2)

使用者可以對資源執行安全範圍內的基本操作，並顯示進度/結果與審計訊息（Dry‑Run/確認）。

**Why this priority**: 常見日常維運需求；但需二次確認避免誤操作。

**Independent Test**:  
- 在 EC2 清單或詳情頁，能觸發 Start/Stop/Reboot（符合 IAM 權限、非生產帳號或需顯式確認）。  
- 在 RDS，能觸發 Start/Stop（僅支援可停止的類型）；Lambda 可觸發 Test Invoke（選擇 payload）。

**Acceptance Scenarios**:
1. Given EC2 實例為 stopped，When 執行 Start 並確認，Then 顯示狀態轉換進度與最後成功狀態。  
2. Given EC2 實例為 running，When 執行 Stop 並確認，Then 在操作逾時上限內回報完成或提示背景進行。  
3. Given Lambda 函式，When 提供測試 payload 並執行，Then 顯示執行結果與延遲/記憶體統計（可用時）。  

---

### User Story 4 - 標籤管理（增/刪/改） (Priority: P2)

使用者可於詳情頁檢視與管理資源標籤（Tags），支援批次套用與衝突檢查。

**Why this priority**: 標籤是成本歸屬、治理與自動化的基礎。

**Independent Test**:  
- 於 EC2/RDS/S3/Lambda 詳情頁，能新增/刪除/更新單一或多個標籤；變更可追蹤（審計訊息/日誌）。

**Acceptance Scenarios**:
1. Given EC2 詳情頁，When 新增 `owner=team-a`，Then 標籤立即可見並與 AWS 同步成功。  
2. Given S3 詳情頁，When 批次更新多個 key/value，Then 顯示變更摘要並於 2 秒內完成提交。  
3. Given 既有 key，When 嘗試重複新增，Then 工具提示並要求改為更新或取消。  

---

### Edge Cases

- 認證/權限不足（AccessDenied）：需清楚呈現服務/動作/資源 ARN 與建議處置。  
- 無任何資源（空清單）：應顯示空狀態與可行動指引。  
- AWS API 節流（Throttling）與速率限制：自動退避重試並提示使用者。  
- 大量結果（>1000）：強制分頁與增量載入，避免 UI 卡頓。  
- CloudWatch 無對應日誌群組/流：給出可選建置指南或僅顯示不可用。  
- 多 Region 切換：需清楚顯示目前 Region 與切換影響。  

## Requirements *(mandatory)*

### Functional Requirements

- **FR‑001（MVP）**：提供基於終端的 GUI（鍵盤快捷鍵、上下/左右切換、快速搜尋）。  
- **FR‑002（MVP）**：支援資源清單與詳情：EC2、RDS、S3、Lambda。  
- **FR‑003（MVP）**：呈現主要關聯：  
  - EC2 ↔ Security Groups、IAM Role、EBS Volumes、VPC/Subnet  
  - RDS ↔ Subnet Group、Parameter Group、Security Group、Endpoint  
  - S3 ↔ Bucket Policy、Lifecycle（有啟用則顯示）  
  - Lambda ↔ IAM Role、Environment Variables、觸發來源（可用則顯示）  
- **FR‑004（MVP）**：清單支援搜尋/篩選/排序與分頁；回應時間（中小帳號）p95 < 1s。  
- **FR‑005（P2）**：整合 CloudWatch 指標（核心 KPI）與最近日誌檢視（可換區間/分頁）。  
- **FR‑006（P2）**：基本操作：  
  - EC2：Start/Stop/Reboot（二次確認與可選 Dry‑Run）  
  - RDS：Start/Stop（僅支援可停止類型，二次確認）  
  - Lambda：Test Invoke（可填 payload，回傳統計）  
- **FR‑007（P2）**：標籤管理：增/刪/改、批次更新、衝突檢查、審計訊息。  
- **FR‑008**：多 Region 切換與顯示目前作用中的認證/角色（避免誤區）。  
- **FR‑009**：錯誤處理與回報：分類（認證/權限/節流/超時/其他），訊息可讀且可追蹤（含 req/trace id 若可）。  
- **FR‑010**：設定管理：預設 Region、頁面大小、CloudWatch 查詢預設區間、超時/重試策略。  

範例需求標示不明情況：  
- **FR‑011**：CloudWatch Logs 來源需指定 group/stream 規則［NEEDS CLARIFICATION：是否自動探索或使用命名慣例？］  
- **FR‑012**：S3 請求指標需明確開通與計費考量［NEEDS CLARIFICATION］  

### 非功能性（依憲章）

- 效能預算：同步操作 p95 < 250ms（單次 API 呼叫）；跨頁彙整 p95 < 1s（中小帳號）。  
- 可觀測性：記錄核心操作延遲/錯誤率；對 AWS SDK 呼叫設置 context deadline 與重試（指數退避）。  
- 安全性：不得輸出敏感憑證於日誌；遵循最小權限；本機設定檔或環境變數載入。  
- 可維護性：模組邊界以資源類型切分；避免過度抽象；第三次重複再抽象。  
- 在地化：所有文件與 UI 文案一律正體中文，專有名詞保留英文字形。  

### Key Entities *(include if feature involves data)*

- **EC2Instance**：id、name、state、type、az、private_ip、public_ip、vpc、subnet、sg_ids[]、iam_role、ebs_vol_ids[]  
- **RDSInstance**：id、engine、engine_version、multi_az、endpoint、subnet_group、parameter_group、sg_ids[]  
- **S3Bucket**：name、region、versioning、encryption、policy(optional)、lifecycle(optional)  
- **LambdaFunction**：name、arn、runtime、memory、timeout、role、env_vars、triggers(optional)  
- **Relations**：  
  - EC2Instance — SecurityGroup (N:N)  
  - EC2Instance — IAMRole (1:0..1)  
  - EC2Instance — EBSVolume (1:N)  
  - RDSInstance — SubnetGroup/ParameterGroup/SecurityGroup  
  - LambdaFunction — IAMRole (1:0..1)  
  - S3Bucket — Policy/Lifecycle (0..1)  

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC‑001（MVP）**：在一個 Region 中，完成 EC2/RDS/S3/Lambda 任一類型的清單/詳情/關聯瀏覽，且清單搜尋在 p95 < 1s。  
- **SC‑002（MVP）**：最少 3 個鍵盤快捷鍵可完成清單切換、搜尋、進入詳情與返回。  
- **SC‑003（P2）**：CloudWatch 指標/日誌在 1.5 秒內完成首次載入（資料可用時）。  
- **SC‑004（P2）**：對 EC2 成功執行 Start/Stop/Reboot 任一操作，並能正確反映最終狀態。  
- **SC‑005（P2）**：成功新增/更新/刪除任一標籤並於 2 秒內回饋結果。  
- **SC‑006（品質）**：單元與關鍵整合測試通過；CI 全綠；主要熱路徑具可觀測性指標。  

--- 

## 備註

- 依憲章「MVP 優先、拒絕過度設計」，第一版聚焦「只讀瀏覽＋關聯」；操作/監控屬第二波增量。  
- 後續擴展（可選）：ELB/ASG/EKS、更多 CloudWatch Dash、跨帳號切換、資源圖視覺化。  


