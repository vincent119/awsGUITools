package keymap

// Entry 描述單一快捷鍵說明。
type Entry struct {
	Key         string
	Description string
}

// Entries 回傳預設快捷鍵列表。
func Entries() []Entry {
	return []Entry{
		{Key: "1", Description: "切換 EC2 清單"},
		{Key: "2", Description: "切換 RDS 清單"},
		{Key: "3", Description: "切換 S3 清單"},
		{Key: "4", Description: "切換 Lambda 清單"},
		{Key: "/", Description: "開啟搜尋輸入"},
		{Key: "g", Description: "重新整理"},
		{Key: "p", Description: "切換 AWS Profile"},
		{Key: "r", Description: "切換 Region"},
		{Key: "t", Description: "切換主題"},
		{Key: "a", Description: "開啟操作面板"},
		{Key: "T", Description: "開啟標籤編輯器"},
		{Key: "Enter", Description: "顯示詳情"},
		{Key: "?", Description: "顯示快捷鍵說明"},
		{Key: "q", Description: "退出程式"},
	}
}
