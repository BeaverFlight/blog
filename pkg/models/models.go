package models

// Article представляет контентную публикацию
// swagger:model article
type Article struct {
	// Уникальный идентификатор статьи
	// required: true
	// example: 1
	ID int `json:"id"`

	// Отображаемое имя автора
	// required: true
	// example: Иван Иванов
	Author string `json:"author"`

	// ID пользователя-владельца статьи
	// required: true
	// example: 5
	UserID int `json:"user_id"`

	// Основное содержимое статьи
	// required: true
	// example: Текст статьи...
	Text string `json:"text"`
}

// User представляет учётную запись пользователя
// swagger:model user
type User struct {
	// Уникальный идентификатор пользователя
	// required: true
	// example: 5
	ID int `json:"id"`

	// Логин для аутентификации
	// required: true
	// example: user123
	Login string `json:"login"`

	// Пароль учётной записи
	// required: true
	// swagger:strfmt password
	// example: mySecretPassword
	Password string `json:"password"`
}

// Request представляет составной входной объект
// swagger:model request
type Request struct {
	// Данные статьи
	// required: true
	Article Article `json:"article"`

	// Учётные данные пользователя
	// required: true
	User User `json:"user"`
}
