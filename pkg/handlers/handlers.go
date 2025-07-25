package handlers

import (
	"blog/pkg/auth"
	"blog/pkg/dbwork"
	"blog/pkg/models"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var logger = log.New(os.Stdout, "[HTTP] ", log.LstdFlags|log.Lshortfile)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logger.Printf("Incoming %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		headers := make(map[string]string)
		for k, v := range r.Header {
			if k != "Authorization" {
				headers[k] = strings.Join(v, ", ")
			}
		}
		logger.Printf("Headers: %+v", headers)

		var bodyBytes []byte
		if r.Body != nil && r.URL.Path != "/login" && r.URL.Path != "/register" {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if len(bodyBytes) > 0 {
				maxBodySize := 1024
				if len(bodyBytes) > maxBodySize {
					logger.Printf("Body (truncated): %s", string(bodyBytes[:maxBodySize]))
				} else {
					logger.Printf("Body: %s", string(bodyBytes))
				}
			}
		}

		lrw := &loggingResponseWriter{ResponseWriter: rw}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		logger.Printf("Completed %s %s | Status: %d | Duration: %v",
			r.Method, r.URL.Path, lrw.status, duration)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
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
	logger.Printf("CreateArticle started")
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
	dbwork.DB.CreateArticle(login, article.Text, ch)
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
	logger.Printf("DeleteArticle started for ID: %s", strId)
	ok, err = dbwork.DB.VerifyArticleToUser(id, login)
	if err != nil {
		models.ResponseErrorServer(rw)
		return
	}

	if !ok {
		models.ResponseNew(rw, "Вы не можете изменять не свои записи", http.StatusBadRequest)
		return
	}

	ch := make(chan error, 1)
	dbwork.DB.DeleteArticle(id, ch)
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
	logger.Printf("Register started for user: %s", user.Login)
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
	dbwork.DB.CreateUser(user.Login, user.Password, ch)
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
	logger.Printf("GetArticle started for ID: %s", strId)
	articles, err := dbwork.DB.GetArticle(id)
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
	logger.Printf("GetAllArticle started")
	article, err := dbwork.DB.GetAllArticle()
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
	logger.Printf("UpdateArticle started for ID: %d", article.ID)
	ok, err = dbwork.DB.VerifyArticleToUser(article.ID, login)
	if err != nil {
		models.ResponseErrorServer(rw)
		return
	}

	if !ok {
		models.ResponseNew(rw, "Вы не можете изменять не свои записи.", http.StatusBadRequest)
		return
	}

	ch := make(chan error, 1)
	dbwork.DB.UpdateArticle(article.ID, article.Text, ch)
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
	logger.Printf("Login attempt for: %s", loginRequest.Login)
	verify, err := dbwork.DB.VerifyPassword(loginRequest.Login, loginRequest.Password)
	if err != nil || !verify {
		models.ResponseNew(rw, "Не верны пароль или логин", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateJWT(
		loginRequest.Login,
	)
	if err != nil {
		models.ResponseErrorServer(rw)
		return
	}

	json.NewEncoder(rw).Encode(map[string]string{
		"token": token,
	})
}
