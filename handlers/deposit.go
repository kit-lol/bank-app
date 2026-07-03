package handlers

import (
	"bank-app/models"
	"bank-app/repository"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func DepositHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Проверяем, что запрос пришел методом POST
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		// 2. Получаем ID пользователя из куки
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID, _ := strconv.Atoi(cookie.Value)

		// 3. Получаем сумму из формы
		amountStr := r.FormValue("amount")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil || amount <= 0 {
			// Если сумма введена неверно, просто возвращаем обратно
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		// 4. Обновляем баланс пользователя
		err = repository.UpdateBalance(db, userID, amount)
		if err != nil {
			fmt.Printf("Ошибка обновления баланса: %v\n", err) // Смотрите в консоль VS Code!
			http.Error(w, "Ошибка БД", http.StatusInternalServerError)
			return
		}

		// 5. Записываем транзакцию в историю
		// Передаем nil в качестве deposit_id, так как это прямое пополнение счета
		err = repository.AddTransaction(db, userID, nil, amount, "DEPOSIT")
		if err != nil {
			http.Error(w, "Ошибка записи истории", http.StatusInternalServerError)
			return
		}

		// 6. Успешный редирект
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func OpenDepositPageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем только численные данные из БД
		types, _ := repository.GetDepositTypes(db)

		// Создаем "Мастер-список" с текстами
		data := []struct {
			ID   int
			Name string
			Desc string
		}{
			{1, "TIDAL FLEX", "Единственный вклад, который подстраивается под вас. Пополняйте и снимайте средства без ограничений, сохраняя контроль над своим капиталом. Гибкость без потери статуса."},
			{2, "DEEP SAVER", "Погрузитесь в мир высокой доходности. Пополняйте счет в любое время и наблюдайте, как ваш капитал растет в спокойных водах океана. Средства надежно зафиксированы до полной готовности."},
			{3, "CORAL FIX", "Незыблемая стабильность и фиксированная прибыль. Идеальный выбор для тех, кто ценит предсказуемость и уверенность в завтрашнем дне. Ваш капитал под надежной защитой."},
			{4, "OCEAN ELITE", "Максимальная доходность для избранных. Полная блокировка средств на весь срок — цена, которую вы платите за статус и исключительную прибыль. Только для тех, кто смотрит на горизонт."},
			{5, "HARBOUR GOLD", "Особые условия для клиентов, которые остаются с нами. Достойная ставка, теплое отношение и надежность, проверенная временем. Ваша спокойная бухта в мире финансов."},
		}

		// Обновляем структуру данных перед отправкой в шаблон
		for i := range types {
			for _, item := range data {
				if types[i].ID == item.ID {
					types[i].Name = item.Name
					types[i].Description = item.Desc
				}
			}
		}

		// ========== ДОБАВЛЕНО: поля для карточек ==========
		type ExtendedDepositType struct {
			models.DepositType
			Class string
			Icon  string
			Badge string
		}

		var extendedTypes []ExtendedDepositType
		for _, t := range types {
			ext := ExtendedDepositType{
				DepositType: t,
			}
			switch t.ID {
			case 1:
				ext.Class = "tidal"
				ext.Icon = "🌊"
				ext.Badge = "ГИБКИЙ"
			case 2:
				ext.Class = "deep"
				ext.Icon = "💎"
				ext.Badge = "НАКОПИТЕЛЬНЫЙ"
			case 3:
				ext.Class = "coral"
				ext.Icon = "🏛️"
				ext.Badge = "СТАБИЛЬНЫЙ"
			case 4:
				ext.Class = "ocean"
				ext.Icon = "⚡"
				ext.Badge = "ПРЕМИУМ"
			case 5:
				ext.Class = "harbour"
				ext.Icon = "🏠"
				ext.Badge = "ЛОЯЛЬНЫЙ"
			}
			extendedTypes = append(extendedTypes, ext)
		}
		// ========== КОНЕЦ ДОБАВЛЕННОГО БЛОКА ==========

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl := template.Must(template.ParseFiles("templates/open_deposit.html"))
		tmpl.Execute(w, extendedTypes) // ← изменено с types на extendedTypes
	}
}

func GetUserIDFromSession(db *sql.DB, r *http.Request) int {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return 0
	}

	// Просто переводим строку из куки обратно в число
	userID, _ := strconv.Atoi(cookie.Value)
	return userID
}

func CreateDepositHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ИСПРАВЛЕНО: Теперь метод POST проверяется строго
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
			return
		}

		userID := GetUserIDFromSession(db, r)
		if userID == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		typeID, _ := strconv.Atoi(r.FormValue("type_id"))
		amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
		if err != nil || amount <= 0 {
			http.Error(w, "Некорректная сумма", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var balance float64
		var currencyID int
		var minAmount float64
		var interestRate float64

		// Получаем данные
		err = tx.QueryRow("SELECT balance, currency_id FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&balance, &currencyID)
		err = tx.QueryRow("SELECT min_amount, interest_rate FROM deposit_types WHERE id = $1", typeID).Scan(&minAmount, &interestRate)

		if err != nil {
			// Вывод реальной ошибки в консоль для отладки
			fmt.Println("Ошибка получения данных:", err)
			http.Error(w, "Тип вклада не найден", http.StatusBadRequest)
			return
		}

		// ВАЛИДАЦИЯ: Чтобы открывать любые вклады, проверь баланс в БД.
		// Если пишет "Недостаточно средств", значит amount > balance.
		if amount < minAmount {
			http.Error(w, "Сумма меньше минимальной", http.StatusBadRequest)
			return
		}
		if balance < amount {
			http.Error(w, "Недостаточно средств на балансе", http.StatusBadRequest)
			return
		}

		// Списание
		_, err = tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2", amount, userID)

		// Вставка (убедись, что currencyID у пользователя не NULL)
		_, err = tx.Exec(`
            INSERT INTO deposits (user_id, type_id, amount, interest_rate, currency_id, status, created_at) 
            VALUES ($1, $2, $3, $4, $5, 'ACTIVE', NOW())`,
			userID, typeID, amount, interestRate, currencyID)

		if err != nil {
			fmt.Println("INSERT ERROR:", err) // ПОСМОТРИ ЭТО В КОНСОЛИ!
			http.Error(w, "Ошибка оформления", http.StatusInternalServerError)
			return
		}

		tx.Commit()
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func DepositToExistingHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserIDFromSession(db, r) // Получаем ID юзера
		if userID == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		depositID := r.FormValue("deposit_id")
		amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)

		tx, _ := db.Begin()
		defer tx.Rollback()

		// 1. Списание с баланса
		_, err := tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1", amount, userID)
		if err != nil { /* обработка ошибки */
			return
		}

		// 2. Пополнение вклада
		_, err = tx.Exec("UPDATE deposits SET amount = amount + $1 WHERE id = $2 AND user_id = $3", amount, depositID, userID)

		// 3. Запись в историю
		_, err = tx.Exec(`INSERT INTO transactions (user_id, deposit_id, amount, operation_type, created_at) 
                          VALUES ($1, $2, $3, 'DEPOSIT_TO_EXISTING', NOW())`,
			userID, depositID, amount)

		tx.Commit()
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func WithdrawHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID, _ := strconv.Atoi(cookie.Value)

		// 1. Считываем deposit_id из формы
		depositID, err := strconv.Atoi(r.FormValue("deposit_id"))
		if err != nil {
			http.Error(w, "Неверный ID вклада", http.StatusBadRequest)
			return
		}

		// 2. Считываем сумму
		amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
		if err != nil || amount <= 0 {
			http.Error(w, "Неверная сумма", http.StatusBadRequest)
			return
		}

		// 3. ПЕРЕДАЕМ depositID в сервис
		err = repository.WithdrawFromDeposit(db, userID, strconv.Itoa(depositID), amount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func CloseDepositHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		depositID := r.FormValue("deposit_id")

		// Выполняем закрытие в БД...
		_, err := db.Exec("DELETE FROM deposits WHERE id = $1", depositID)

		if err != nil {
			// Если ошибка, возвращаем код 500 и НЕ делаем редирект
			http.Error(w, "Ошибка удаления вклада", http.StatusInternalServerError)
			return // ОБЯЗАТЕЛЬНО ВЫХОДИМ
		}

		// Если все хорошо — редирект
		// Проверяем, откуда пришел запрос (админ или юзер)
		if r.Header.Get("Referer") != "" {
			http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}
}
