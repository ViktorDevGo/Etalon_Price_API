package email

import (
	"fmt"
	"time"
)

// NomenclatureSyncResult contains sync statistics
type NomenclatureSyncResult struct {
	Type            string // "tyres" or "rims"
	NewCount        int
	SkippedCount    int
	TotalInDB       int
	Duration        time.Duration
	StartedAt       time.Time
	CompletedAt     time.Time
	FilteredOutType string // e.g., "Грузовая, Спецшина"
	FilteredCount   int
}

// BuildNomenclatureSyncMessage creates email message for nomenclature sync
func BuildNomenclatureSyncMessage(result NomenclatureSyncResult) Message {
	var typeRu string
	switch result.Type {
	case "tyres":
		typeRu = "Шины"
	case "rims":
		typeRu = "Диски"
	default:
		typeRu = result.Type
	}

	subject := fmt.Sprintf("✅ Синхронизация номенклатуры: %s - обновлено", typeRu)

	body := fmt.Sprintf(`Добрый день!

Синхронизация номенклатуры завершена успешно.

Тип: %s
Дата: %s
Время выполнения: %s

Результаты:
───────────────────────────────────────
• Добавлено новых записей: %d
• Пропущено (дубли по артикулу): %d
• Всего записей в БД: %d

Детали:
───────────────────────────────────────
Начало: %s
Завершение: %s
`,
		typeRu,
		result.CompletedAt.Format("02.01.2006"),
		formatDuration(result.Duration),
		result.NewCount,
		result.SkippedCount,
		result.TotalInDB,
		result.StartedAt.Format("15:04:05"),
		result.CompletedAt.Format("15:04:05"),
	)

	if result.FilteredCount > 0 {
		body += fmt.Sprintf("\nОтфильтровано при загрузке (%s): %d\n",
			result.FilteredOutType,
			result.FilteredCount,
		)
	}

	body += "\n───────────────────────────────────────\n"
	body += "Система: Etalon Price API\n"
	body += "Автоматическое уведомление, не отвечайте на это письмо.\n"

	return Message{
		To:      []string{"v.boyarkin@etalon-shina.ru"},
		Subject: subject,
		Body:    body,
		HTML:    false,
	}
}

// BuildNomenclatureSyncHTMLMessage creates HTML email message
func BuildNomenclatureSyncHTMLMessage(result NomenclatureSyncResult) Message {
	var typeRu string
	switch result.Type {
	case "tyres":
		typeRu = "Шины"
	case "rims":
		typeRu = "Диски"
	default:
		typeRu = result.Type
	}

	subject := fmt.Sprintf("✅ Синхронизация номенклатуры: %s - обновлено", typeRu)

	// Status icon and color
	statusIcon := "✅"
	statusColor := "#28a745"
	if result.NewCount == 0 {
		statusIcon = "ℹ️"
		statusColor = "#17a2b8"
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: %s; color: white; padding: 20px; border-radius: 5px 5px 0 0; }
        .content { background: #f8f9fa; padding: 20px; border: 1px solid #dee2e6; }
        .stats { background: white; padding: 15px; margin: 15px 0; border-radius: 5px; }
        .stat-row { display: flex; justify-content: space-between; margin: 10px 0; padding: 8px; }
        .stat-row:nth-child(odd) { background: #f8f9fa; }
        .stat-label { font-weight: bold; }
        .stat-value { color: %s; font-weight: bold; }
        .footer { background: #e9ecef; padding: 15px; border-radius: 0 0 5px 5px; font-size: 12px; color: #6c757d; text-align: center; }
        .highlight { background: #fff3cd; padding: 10px; border-left: 4px solid #ffc107; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>%s Синхронизация номенклатуры</h2>
            <p>%s | %s</p>
        </div>

        <div class="content">
            <h3>Результаты синхронизации</h3>

            <div class="stats">
                <div class="stat-row">
                    <span class="stat-label">Тип номенклатуры:</span>
                    <span>%s</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Добавлено новых записей:</span>
                    <span class="stat-value">%d</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Пропущено (дубли):</span>
                    <span>%d</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Всего в базе данных:</span>
                    <span class="stat-value">%d</span>
                </div>
                <div class="stat-row">
                    <span class="stat-label">Время выполнения:</span>
                    <span>%s</span>
                </div>
            </div>

            %s

            <div style="margin-top: 15px; font-size: 14px;">
                <strong>Период выполнения:</strong><br>
                Начало: %s<br>
                Завершение: %s
            </div>
        </div>

        <div class="footer">
            <p>Система: Etalon Price API<br>
            Автоматическое уведомление, не отвечайте на это письмо.</p>
        </div>
    </div>
</body>
</html>`,
		statusColor,
		statusColor,
		statusIcon,
		typeRu,
		result.CompletedAt.Format("02.01.2006"),
		typeRu,
		result.NewCount,
		result.SkippedCount,
		result.TotalInDB,
		formatDuration(result.Duration),
		buildFilteredMessage(result),
		result.StartedAt.Format("15:04:05"),
		result.CompletedAt.Format("15:04:05"),
	)

	return Message{
		To:      []string{"v.boyarkin@etalon-shina.ru"},
		Subject: subject,
		Body:    body,
		HTML:    true,
	}
}

func buildFilteredMessage(result NomenclatureSyncResult) string {
	if result.FilteredCount > 0 {
		return fmt.Sprintf(`<div class="highlight">
			<strong>ℹ️ Отфильтровано при загрузке:</strong><br>
			%s: %d записей
		</div>`, result.FilteredOutType, result.FilteredCount)
	}
	return ""
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d сек", int(d.Seconds()))
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d мин %d сек", minutes, seconds)
}
