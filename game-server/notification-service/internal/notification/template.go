package notification

import (
	"bytes"
	"fmt"
	"text/template"
)

type NotificationTemplate struct {
	Title            string
	MessageTemplate  string   // 可以用 {{.Name}} 這種 placeholder
	NotificationType string   // "in_app", "email", "push"
	DataFields       []string // 需要的額外資料欄位
}

var NotificationTemplates = map[string]NotificationTemplate{
	"member.signedup": {
		Title:            "歡迎加入 Cosmic Void!",
		MessageTemplate:  "Hi {{.Name}}，歡迎註冊！你的帳號 {{.Email}} 已成功建立。",
		NotificationType: "in_app",
		DataFields:       []string{"email", "signedUpAt"},
	},
	"match.ended": {
		Title:            "遊戲結束",
		MessageTemplate:  "你的遊戲已結束！{{.Result}}。擊殺: {{.Kills}}, 死亡: {{.Deaths}}",
		NotificationType: "in_app",
		DataFields:       []string{"matchId", "result", "kills", "deaths"},
	},
	"item.acquired": {
		Title:            "獲得新物品",
		MessageTemplate:  "恭喜！你獲得了 {{.ItemName}}",
		NotificationType: "in_app",
		DataFields:       []string{"itemId", "itemName", "rarity"},
	},
}

// RenderTemplate 根據 event type 和 data 渲染通知內容
func RenderTemplate(eventType string, data map[string]any) (title, message string, notificationType string, err error) {
	tmpl, exists := NotificationTemplates[eventType]
	if !exists {
		return "", "", "", fmt.Errorf("template not found for event type: %s", eventType)
	}

	// 使用 Go template 渲染訊息
	t, err := template.New("message").Parse(tmpl.MessageTemplate)
	if err != nil {
		return "", "", "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", "", "", err
	}

	return tmpl.Title, buf.String(), tmpl.NotificationType, nil
}
