package handlers

import (
	"blog/pkg/dbwork"
	"blog/pkg/models"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

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
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	request := models.Request{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	verify, err := DB.VerifyPassword(request.User.Login, request.User.Password)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if verify {
		ch := make(chan error, 1)
		DB.CreateArticle(request.User.Login, request.Article.Text, ch)
		err := <-ch
		if err != nil {
			http.Error(rw, "Ошибка создания записи", http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusCreated)
		return
	}
	rw.WriteHeader(http.StatusUnauthorized)
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
	vars := mux.Vars(r)
	strId := vars["id"]
	id, err := strconv.Atoi(strId)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	request := models.Request{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(request)

	verify, err := DB.VerifyPassword(request.User.Login, request.User.Password)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	ok, err := DB.VerifyArticleToUser(id, request.User.Login)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if verify && ok {
		ch := make(chan error, 1)
		DB.DeleteArticle(id, ch)
		err = <-ch
		if err != nil {
			http.Error(rw, "Ошибка удаления записи", http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusCreated)
		return
	}
	if !ok {
		http.Error(rw, "Вы не можете изменять не свои записи.", http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusUnauthorized)
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

	request := models.Request{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	ch := make(chan error, 1)
	DB.CreateUser(request.User.Login, request.User.Password, ch)
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
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	articles, err := DB.GetArticle(id)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(rw)
	err = encoder.Encode(articles)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
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
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(rw)
	err = encoder.Encode(article)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
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
// # Требует аутентификации и проверки владельца статьи
//
// responses:
//
//	200: description:Статья успешно обновлена
//	400: description:Неверный формат запроса
//	401: description:Неавторизованный доступ
//	403: description:Запрещено (не владелец статьи)
//	500: description:Ошибка сервера
//
// Параметры:
//   - name: body
//     in: body
//     required: true
//     schema:
func UpdateArticle(rw http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	request := models.Request{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(request)

	verify, err := DB.VerifyPassword(request.User.Login, request.User.Password)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	ok, err := DB.VerifyArticleToUser(request.Article.ID, request.User.Login)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if verify && ok {
		ch := make(chan error, 1)
		DB.UpdateArticle(request.Article.ID, request.Article.Text, ch)
		err = <-ch
		if err != nil {
			http.Error(rw, "Ошибка обновления записи", http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusCreated)
		return
	}
	if !ok {
		http.Error(rw, "Вы не можете изменять не свои записи.", http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusUnauthorized)
}
