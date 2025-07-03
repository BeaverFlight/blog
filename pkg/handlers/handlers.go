package handlers

import (
	"blog/pkg/auth"
	"blog/pkg/dbwork"
	"blog/pkg/models"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var DB dbwork.DataBase

func StartDB() error {
	var err error

	DB, err = dbwork.InitializationDB()
	if err != nil {
		return err
	}

	DB.Run()
	return nil
}

// Стандартная ошибка API
// swagger:model
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// swagger:route POST /article article createArticle
//
// # Создание новой статьи
//
// # Требует аутентификации пользователя
//
// responses:
//
//	201: description:Статья успешно создана
//	400: description:Неверный формат запроса
//	401: description:Неавторизованный доступ
//	500: description:Ошибка сервера
//
// Параметры:
//   - name: body
//     in: body
//     required: true
//     schema:
//     $ref: "#/definitions/Request"
func CreateArticle(rw http.ResponseWriter, r *http.Request) {
	login, ok := r.Context().Value("login").(string)
	if !ok {
		http.Error(rw, "Вы не авторизованы", http.StatusUnauthorized)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(rw, "Ошибка чтения запроса", http.StatusInternalServerError)
		return
	}

	article := models.Article{}
	err = json.Unmarshal(data, &article)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	ch := make(chan error, 1)
	DB.CreateArticle(login, article.Text, ch)
	err = <-ch
	if err != nil {
		http.Error(rw, "Ошибка создания записи", http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

// swagger:route DELETE /article/{id} article deleteArticle
//
// # Удаление статьи
//
// # Требует аутентификации и проверки владельца статьи
//
// responses:
//
//	200: description:Статья успешно удалена
//	400: description:Неверный ID статьи
//	401: description:Неавторизованный доступ
//	403: description:Запрещено (не владелец статьи)
//	500: description:Ошибка сервера
//
// Параметры:
//   - name: id
//     in: path
//     description: ID статьи
//     required: true
//     type: integer
//     format: int64
//   - name: body
//     in: body
//     required: true
//     schema:
//     $ref: "#/definitions/Request"
func DeleteArticle(rw http.ResponseWriter, r *http.Request) {
	login, ok := r.Context().Value("login").(string)
	if !ok {
		http.Error(rw, "Вы не авторизованы", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	strId := vars["id"]
	id, err := strconv.Atoi(strId)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	ok, err = DB.VerifyArticleToUser(id, login)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if !ok {
		http.Error(rw, "Вы не можете изменять не свои записи", http.StatusBadRequest)
		return
	}

	ch := make(chan error, 1)
	DB.DeleteArticle(id, ch)
	err = <-ch
	if err != nil {
		http.Error(rw, "Ошибка удаления записи", http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

// swagger:route POST /register user register
//
// # Регистрация нового пользователя
//
// responses:
//
//	201: description:Пользователь успешно зарегистрирован
//	400: description:Неверный формат запроса
//	409: description:Пользователь уже существует
//	500: description:Ошибка сервера
//
// Параметры:
//   - name: body
//     in: body
//     required: true
//     schema:
//     $ref: "#/definitions/Request"
func Register(rw http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	user := models.User{}
	err = json.Unmarshal(data, &user)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(user.Login) < 5 || len(user.Password) < 8 {
		http.Error(
			rw,
			"Логин не может быть короче 5 символов, а пароль короче 8 символов",
			http.StatusBadRequest,
		)
		return
	}

	if len(user.Login) >= 16 || len(user.Password) >= 20 {
		http.Error(
			rw,
			"Логин не может быть длиннее 16 символов, а пороль длинее 20 символов",
			http.StatusBadRequest,
		)
		return
	}

	ch := make(chan error, 1)
	DB.CreateUser(user.Login, user.Password, ch)
	err = <-ch
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

// swagger:route GET /article/{id} article getArticle
//
// # Получение статьи по ID
//
// responses:
//
//	200: articleResponse
//	400: description:Неверный ID статьи
//	404: description:Статья не найдена
//	500: description:Ошибка сервера
//
// Параметры:
//   - name: id
//     in: path
//     description: ID статьи
//     required: true
//     type: integer
//     format: int64
func GetArticle(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId := vars["id"]
	id, err := strconv.Atoi(strId)
	if err != nil {
		http.Error(rw, "Данные не найдены", http.StatusNotFound)
		return
	}

	articles, err := DB.GetArticle(id)
	if err != nil {
		http.Error(rw, "Ошибка поиска данных", http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(rw)
	err = encoder.Encode(articles)
	if err != nil {
		http.Error(rw, "Ошибка отправки данных", http.StatusInternalServerError)
		return
	}
}

// swagger:response articleResponse
type ArticleResponse struct {
	// in:body
	Body models.Article
}

// swagger:route GET /article article getAllArticles
//
// # Получение всех статей
//
// responses:
//
//	200: articlesResponse
//	500: description:Ошибка сервера
func GetAllArticle(rw http.ResponseWriter, r *http.Request) {
	article, err := DB.GetAllArticle()
	if err != nil {
		http.Error(rw, "Ошибка поиска данных", http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(rw)
	err = encoder.Encode(article)
	if err != nil {
		http.Error(rw, "Ошибка отправки данных", http.StatusInternalServerError)
		return
	}
}

// swagger:response articlesResponse
type ArticlesResponse struct {
	// in:body
	Body []models.Article
}

// swagger:route PUT /article article updateArticle
// Обновление существующей статьи
// Требует аутентификации и проверки владельца статьи
// responses:
//
//	200: description: Статья успешно обновлена
//	400: description: Неверный формат запроса
//	401: description: Неавторизованный доступ
//	403: description: Запрещено (не владелец статьи)
//	500: description: Ошибка сервера
//
// Параметры:
//   - name: articleData
//     in: body
//     required: true
//     schema:
//     $ref: "#/definitions/Article"
func UpdateArticle(rw http.ResponseWriter, r *http.Request) {
	login, ok := r.Context().Value("login").(string)
	if !ok {
		http.Error(rw, "Вы не авторизованы", http.StatusUnauthorized)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, "Ошибка запроса к серверу", http.StatusInternalServerError)
		return
	}

	article := models.Article{}
	err = json.Unmarshal(data, &article)
	if err != nil {
		log.Println(err)
		http.Error(rw, "Данные в запросе не найдены", http.StatusBadRequest)
		return
	}

	ok, err = DB.VerifyArticleToUser(article.ID, login)
	if err != nil {
		http.Error(rw, "Ошибка запроса", http.StatusInternalServerError)
		return
	}

	if !ok {
		http.Error(rw, "Вы не можете изменять не свои записи.", http.StatusInternalServerError)
		return
	}

	ch := make(chan error, 1)
	DB.UpdateArticle(article.ID, article.Text, ch)
	err = <-ch
	if err != nil {
		http.Error(rw, "Ошибка обновления записи", http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

// swagger:route POST /login user login
// Аутентификация пользователя и получение JWT токена
// responses:
//
//	200: description: Успешная аутентификация
//	400: description: Неверный формат запроса
//	401: description: Неверные учетные данные
//	500: description: Ошибка сервера
//
// Параметры:
//   - name: credentials
//     in: body
//     required: true
//     schema:
//     $ref: "#/definitions/User"
func LoginHandler(rw http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, "Ошибка запроса к серверу", http.StatusBadRequest)
		return
	}

	loginRequest := models.User{}

	if err := json.Unmarshal(data, &loginRequest); err != nil {
		http.Error(rw, "Данные в запросе не найдены", http.StatusBadRequest)
		return
	}

	verify, err := DB.VerifyPassword(loginRequest.Login, loginRequest.Password)
	if err != nil || !verify {
		http.Error(rw, "Не верны пароль или логин", http.StatusUnauthorized)
		return
	}

	configFile, _ := os.ReadFile("JWTSecret.json")
	var config struct {
		JWTSecret          string `json:"jwt_secret"`
		JWTExpirationHours int    `json:"jwt_expiration_hours"`
	}
	json.Unmarshal(configFile, &config)

	token, err := auth.GenerateJWT(
		loginRequest.Login,
		config.JWTSecret,
		time.Duration(config.JWTExpirationHours)*time.Hour,
	)
	if err != nil {
		http.Error(rw, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(rw).Encode(map[string]string{
		"token": token,
	})
}
