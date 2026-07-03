package utils

import (
	"net/http"
	"text/template"
)

func GetDepositName(typeID int) string {
	switch typeID {
	case 1:
		return "TIDAL FLEX"
	case 2:
		return "DEEP SAVER"
	case 3:
		return "CORAL FIX"
	case 4:
		return "OCEAN ELITE"
	case 5:
		return "HARBOUR GOLD"
	default:
		return "Обычный вклад"
	}
}

func TranslateOperationType(opType string) string {
	switch opType {
	case "DEPOSIT":
		return "Пополнение"
	case "DEPOSIT_TO_EXISTING":
		return "Пополнение вклада"
	case "WITHDRAW":
		return "Снятие со вклада"
	case "CLOSE_DEPOSIT":
		return "Закрытие вклада"
	case "ACCRUAL":
		return "Начисление процентов" // Добавили это
	default:
		return opType
	}
}

// RenderTemplate — универсальная функция для вывода шаблонов
func RenderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	// Парсим layout и конкретный шаблон страницы
	tmpl, err := template.ParseFiles("templates/layout.html", "templates/"+tmplName+".html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Выполняем шаблон
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}
