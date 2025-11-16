package ai

import (
	"alfa-hack-backend/internal/models"
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// GenerateResponse генерирует ответ на основе сообщения пользователя, категории, username, названия бизнеса, специализации и загруженных файлов
func GenerateResponse(message, category, username, businessName, specialization string, files []models.File) (string, error) {
	// Используем бесплатный API (например, Hugging Face Inference API или локальное решение)
	// Для демо используем простую логику с возможностью подключения реального API

	// Чтение содержимого файлов
	fileContents := make([]string, 0)
	for _, file := range files {
		content, err := readFileContent(file.FilePath)
		if err != nil {
			// Логируем ошибку, но продолжаем работу
			fmt.Printf("Ошибка чтения файла %s: %v\n", file.FilePath, err)
			continue
		}
		if len(content) > 0 {
			fileContents = append(fileContents, fmt.Sprintf("Файл: %s\n%s", file.Filename, content))
		}
	}

	fmt.Printf("Загружено файлов: %d, Прочитано содержимого: %d\n", len(files), len(fileContents))

	// Формирование промпта
	prompt := buildPrompt(message, category, username, businessName, specialization, fileContents)

	// Используем ТОЛЬКО OpenRouter API
	openRouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openRouterKey == "" {
		fmt.Println("❌ OPENROUTER_API_KEY не найден! AI не будет работать.")
		fmt.Println("💡 Добавьте OPENROUTER_API_KEY в файл .env")
		return generateSimpleResponse(message, category, username, businessName, specialization, fileContents), nil
	}

	fmt.Println("🤖 Использую OpenRouter API...")
	orResponse, err := callOpenRouter(prompt, openRouterKey)
	if err == nil && orResponse != "" {
		cleaned := cleanAIResponse(orResponse)
		if cleaned != "" {
			fmt.Println("✅ Успешно использован: OpenRouter API")
			return cleaned, nil
		}
	}

	if err != nil {
		fmt.Printf("❌ OpenRouter API не сработал: %v\n", err)
	}

	// Fallback -- шаблонный ответ если API не сработал
	fmt.Println("⚠️  OpenRouter API не сработал, использую шаблонный fallback-ответ")
	return generateSimpleResponse(message, category, username, businessName, specialization, fileContents), nil
}

func buildPrompt(message, category, username, businessName, specialization string, fileContents []string) string {
	var prompt strings.Builder

	// Улучшенный промпт для качественного анализа
	prompt.WriteString("Ты - профессиональный бизнес-консультант с опытом работы с малым бизнесом. Твоя задача - давать конкретные, практические и полезные советы на основе реальных данных.\n\n")

	// Используем информацию о владельце и бизнесе
	if username != "" {
		prompt.WriteString(fmt.Sprintf("ВЛАДЕЛЕЦ БИЗНЕСА: %s\n", username))
	}
	if businessName != "" {
		prompt.WriteString(fmt.Sprintf("НАЗВАНИЕ БИЗНЕСА: %s\n", businessName))
	}
	if specialization != "" {
		prompt.WriteString(fmt.Sprintf("СПЕЦИАЛИЗАЦИЯ БИЗНЕСА: %s\n", specialization))
	}
	if username != "" || businessName != "" || specialization != "" {
		prompt.WriteString("\n")
	}

	if len(fileContents) > 0 {
		prompt.WriteString("═══════════════════════════════════════════════════════\n")
		prompt.WriteString("ДОСТУПНЫЕ ДАННЫЕ О БИЗНЕСЕ:\n")
		prompt.WriteString("═══════════════════════════════════════════════════════\n")
		for i, content := range fileContents {
			prompt.WriteString(fmt.Sprintf("\n[Файл %d]\n", i+1))
			prompt.WriteString(content)
			prompt.WriteString("\n" + strings.Repeat("-", 55) + "\n")
		}
		prompt.WriteString("\n⚠️ КРИТИЧЕСКИ ВАЖНО: Внимательно изучи ВСЕ данные из файлов выше перед формированием ответа!\n\n")
	} else {
		prompt.WriteString("⚠️ ВНИМАНИЕ: Файлы с данными о бизнесе не загружены.\n")
		prompt.WriteString("Если вопрос требует данных из файлов, вежливо попроси пользователя загрузить их.\n\n")
	}

	if category != "" {
		categoryNames := map[string]string{
			"financial": "💰 Финансовый анализ",
			"legal":     "⚖️ Юридические вопросы",
			"hr":        "👥 Управление персоналом",
			"marketing": "📢 Маркетинг и продвижение",
			"growth":    "📈 Рост и развитие бизнеса",
			"reports":   "📊 Анализ отчетов и данных",
		}
		if name, ok := categoryNames[category]; ok {
			prompt.WriteString(fmt.Sprintf("КАТЕГОРИЯ ВОПРОСА: %s\n\n", name))
		}
	}

	prompt.WriteString(fmt.Sprintf("ВОПРОС ВЛАДЕЛЬЦА БИЗНЕСА:\n%s\n\n", message))

	prompt.WriteString("═══════════════════════════════════════════════════════\n")
	prompt.WriteString("ТРЕБОВАНИЯ К ОТВЕТУ:\n")
	prompt.WriteString("═══════════════════════════════════════════════════════\n\n")

	prompt.WriteString("1. КОНКРЕТНОСТЬ:\n")
	prompt.WriteString("   - Используй ТОЧНЫЕ цифры, имена, даты из файлов\n")
	prompt.WriteString("   - Приводи примеры из загруженных данных\n")
	prompt.WriteString("   - Избегай общих фраз без привязки к данным\n\n")

	prompt.WriteString("2. СТРУКТУРИРОВАННОСТЬ:\n")
	prompt.WriteString("   - Начни с краткого вывода/резюме\n")
	prompt.WriteString("   - Используй списки и пункты для читаемости\n")
	prompt.WriteString("   - Выделяй ключевые моменты\n\n")

	prompt.WriteString("3. ПРАКТИЧНОСТЬ:\n")
	prompt.WriteString("   - Давай конкретные рекомендации, которые можно применить\n")
	prompt.WriteString("   - Предлагай шаги для решения проблемы\n")
	if username != "" || businessName != "" || specialization != "" {
		prompt.WriteString("   - Учитывай специфику ")
		if username != "" {
			prompt.WriteString(fmt.Sprintf("бизнеса владельца %s", username))
		}
		if businessName != "" {
			if username != "" {
				prompt.WriteString(fmt.Sprintf(" (\"%s\")", businessName))
			} else {
				prompt.WriteString(fmt.Sprintf("бизнеса \"%s\"", businessName))
			}
		}
		if specialization != "" {
			if username != "" || businessName != "" {
				prompt.WriteString(fmt.Sprintf(" в сфере %s", specialization))
			} else {
				prompt.WriteString(fmt.Sprintf("бизнеса в сфере %s", specialization))
			}
		}
		prompt.WriteString("\n\n")
	} else {
		prompt.WriteString("   - Учитывай специфику малого бизнеса\n\n")
	}

	prompt.WriteString("4. АНАЛИТИЧНОСТЬ:\n")
	prompt.WriteString("   - Сравнивай данные между периодами/категориями\n")
	prompt.WriteString("   - Выявляй тренды и закономерности\n")
	prompt.WriteString("   - Указывай на проблемы и возможности\n\n")

	prompt.WriteString("5. ПРОФЕССИОНАЛИЗМ:\n")
	prompt.WriteString("   - Пиши деловым, но понятным языком\n")
	prompt.WriteString("   - Избегай шаблонных фраз\n")
	prompt.WriteString("   - Будь честным: если данных недостаточно, скажи об этом\n\n")

	prompt.WriteString("6. ФОРМАТ:\n")
	prompt.WriteString("   - Отвечай ТОЛЬКО на русском языке\n")
	prompt.WriteString("   - Используй абзацы для структуры\n")
	prompt.WriteString("   - НЕ повторяй вопрос в начале ответа\n")
	prompt.WriteString("   - Начинай сразу с сути\n\n")

	prompt.WriteString("═══════════════════════════════════════════════════════\n")
	prompt.WriteString("НАЧНИ СВОЙ ОТВЕТ:\n")
	prompt.WriteString("═══════════════════════════════════════════════════════\n")

	return prompt.String()
}

func cleanAIResponse(text string) string {
	if text == "" {
		return ""
	}

	// Убираем пробелы в начале и конце
	text = strings.TrimSpace(text)

	// Убираем повторяющиеся символы (более 3 подряд)
	text = removeRepeatingChars(text, 3)

	// Убираем теги инструкций
	text = strings.ReplaceAll(text, "[INST]", "")
	text = strings.ReplaceAll(text, "[/INST]", "")
	text = strings.ReplaceAll(text, "<s>", "")
	text = strings.ReplaceAll(text, "</s>", "")

	// Убираем повторяющийся промпт
	markers := []string{
		"НАЧНИ СВОЙ ОТВЕТ:",
		"ОТВЕТ:",
		"ВОПРОС ВЛАДЕЛЬЦА БИЗНЕСА:",
		"═══════════════════════════════════════════════════════",
	}

	for _, marker := range markers {
		if idx := strings.Index(text, marker); idx > 0 {
			text = strings.TrimSpace(text[idx+len(marker):])
		}
	}

	// Убираем множественные переносы строк (более 2 подряд)
	text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	text = strings.ReplaceAll(text, "\r\n\r\n\r\n", "\r\n\r\n")

	// Убираем лишние пробелы
	text = strings.ReplaceAll(text, "  ", " ")

	return strings.TrimSpace(text)
}

func removeRepeatingChars(text string, maxRepeat int) string {
	if len(text) == 0 {
		return text
	}

	var result strings.Builder
	var lastChar rune
	count := 0

	for _, char := range text {
		if char == lastChar {
			count++
			if count <= maxRepeat {
				result.WriteRune(char)
			}
		} else {
			count = 1
			result.WriteRune(char)
			lastChar = char
		}
	}

	return result.String()
}

func callHuggingFaceAPI(prompt, apiKey string) (string, error) {
	// Используем актуальные модели Hugging Face (бесплатные, работают без VPN)
	models := []string{
		"mistralai/Mistral-7B-Instruct-v0.2", // Хорошая для инструкций
		"meta-llama/Llama-2-7b-chat-hf",      // Альтернатива
		"google/flan-t5-xxl",                 // Fallback
		"microsoft/DialoGPT-large",           // Для диалогов
	}

	var lastErr error
	for _, modelName := range models {
		// Используем новый endpoint
		url := fmt.Sprintf("https://router.huggingface.co/hf-inference/models/%s", modelName)
		fmt.Printf("DEBUG: Пробую модель: %s\n", modelName)

		result, err := tryModel(url, prompt, apiKey, modelName)
		if err == nil && result != "" {
			fmt.Printf("DEBUG: Успешно использована модель: %s\n", modelName)
			return result, nil
		}
		lastErr = err
		fmt.Printf("DEBUG: Модель %s не сработала: %v\n", modelName, err)
	}

	return "", fmt.Errorf("все модели не сработали: %v", lastErr)
}

func tryModel(url, prompt, apiKey, modelName string) (string, error) {
	fmt.Printf("DEBUG: Вызываю Hugging Face API, модель: %s, длина промпта: %d\n", modelName, len(prompt))

	// Формируем промпт в зависимости от модели
	var formattedPrompt string
	if strings.Contains(modelName, "Mistral") || strings.Contains(modelName, "Llama") {
		// Для инструкционных моделей используем специальный формат
		formattedPrompt = fmt.Sprintf("<s>[INST] %s [/INST]", prompt)
	} else {
		formattedPrompt = prompt
	}

	// Формат запроса для нового API
	payload := map[string]interface{}{
		"inputs": formattedPrompt,
		"parameters": map[string]interface{}{
			"max_new_tokens": 800,
			"temperature":    0.7,
			"top_p":          0.9,
			"do_sample":      true,
		},
		"options": map[string]interface{}{
			"wait_for_model": true, // Ждем загрузки модели если нужно
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 60 * time.Second, // Увеличиваем таймаут для больших моделей
	}

	fmt.Printf("DEBUG: Отправляю запрос к Hugging Face API...\n")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("DEBUG: Статус ответа: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("DEBUG: Размер ответа: %d байт\n", len(body))
	if len(body) > 0 {
		fmt.Printf("DEBUG: Полный ответ API: %s\n", string(body))
	}

	// Проверяем статус код
	if resp.StatusCode != 200 {
		errorMsg := string(body)
		if len(errorMsg) > 200 {
			errorMsg = errorMsg[:200]
		}
		fmt.Printf("ERROR: API вернул ошибку: %d, тело: %s\n", resp.StatusCode, errorMsg)
		return "", fmt.Errorf("API вернул статус %d: %s", resp.StatusCode, errorMsg)
	}

	// Пробуем разные форматы ответа
	var result []map[string]interface{}
	if err := json.Unmarshal(body, &result); err == nil && len(result) > 0 {
		// Формат массива
		fmt.Printf("DEBUG: Ответ в формате массива, элементов: %d\n", len(result))
		for i, item := range result {
			fmt.Printf("DEBUG: Элемент %d: %+v\n", i, item)
			if generatedText, ok := item["generated_text"].(string); ok {
				fmt.Printf("DEBUG: Найден generated_text, длина: %d\n", len(generatedText))
				cleaned := cleanGeneratedText(generatedText)
				fmt.Printf("DEBUG: Очищенный текст, длина: %d, начало: %s\n", len(cleaned), cleaned[:min(100, len(cleaned))])
				return cleaned, nil
			}
		}
	}

	// Пробуем формат одного объекта
	var singleResult map[string]interface{}
	if err := json.Unmarshal(body, &singleResult); err == nil {
		fmt.Printf("DEBUG: Ответ в формате объекта: %+v\n", singleResult)
		if generatedText, ok := singleResult["generated_text"].(string); ok {
			fmt.Printf("DEBUG: Найден generated_text в объекте, длина: %d\n", len(generatedText))
			cleaned := cleanGeneratedText(generatedText)
			return cleaned, nil
		}
		// Пробуем другие возможные поля
		for key, value := range singleResult {
			if str, ok := value.(string); ok && len(str) > 50 {
				fmt.Printf("DEBUG: Найдено текстовое поле '%s', длина: %d\n", key, len(str))
				return cleanGeneratedText(str), nil
			}
		}
	}

	fmt.Printf("ERROR: Не удалось найти generated_text в ответе. Структура: %s\n", string(body))
	return "", fmt.Errorf("не удалось распарсить ответ API")
}

func cleanGeneratedText(text string) string {
	// Убираем лишние части промпта из ответа
	text = strings.TrimSpace(text)

	// Убираем теги инструкций если есть
	text = strings.ReplaceAll(text, "[INST]", "")
	text = strings.ReplaceAll(text, "[/INST]", "")
	text = strings.ReplaceAll(text, "<s>", "")
	text = strings.ReplaceAll(text, "</s>", "")

	// Убираем повторяющийся промпт в начале (модель иногда возвращает промпт + ответ)
	// Ищем где заканчивается промпт и начинается ответ
	if strings.Contains(text, "НАЧНИ СВОЙ АНАЛИЗ И ОТВЕТ:") {
		parts := strings.Split(text, "НАЧНИ СВОЙ АНАЛИЗ И ОТВЕТ:")
		if len(parts) > 1 {
			text = strings.TrimSpace(parts[1])
		}
	}

	// Убираем повторяющиеся части промпта
	if strings.Contains(text, "ВОПРОС ВЛАДЕЛЬЦА БИЗНЕСА:") {
		// Берем только часть после вопроса
		idx := strings.Index(text, "ВОПРОС ВЛАДЕЛЬЦА БИЗНЕСА:")
		if idx > 0 {
			// Ищем где начинается ответ (обычно после инструкций)
			afterQuestion := text[idx:]
			if strings.Contains(afterQuestion, "ИНСТРУКЦИИ ДЛЯ ОТВЕТА:") {
				parts := strings.Split(afterQuestion, "ИНСТРУКЦИИ ДЛЯ ОТВЕТА:")
				if len(parts) > 1 {
					// Берем часть после инструкций
					afterInstructions := strings.Split(parts[1], "НАЧНИ СВОЙ АНАЛИЗ И ОТВЕТ:")
					if len(afterInstructions) > 1 {
						text = strings.TrimSpace(afterInstructions[1])
					} else {
						text = strings.TrimSpace(parts[1])
					}
				}
			}
		}
	}

	return strings.TrimSpace(text)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func generateSimpleResponse(message, category, username, businessName, specialization string, fileContents []string) string {
	var response strings.Builder
	messageLower := strings.ToLower(message)

	// Анализ файлов для поиска информации
	allFileText := strings.Join(fileContents, "\n\n")
	allFileTextLower := strings.ToLower(allFileText)

	fmt.Printf("DEBUG: message='%s', category='%s', files=%d, textLength=%d\n", message, category, len(fileContents), len(allFileText))

	// Финансовые вопросы
	if category == "financial" || strings.Contains(messageLower, "прибыль") || strings.Contains(messageLower, "выручка") || strings.Contains(messageLower, "доход") || strings.Contains(messageLower, "расход") {
		response.WriteString("📊 **Финансовый анализ:**\n\n")

		if len(fileContents) > 0 {
			// Поиск данных о прибыли
			if strings.Contains(allFileTextLower, "прибыль") {
				profitLines := extractLinesContaining(allFileText, []string{"прибыль", "чистая прибыль"})
				if len(profitLines) > 0 {
					response.WriteString("**Анализ прибыли:**\n")
					for _, line := range profitLines {
						if len(line) > 0 && len(line) < 200 {
							response.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(line)))
						}
					}
					response.WriteString("\n")
				}
			}

			// Поиск данных о выручке
			if strings.Contains(allFileTextLower, "выручка") {
				revenueLines := extractLinesContaining(allFileText, []string{"выручка", "общая выручка"})
				if len(revenueLines) > 0 {
					response.WriteString("**Анализ выручки:**\n")
					for _, line := range revenueLines {
						if len(line) > 0 && len(line) < 200 {
							response.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(line)))
						}
					}
					response.WriteString("\n")
				}
			}

			// Поиск данных о росте
			if strings.Contains(allFileTextLower, "рост") {
				growthLines := extractLinesContaining(allFileText, []string{"рост", "увеличил"})
				if len(growthLines) > 0 {
					response.WriteString("**Динамика роста:**\n")
					for i, line := range growthLines {
						if i < 3 && len(line) > 0 && len(line) < 200 {
							response.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(line)))
						}
					}
					response.WriteString("\n")
				}
			}

			// Если ничего не найдено, даем общий ответ
			if !strings.Contains(allFileTextLower, "прибыль") && !strings.Contains(allFileTextLower, "выручка") {
				response.WriteString("Я проанализировал ваши файлы, но не нашел детальной финансовой информации.\n")
				response.WriteString("Пожалуйста, загрузите файлы с отчетами о продажах для более точного анализа.\n\n")
			}
		} else {
			response.WriteString("Для анализа финансовых данных загрузите файлы с отчетами о продажах.\n\n")
		}
	}

	// Вопросы по персоналу
	if category == "hr" || strings.Contains(messageLower, "сотрудник") || strings.Contains(messageLower, "работник") || strings.Contains(messageLower, "персонал") {
		response.WriteString("👥 **Информация о персонале:**\n\n")

		if len(fileContents) > 0 {
			if strings.Contains(allFileTextLower, "сотрудник") || strings.Contains(allFileTextLower, "работник") {
				// Поиск информации о сотрудниках
				employeeInfo := extractEmployeeInfo(allFileText)
				if len(employeeInfo) > 0 {
					response.WriteString(employeeInfo)
				} else {
					response.WriteString("В загруженных файлах найдена информация о сотрудниках, но требуется более детальный анализ.\n\n")
				}
			} else {
				response.WriteString("В загруженных файлах не найдена информация о сотрудниках.\n")
				response.WriteString("Загрузите файл с данными о персонале для получения подробной информации.\n\n")
			}
		} else {
			response.WriteString("Загрузите файл с информацией о сотрудниках для работы с данными о персонале.\n\n")
		}
	}

	// Юридические вопросы
	if category == "legal" {
		response.WriteString("⚖️ **Юридический вопрос:**\n\n")
		response.WriteString("Для точных ответов на юридические вопросы рекомендую проконсультироваться с юристом.\n")
		response.WriteString("Я могу помочь с общими вопросами, но не могу давать юридические консультации.\n\n")
	}

	// Общие вопросы или если категория не определена
	if category == "" || category == "marketing" {
		// Проверяем наличие файлов
		if len(fileContents) > 0 {
			// Анализируем вопрос пользователя
			hasFinancialQuestion := strings.Contains(messageLower, "прибыль") ||
				strings.Contains(messageLower, "выручка") ||
				strings.Contains(messageLower, "доход") ||
				strings.Contains(messageLower, "расход") ||
				strings.Contains(messageLower, "продаж")

			hasEmployeeQuestion := strings.Contains(messageLower, "сотрудник") ||
				strings.Contains(messageLower, "работник") ||
				strings.Contains(messageLower, "персонал")

			hasGrowthQuestion := strings.Contains(messageLower, "как") &&
				(strings.Contains(messageLower, "вырос") || strings.Contains(messageLower, "рост"))

			// Если есть конкретный вопрос, пытаемся найти ответ
			if hasFinancialQuestion {
				financialInfo := extractFinancialInfo(allFileText, messageLower)
				if financialInfo != "" {
					response.WriteString(financialInfo)
				} else {
					response.WriteString("📊 **Финансовый анализ:**\n\n")
					response.WriteString("Я проанализировал ваши файлы, но не нашел точной информации по вашему вопросу.\n")
					response.WriteString("Попробуйте задать вопрос более конкретно, например:\n")
					response.WriteString("- Какая прибыль в ноябре?\n")
					response.WriteString("- Сколько выручки в декабре?\n\n")
				}
			} else if hasEmployeeQuestion {
				employeeInfo := extractEmployeeInfo(allFileText)
				if employeeInfo != "" {
					response.WriteString(employeeInfo)
				} else {
					response.WriteString("👥 **Информация о персонале:**\n\n")
					response.WriteString("В ваших файлах найдена информация о сотрудниках.\n")
					response.WriteString("Задайте конкретный вопрос, например:\n")
					response.WriteString("- Сколько у меня сотрудников?\n")
					response.WriteString("- Какая зарплата у бариста?\n\n")
				}
			} else if hasGrowthQuestion {
				growthInfo := extractGrowthInfo(allFileText)
				if growthInfo != "" {
					response.WriteString(growthInfo)
				} else {
					// Пытаемся найти данные о росте вручную
					if strings.Contains(allFileTextLower, "ноябрь") && strings.Contains(allFileTextLower, "декабрь") {
						response.WriteString("📈 **Анализ роста:**\n\n")
						response.WriteString("Найдены данные за ноябрь и декабрь. Сравниваю показатели...\n\n")

						// Ищем прибыль
						if strings.Contains(allFileTextLower, "прибыль") {
							profitLines := extractLinesContaining(allFileText, []string{"прибыль", "чистая прибыль"})
							for i, line := range profitLines {
								if i < 3 && len(line) < 150 {
									response.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(line)))
								}
							}
							response.WriteString("\n")
						}

						// Ищем выручку
						if strings.Contains(allFileTextLower, "выручка") {
							revenueLines := extractLinesContaining(allFileText, []string{"выручка", "общая выручка"})
							for i, line := range revenueLines {
								if i < 3 && len(line) < 150 {
									response.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(line)))
								}
							}
							response.WriteString("\n")
						}
					} else {
						response.WriteString("Проанализировал ваши файлы. Для точного ответа о росте загрузите отчеты за разные периоды.\n\n")
					}
				}
			} else {
				// Общий ответ, если файлы есть, но вопрос не специфичный
				response.WriteString(fmt.Sprintf("Я проанализировал ваши файлы (%d файлов). ", len(fileContents)))
				response.WriteString("Могу помочь с анализом данных о вашем бизнесе.\n\n")

				// Показываем что найдено
				if strings.Contains(allFileTextLower, "прибыль") || strings.Contains(allFileTextLower, "выручка") {
					response.WriteString("✅ Найдены финансовые данные\n")
				}
				if strings.Contains(allFileTextLower, "сотрудник") || strings.Contains(allFileTextLower, "работник") {
					response.WriteString("✅ Найдена информация о персонале\n")
				}
				if strings.Contains(allFileTextLower, "ноябрь") || strings.Contains(allFileTextLower, "декабрь") {
					response.WriteString("✅ Найдены отчеты за периоды\n")
				}

				response.WriteString("\n**Задайте конкретный вопрос, например:**\n")
				response.WriteString("- Как выросла прибыль?\n")
				response.WriteString("- Сколько у меня сотрудников?\n")
				response.WriteString("- Какая выручка в декабре?\n\n")
			}
		} else {
			// Нет файлов
			response.WriteString("Для более точных ответов загрузите файлы с данными о вашем бизнесе.\n\n")
			response.WriteString("**Рекомендуемые файлы:**\n")
			response.WriteString("- Отчеты о продажах\n")
			response.WriteString("- Информация о сотрудниках\n")
			response.WriteString("- Финансовые отчеты\n\n")
		}
	}

	if businessName != "" || specialization != "" {
		response.WriteString("**Информация о вашем бизнесе:**\n")
		if businessName != "" {
			response.WriteString(fmt.Sprintf("- Название: %s\n", businessName))
		}
		if specialization != "" {
			response.WriteString(fmt.Sprintf("- Специализация: %s\n", specialization))
		}
		response.WriteString("\n")
	}

	return response.String()
}

// Вспомогательные функции для извлечения информации

func extractLinesContaining(text string, keywords []string) []string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		for _, keyword := range keywords {
			if strings.Contains(lineLower, strings.ToLower(keyword)) {
				result = append(result, line)
				break
			}
		}
	}
	return result
}

func extractEmployeeInfo(text string) string {
	var result strings.Builder
	lines := strings.Split(text, "\n")

	employeeCount := 0
	var employees []string

	for i, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "сотрудник") || strings.Contains(lineLower, "работник") {
			employeeCount++
			// Берем следующие несколько строк как информацию о сотруднике
			employeeInfo := ""
			for j := i; j < len(lines) && j < i+5; j++ {
				if strings.TrimSpace(lines[j]) != "" {
					employeeInfo += lines[j] + "\n"
				}
			}
			if len(employeeInfo) < 300 {
				employees = append(employees, employeeInfo)
			}
			if len(employees) >= 3 {
				break
			}
		}
	}

	if employeeCount > 0 {
		result.WriteString("**Найдено информации о сотрудниках:**\n")
		result.WriteString("- Всего сотрудников: найдено упоминаний\n")
		for i, emp := range employees {
			if i < 2 {
				result.WriteString(fmt.Sprintf("\n**Сотрудник %d:**\n%s", i+1, emp))
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

func extractGrowthInfo(text string) string {
	var result strings.Builder
	textLower := strings.ToLower(text)

	// Поиск информации о росте
	if strings.Contains(textLower, "рост") {
		growthLines := extractLinesContaining(text, []string{"рост", "вырос", "увеличил", "сравнение"})
		if len(growthLines) > 0 {
			result.WriteString("📈 **Анализ роста:**\n\n")
			for i, line := range growthLines {
				if i < 5 && len(line) > 0 && len(line) < 200 {
					result.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(line)))
				}
			}
			result.WriteString("\n")
		}
	}

	// Поиск сравнения периодов
	if strings.Contains(textLower, "ноябрь") && strings.Contains(textLower, "декабрь") {
		result.WriteString("📊 **Сравнение периодов:**\n\n")

		// Ищем конкретные числа
		lines := strings.Split(text, "\n")
		var novemberProfit, decemberProfit string
		var novemberRevenue, decemberRevenue string

		for i, line := range lines {
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, "ноябрь") {
				// Ищем прибыль в ноябре
				if strings.Contains(lineLower, "прибыль") {
					for j := i; j < len(lines) && j < i+3; j++ {
						if strings.Contains(strings.ToLower(lines[j]), "прибыль") {
							novemberProfit = strings.TrimSpace(lines[j])
							break
						}
					}
				}
				// Ищем выручку в ноябре
				if strings.Contains(lineLower, "выручка") {
					for j := i; j < len(lines) && j < i+3; j++ {
						if strings.Contains(strings.ToLower(lines[j]), "выручка") {
							novemberRevenue = strings.TrimSpace(lines[j])
							break
						}
					}
				}
			}
			if strings.Contains(lineLower, "декабрь") {
				// Ищем прибыль в декабре
				if strings.Contains(lineLower, "прибыль") {
					for j := i; j < len(lines) && j < i+3; j++ {
						if strings.Contains(strings.ToLower(lines[j]), "прибыль") {
							decemberProfit = strings.TrimSpace(lines[j])
							break
						}
					}
				}
				// Ищем выручку в декабре
				if strings.Contains(lineLower, "выручка") {
					for j := i; j < len(lines) && j < i+3; j++ {
						if strings.Contains(strings.ToLower(lines[j]), "выручка") {
							decemberRevenue = strings.TrimSpace(lines[j])
							break
						}
					}
				}
			}
		}

		if novemberProfit != "" || decemberProfit != "" {
			result.WriteString("**Прибыль:**\n")
			if novemberProfit != "" {
				result.WriteString(fmt.Sprintf("Ноябрь: %s\n", novemberProfit))
			}
			if decemberProfit != "" {
				result.WriteString(fmt.Sprintf("Декабрь: %s\n", decemberProfit))
			}
			result.WriteString("\n")
		}

		if novemberRevenue != "" || decemberRevenue != "" {
			result.WriteString("**Выручка:**\n")
			if novemberRevenue != "" {
				result.WriteString(fmt.Sprintf("Ноябрь: %s\n", novemberRevenue))
			}
			if decemberRevenue != "" {
				result.WriteString(fmt.Sprintf("Декабрь: %s\n", decemberRevenue))
			}
			result.WriteString("\n")
		}

		// Ищем строки с ростом
		growthLines := extractLinesContaining(text, []string{"рост", "вырос", "увеличил", "+"})
		if len(growthLines) > 0 {
			result.WriteString("**Динамика:**\n")
			for i, line := range growthLines {
				if i < 3 && len(line) < 150 {
					result.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(line)))
				}
			}
			result.WriteString("\n")
		}
	}

	if result.Len() == 0 {
		return ""
	}

	return result.String()
}

func extractFinancialInfo(text string, messageLower string) string {
	var result strings.Builder

	result.WriteString("📊 **Финансовый анализ:**\n\n")

	// Ищем прибыль
	if strings.Contains(messageLower, "прибыль") {
		profitLines := extractLinesContaining(text, []string{"прибыль", "чистая прибыль"})
		if len(profitLines) > 0 {
			result.WriteString("**Прибыль:**\n")
			for i, line := range profitLines {
				if i < 5 && len(line) > 0 && len(line) < 200 {
					result.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(line)))
				}
			}
			result.WriteString("\n")
		}
	}

	// Ищем выручку
	if strings.Contains(messageLower, "выручка") || strings.Contains(messageLower, "доход") {
		revenueLines := extractLinesContaining(text, []string{"выручка", "общая выручка", "доход"})
		if len(revenueLines) > 0 {
			result.WriteString("**Выручка:**\n")
			for i, line := range revenueLines {
				if i < 5 && len(line) > 0 && len(line) < 200 {
					result.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(line)))
				}
			}
			result.WriteString("\n")
		}
	}

	// Ищем расходы
	if strings.Contains(messageLower, "расход") {
		expenseLines := extractLinesContaining(text, []string{"расход", "затрат"})
		if len(expenseLines) > 0 {
			result.WriteString("**Расходы:**\n")
			for i, line := range expenseLines {
				if i < 5 && len(line) > 0 && len(line) < 200 {
					result.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(line)))
				}
			}
			result.WriteString("\n")
		}
	}

	if result.Len() < 50 {
		return ""
	}

	return result.String()
}

func readFileContent(filePath string) (string, error) {
	// Определяем тип файла по расширению
	lowerPath := strings.ToLower(filePath)

	// Word документы (.docx)
	if strings.HasSuffix(lowerPath, ".docx") {
		return readDocxFile(filePath)
	}

	// Excel файлы (.xlsx, .xls)
	if strings.HasSuffix(lowerPath, ".xlsx") || strings.HasSuffix(lowerPath, ".xls") {
		return readExcelFile(filePath)
	}

	// Текстовые файлы (.txt, .csv, и т.д.)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Ограничиваем размер для API (увеличиваем лимит для больших файлов)
	maxSize := 10000 // Увеличили лимит для лучшего анализа
	if len(content) > maxSize {
		return string(content[:maxSize]) + "\n\n[Файл обрезан, показаны первые " + fmt.Sprintf("%d", maxSize) + " символов]", nil
	}

	return string(content), nil
}

func readDocxFile(filePath string) (string, error) {
	// .docx это ZIP архив, открываем его
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка открытия Word файла как ZIP: %v", err)
	}
	defer r.Close()

	var result strings.Builder

	// Ищем файл word/document.xml внутри ZIP
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				continue
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// Парсим XML и извлекаем текст
			text := extractTextFromDocxXML(content)
			if text != "" {
				result.WriteString(text)
			}
			break
		}
	}

	text := result.String()
	if text == "" {
		return fmt.Sprintf("[Word документ: %s. Не удалось извлечь текст. Попробуйте сохранить файл как .txt]", filePath), nil
	}

	if len(text) > 15000 {
		return text[:15000] + "\n\n[Текст обрезан, показаны первые 15000 символов]", nil
	}

	return text, nil
}

func extractTextFromDocxXML(xmlContent []byte) string {
	// Простой парсинг XML - ищем текст между тегами <w:t>
	var result strings.Builder
	content := string(xmlContent)

	// Ищем все вхождения <w:t>...</w:t>
	startTag := "<w:t"
	endTag := "</w:t>"

	pos := 0
	for {
		startIdx := strings.Index(content[pos:], startTag)
		if startIdx == -1 {
			break
		}
		startIdx += pos

		// Находим закрывающий тег >
		closeIdx := strings.Index(content[startIdx:], ">")
		if closeIdx == -1 {
			break
		}
		closeIdx += startIdx + 1

		// Находим закрывающий тег </w:t>
		endIdx := strings.Index(content[closeIdx:], endTag)
		if endIdx == -1 {
			break
		}
		endIdx += closeIdx

		// Извлекаем текст между тегами
		text := content[closeIdx:endIdx]
		// Декодируем XML entities
		text = strings.ReplaceAll(text, "&lt;", "<")
		text = strings.ReplaceAll(text, "&gt;", ">")
		text = strings.ReplaceAll(text, "&amp;", "&")
		text = strings.ReplaceAll(text, "&quot;", "\"")
		text = strings.ReplaceAll(text, "&apos;", "'")

		if strings.TrimSpace(text) != "" {
			result.WriteString(strings.TrimSpace(text))
			result.WriteString(" ")
		}

		pos = endIdx + len(endTag)
	}

	return result.String()
}

func readExcelFile(filePath string) (string, error) {
	// .xlsx это ZIP архив, открываем его
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка открытия Excel файла как ZIP: %v", err)
	}
	defer r.Close()

	var result strings.Builder

	// Ищем файлы xl/sharedStrings.xml и xl/worksheets/sheet*.xml
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "xl/sharedStrings.xml") || strings.HasPrefix(f.Name, "xl/worksheets/sheet") {
			rc, err := f.Open()
			if err != nil {
				continue
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// Парсим XML и извлекаем текст
			text := extractTextFromExcelXML(content)
			if text != "" {
				result.WriteString(text)
				result.WriteString("\n")
			}
		}
	}

	text := result.String()
	if text == "" {
		return fmt.Sprintf("[Excel файл: %s. Не удалось извлечь данные. Попробуйте экспортировать в .csv или .txt]", filePath), nil
	}

	if len(text) > 15000 {
		return text[:15000] + "\n\n[Данные обрезаны, показаны первые 15000 символов]", nil
	}

	return text, nil
}

func extractTextFromExcelXML(xmlContent []byte) string {
	// Парсим XML Excel - ищем текст в тегах <t> или <v>
	var result strings.Builder
	content := string(xmlContent)

	// Ищем все вхождения <t>...</t> (текст) и <v>...</v> (значения)
	patterns := []string{"<t>", "</t>", "<v>", "</v>"}

	pos := 0
	for pos < len(content) {
		// Ищем следующий тег
		nextTag := -1
		tagType := -1

		for i, pattern := range patterns {
			idx := strings.Index(content[pos:], pattern)
			if idx != -1 && (nextTag == -1 || idx < nextTag) {
				nextTag = idx
				tagType = i
			}
		}

		if nextTag == -1 {
			break
		}

		nextTag += pos

		// Если это открывающий тег <t> или <v>
		if tagType == 0 || tagType == 2 {
			closeTag := patterns[tagType+1]
			closeIdx := strings.Index(content[nextTag+len(patterns[tagType]):], closeTag)
			if closeIdx != -1 {
				closeIdx += nextTag + len(patterns[tagType])
				text := content[nextTag+len(patterns[tagType]) : closeIdx]

				// Декодируем XML entities
				text = strings.ReplaceAll(text, "&lt;", "<")
				text = strings.ReplaceAll(text, "&gt;", ">")
				text = strings.ReplaceAll(text, "&amp;", "&")
				text = strings.ReplaceAll(text, "&quot;", "\"")
				text = strings.ReplaceAll(text, "&apos;", "'")

				if strings.TrimSpace(text) != "" {
					result.WriteString(strings.TrimSpace(text))
					result.WriteString(" | ")
				}

				pos = closeIdx + len(closeTag)
				continue
			}
		}

		pos = nextTag + len(patterns[tagType])
	}

	text := result.String()
	// Убираем последний разделитель
	text = strings.TrimSuffix(text, " | ")

	return text
}

func callAioNet(prompt, apiKey string) (string, error) {
	url := "https://api.ai.io.net/v1/chat/completions"
	payload := map[string]interface{}{
		"model": "io-nexus-70b-chat",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  800,
		"temperature": 0.7,
		"top_p":       0.9,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	type aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	var result aiResp
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if len(result.Choices) > 0 {
		return strings.TrimSpace(result.Choices[0].Message.Content), nil
	}
	return "", fmt.Errorf("no choices in ai.io.net response: %s", string(body))
}

func callOpenRouter(prompt, apiKey string) (string, error) {
	url := "https://openrouter.ai/api/v1/chat/completions"

	// Используем ТОЛЬКО полностью бесплатные модели OpenRouter (max_price=0)
	// Обновленный список на основе реально работающих моделей (проверено на практике)
	freeModels := []string{
		"mistralai/mistral-7b-instruct:free",    // Mistral 7B Instruct (free) - ОСНОВНАЯ, работает стабильно
		"google/gemini-2.0-flash-exp:free",      // Gemini 2.0 Flash - работает отлично, быстрая
		"meta-llama/llama-3.2-3b-instruct:free", // Llama 3.2 3B - fallback (может быть rate-limited)
	}

	var lastErr error
	for _, modelName := range freeModels {
		payload := map[string]interface{}{
			"model": modelName,
			"messages": []map[string]string{
				{"role": "user", "content": prompt},
			},
			"max_tokens":  2000,
			"temperature": 0.7,
			"top_p":       0.9,
		}

		result, err := tryOpenRouterModel(url, payload, apiKey, modelName)
		if err == nil && result != "" {
			fmt.Printf("DEBUG: Успешно использована модель OpenRouter: %s\n", modelName)
			return result, nil
		}
		lastErr = err
		fmt.Printf("DEBUG: Модель OpenRouter %s не сработала: %v\n", modelName, err)
	}

	return "", fmt.Errorf("все модели OpenRouter не сработали: %v", lastErr)
}

func tryOpenRouterModel(url string, payload map[string]interface{}, apiKey, modelName string) (string, error) {
	fmt.Printf("DEBUG: Пробую модель OpenRouter: %s\n", modelName)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://alfa-hack.com")
	req.Header.Set("X-Title", "AlfaChatDemo")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Проверяем статус код
	if resp.StatusCode != 200 {
		errorMsg := string(body)
		if len(errorMsg) > 200 {
			errorMsg = errorMsg[:200]
		}
		fmt.Printf("ERROR: OpenRouter API вернул ошибку: %d, тело: %s\n", resp.StatusCode, errorMsg)
		return "", fmt.Errorf("OpenRouter API вернул статус %d: %s", resp.StatusCode, errorMsg)
	}

	type aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	var result aiResp
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("ошибка парсинга ответа OpenRouter: %v, тело: %s", err, string(body))
	}

	if result.Error != nil {
		return "", fmt.Errorf("OpenRouter API ошибка: %s (тип: %s)", result.Error.Message, result.Error.Type)
	}

	if len(result.Choices) > 0 {
		return strings.TrimSpace(result.Choices[0].Message.Content), nil
	}
	return "", fmt.Errorf("no choices in OpenRouter response: %s", string(body))
}

func callGroq(prompt, apiKey string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"
	// Используем актуальные модели Groq (пробуем несколько по очереди)
	// llama-3.1-70b-versatile была снята с поддержки, используем актуальные модели
	models := []string{
		"llama-3.1-8b-instant",    // Быстрая и надежная модель (основная)
		"mixtral-8x7b-32768",      // Альтернатива Mixtral
		"llama-3.3-70b-versatile", // Новая версия (если доступна)
		"llama-3.1-70b-versatile", // Старая (пробуем на всякий случай)
	}

	var lastErr error
	for _, modelName := range models {
		payload := map[string]interface{}{
			"model": modelName,
			"messages": []map[string]string{
				{"role": "user", "content": prompt},
			},
			"max_tokens":  2000,
			"temperature": 0.7,
			"top_p":       0.9,
		}

		result, err := tryGroqModel(url, payload, apiKey, modelName)
		if err == nil && result != "" {
			fmt.Printf("DEBUG: Успешно использована модель: %s\n", modelName)
			return result, nil
		}
		lastErr = err
		fmt.Printf("DEBUG: Модель %s не сработала: %v\n", modelName, err)
	}

	return "", fmt.Errorf("все модели не сработали: %v", lastErr)
}

func tryGroqModel(url string, payload map[string]interface{}, apiKey, modelName string) (string, error) {
	fmt.Printf("DEBUG: Пробую модель Groq: %s\n", modelName)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Проверяем статус код
	if resp.StatusCode != 200 {
		errorMsg := string(body)
		if len(errorMsg) > 200 {
			errorMsg = errorMsg[:200]
		}
		fmt.Printf("ERROR: Groq API вернул ошибку: %d, тело: %s\n", resp.StatusCode, errorMsg)
		return "", fmt.Errorf("groq API вернул статус %d: %s", resp.StatusCode, errorMsg)
	}

	type aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	var result aiResp
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("ошибка парсинга ответа Groq: %v, тело: %s", err, string(body))
	}

	if result.Error != nil {
		return "", fmt.Errorf("groq API ошибка: %s (тип: %s)", result.Error.Message, result.Error.Type)
	}

	if len(result.Choices) > 0 {
		return strings.TrimSpace(result.Choices[0].Message.Content), nil
	}
	return "", fmt.Errorf("no choices in Groq response: %s", string(body))
}
