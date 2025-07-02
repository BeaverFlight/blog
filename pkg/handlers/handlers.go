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
