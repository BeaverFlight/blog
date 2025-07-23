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

// # Создание новой статьи
//
// Требует аутентификации.
//
// responses:
//
//	201: Response
//	400: Response
//	401: Response
//	500: Response
//
// Параметры:
//   - name: article
//     in: body
//     required: true
//     schema:
//     $ref: "#/definitions/Article"
func CreateArticle(rw http.ResponseWriter, r *http.Request) {
	login, ok := r.Context().Value("login").(string)
	if !ok {
		models.ResponseUnauthorized(rw)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		models.ResponseErrorServer(rw)
		return
	}

	article := models.Article{}
	err = json.Unmarshal(data, &article)
	if err != nil {
		log.Println(err)
		models.ResponseBadRequest(rw)
		return
	}

	ch := make(chan error, 1)
	DB.CreateArticle(login, article.Text, ch)
	err = <-ch
	if err != nil {
		models.ResponseErrorServer(rw)
		return
	}
	models.ResponseCreated(rw)
}

// swagger:route DELETE /article/{id} article deleteArticle
//
// # Удаление статьи
//
// Требует аутентификации и проверки владельца.
//
// responses:
//
//	200: Response
//	400: Response
//	401: Response
//	403: Response
//	500: Response
//
// Параметры:
//   - name: id
//     in: path
//     required: true
//     type: integer
func DeleteArticle(rw http.ResponseWriter, r *http.Request) {
	login, ok := r.Context().Value("login").(string)
	if !ok {
		models.ResponseUnauthorized(rw)
		return
	}
	vars := mux.Vars(r)
	strId := vars["id"]
	id, err := strconv.Atoi(strId)
	if err != nil {
		models.ResponseBadRequest(rw)
		return
	}

	ok, err = DB.VerifyArticleToUser(id, login)
	if err != nil {
		models.ResponseErrorServer(rw)
		return
	}

	if !ok {
		models.ResponseNew(rw, "Вы не можете изменять не свои записи", http.StatusBadRequest)
		return
	}

	ch := make(chan error, 1)
	DB.DeleteArticle(id, ch)
	err = <-ch
	if err != nil {
		models.ResponseErrorServer(rw)
		return
	}
	models.ResponseOK(rw)
}

// swagger:route POST /register user register
//
// # Регистрация пользователя
//
// responses:
//
//	201: Response
//	400: Response
//	409: Response
//	500: Response
//
// Параметры:
//   - name: user
//     in: body
//     required: true
//     schema:
//     $ref: "#/definitions/User"
func Register(rw http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		models.ResponseErrorServer(rw)
		return
	}

	user := models.User{}
	err = json.Unmarshal(data, &user)
	if err != nil {
		models.ResponseBadRequest(rw)
		return
	}

	if len(user.Login) < 5 || len(user.Password) < 8 {
		models.ResponseNew(
			rw,
			"Логин не может быть короче 5 символов, а пароль короче 8 символов",
			http.StatusBadRequest,
		)
		return
	}

	if len(user.Login) >= 16 || len(user.Password) >= 20 {
		models.ResponseNew(
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
		models.ResponseErrorServer(rw)
		return
	}
	models.ResponseCreated(rw)
}

// swagger:route GET /article/{id} article getArticle
// ...
// responses:
//   200: articleResponse
//   400: Response
//   404: Response
//   500: Response

// swagger:route GET /article article getAllArticles
// ...
// responses:
//
//	200: articlesResponse
//	500: Response
func GetArticle(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId := vars["id"]
	id, err := strconv.Atoi(strId)
	if err != nil {
		models.ResponseNotFound(rw)
		return
	}

	articles, err := DB.GetArticle(id)
	if err != nil {
		models.ResponseErrorServer(rw)
		return
	}

	encoder := json.NewEncoder(rw)
	err = encoder.Encode(articles)
	if err != nil {
		models.ResponseErrorServer(rw)
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
//	500: Response
func GetAllArticle(rw http.ResponseWriter, r *http.Request) {
	article, err := DB.GetAllArticle()
	if err != nil {
		models.ResponseNotFound(rw)
		return
	}

	encoder := json.NewEncoder(rw)
	err = encoder.Encode(article)
	if err != nil {
		models.ResponseErrorServer(rw)
		return
	}
}

// swagger:response articlesResponse
type ArticlesResponse struct {
	// in:body
	Body []models.Article
}

// swagger:route PUT /article article updateArticle
//
// # Обновление статьи
//
// Требует аутентификации и проверки владельца.
//
// responses:
//
//	200: Response
//	400: Response
//	401: Response
//	403: Response
//	500: Response
//
// Параметры:
//   - name: article
//     in: body
//     required: true
//     schema:
//     $ref: "#/definitions/Article"
func UpdateArticle(rw http.ResponseWriter, r *http.Request) {
	login, ok := r.Context().Value("login").(string)
	if !ok {
		models.ResponseUnauthorized(rw)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		models.ResponseBadRequest(rw)
		return
	}

	article := models.Article{}
	err = json.Unmarshal(data, &article)
	if err != nil {
		log.Println(err)
		models.ResponseBadRequest(rw)
		return
	}

	ok, err = DB.VerifyArticleToUser(article.ID, login)
	if err != nil {
		models.ResponseErrorServer(rw)
		return
	}

	if !ok {
		models.ResponseNew(rw, "Вы не можете изменять не свои записи.", http.StatusBadRequest)
		return
	}

	ch := make(chan error, 1)
	DB.UpdateArticle(article.ID, article.Text, ch)
	err = <-ch
	if err != nil {
		models.ResponseErrorServer(rw)
		return
	}
	models.ResponseOK(rw)
}

// swagger:route POST /login user login
//
// # Аутентификация
//
// responses:
//
//	200: jwtToken
//	400: Response
//	401: Response
//	500: Response
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
		models.ResponseErrorServer(rw)
		return
	}

	loginRequest := models.User{}

	if err := json.Unmarshal(data, &loginRequest); err != nil {
		models.ResponseBadRequest(rw)
		return
	}

	verify, err := DB.VerifyPassword(loginRequest.Login, loginRequest.Password)
	if err != nil || !verify {
		models.ResponseNew(rw, "Не верны пароль или логин", http.StatusUnauthorized)
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
		models.ResponseErrorServer(rw)
		return
	}

	json.NewEncoder(rw).Encode(map[string]string{
		"token": token,
	})
}
