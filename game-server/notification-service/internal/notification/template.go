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
		Title:            "Welcome to Cosmic Void!",
		MessageTemplate:  "Hi {{.Name}}, welcome! Your account {{.Email}} has been created.",
		NotificationType: "in_app",
		DataFields:       []string{"email", "signedUpAt"},
	},
	"match.ended": {
		Title:            "Game Over",
		MessageTemplate:  "Your game has ended! {{.Result}}. Kills: {{.Kills}}, Deaths: {{.Deaths}}",
		NotificationType: "in_app",
		DataFields:       []string{"matchId", "result", "kills", "deaths"},
	},
	"item.acquired": {
		Title:            "New Item Acquired",
		MessageTemplate:  "Congratulations! You obtained {{.ItemName}}",
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
