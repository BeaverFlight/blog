package dbwork

import (
	"blog/pkg/models"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type DataBase interface {
	DeleteArticle(id int)
	CreateArticle(author, text string)
	GetArticle(id int) (models.Article, error)
	UpdateArticle(id int, text string)
	CreateUser(login, password string)
	GetAllArticle() ([]models.Article, error)
	VerifyPassword(login, password string) (bool, error)
	VerifyArticleToUser(id int, login string) (bool, error)
	Run()
}

type PostgresDataBase struct {
	db     *sql.DB
	events chan event
}

type event struct {
	id        int
	userID    int
	eventType eventType
	author    string
	text      string
	login     string
	password  string
}

type eventType byte

const (
	_                     = iota
	eventDelete eventType = iota
	eventCreate
	eventUpdate
	eventCreateUser
)

type postgresDBParams struct {
	DBName   string `json:"dbName"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	SslMode  string `json:"sslmode"`
}

func InitializationDB() (DataBase, error) {
	configFile, err := os.ReadFile("pkg/dbwork/config.json")
	if err != nil {
		return nil, err
	}

	var config postgresDBParams

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}

	connStr := fmt.Sprintf(
		"host=%s dbname=%s user=%s password=%s sslmode=%s",
		config.Host,
		config.DBName,
		config.User,
		config.Password,
		config.SslMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	events := make(chan event, 16)

	postgres := &PostgresDataBase{db: db, events: events}

	err = postgres.verifyTableAndCreate()
	if err != nil {
		return nil, err
	}

	return postgres, nil
}

func (postgres *PostgresDataBase) verifyTableAndCreate() error {
	exists, err := postgres.verifyTableExists("users")
	if err != nil {
		return err
	}
	if !exists {
		createUsersQuery := `CREATE TABLE users(
		id BIGSERIAL PRIMARY KEY,
		login TEXT,
		password VARCHAR
		);`

		postgres.createTable(createUsersQuery)
	}

	exists, err = postgres.verifyTableExists("articles")
	if err != nil {
		return err
	}

	if !exists {
		createArticlesQuery := `CREATE TABLE articles(
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT,
		text TEXT
		);`

		postgres.createTable(createArticlesQuery)
	}
	return nil
}

func (postgres *PostgresDataBase) verifyTableExists(name string) (bool, error) {
	var result string

	rows, err := postgres.db.Query(fmt.Sprintf("SELECT to_regclass('public.%s');", name))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() && result != name {
		rows.Scan(&result)
	}
	return result == name, rows.Err()
}

func (postgres *PostgresDataBase) createTable(createQuery string) error {
	_, err := postgres.db.Exec(createQuery)
	if err != nil {
		return err
	}
	return nil
}

func (postgres *PostgresDataBase) DeleteArticle(id int) {
	postgres.events <- event{eventType: eventDelete, id: id}
}

func (postgres *PostgresDataBase) CreateArticle(author, text string) {
	id, err := postgres.getUserID(author)
	if err != nil {
		log.Println(err)
		return
	}
	postgres.events <- event{eventType: eventCreate, userID: id, text: text}
}

func (postgres *PostgresDataBase) getUserID(login string) (int, error) {
	id := -1
	getUserQuery := `SELECT id FROM users WHERE login=$1`
	rows, err := postgres.db.Query(getUserQuery, login)
	if err != nil {
		return id, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return -1, err
		}
	}
	return id, err
}

func (postgres *PostgresDataBase) VerifyArticleToUser(id int, login string) (bool, error) {
	getQuery := `SELECT articles.id FROM users, articles WHERE users.login=$1 AND users.id = articles.user_id AND articles.id = $2`
	rows, err := postgres.db.Query(getQuery, login, id)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	temp := -1

	for rows.Next() {
		err = rows.Scan(&temp)
		if err != nil {
			return false, err
		}
	}

	return temp == id, nil
}

func (postgres *PostgresDataBase) GetArticle(id int) (models.Article, error) {
	getArticleQuery := `SELECT * FROM articles WHERE id=$1`
	article := models.Article{}

	rows, err := postgres.db.Query(getArticleQuery, id)
	if err != nil {
		return article, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&article.ID, &article.UserID, &article.Text)
		if err != nil {
			return article, err
		}
		name, err := postgres.getUserName(article.UserID)
		if err != nil {
			return article, err
		}
		article.Author = name
	}
	return article, nil
}

func (postgres *PostgresDataBase) GetAllArticle() ([]models.Article, error) {
	getArticleQuery := `SELECT articles.id, articles.text, users.login  FROM articles, users WHERE articles.user_id = users.id`
	articles := make([]models.Article, 0)
	rows, err := postgres.db.Query(getArticleQuery)
	if err != nil {
		return articles, err
	}
	defer rows.Close()

	for rows.Next() {
		temp := models.Article{}
		err = rows.Scan(&temp.ID, &temp.Text, &temp.Author)
		if err != nil {
			return articles, err
		}
		articles = append(articles, temp)
	}
	return articles, nil
}

func (postgres *PostgresDataBase) getUserName(id int) (string, error) {
	getUserQuery := `SELECT login FROM users WHERE id=$1`
	name := "None"
	rows, err := postgres.db.Query(getUserQuery, id)
	if err != nil {
		return name, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			return name, err
		}
	}
	return name, err
}

func (postgres *PostgresDataBase) UpdateArticle(id int, text string) {
	postgres.events <- event{eventType: eventUpdate, id: id, text: text}
}

func (postgres *PostgresDataBase) CreateUser(login, password string) {
	postgres.events <- event{eventType: eventCreateUser, login: login, password: password}
}

func (postgres *PostgresDataBase) Run() {
	go func() {
		defer postgres.db.Close()
		defer log.Println("Управляющая горутина заверишлась")
		for event := range postgres.events {
			switch event.eventType {
			case eventDelete:
				err := postgres.deleteAticleInDB(event.id)
				if err != nil {
					log.Println(err)
				}
			case eventCreate:
				err := postgres.createArticleInDB(event.userID, event.text)
				if err != nil {
					log.Println(err)
				}

			case eventUpdate:
				err := postgres.updateArticleInDB(event.id, event.text)
				if err != nil {
					log.Println(err)
				}
			case eventCreateUser:
				err := postgres.createUserInDB(event.login, event.password)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()
}

func (postgres *PostgresDataBase) deleteAticleInDB(id int) error {
	deleteArticleQuery := `DELETE FROM articles WHERE id = $1`
	_, err := postgres.db.Exec(deleteArticleQuery, id)
	if err != nil {
		return err
	}
	return nil
}

func (postgres *PostgresDataBase) createArticleInDB(userID int, text string) error {
	createArticleQuery := `INSERT INTO articles
	                        (user_id, text)
	                        VALUES($1, $2)`
	_, err := postgres.db.Exec(createArticleQuery, userID, text)
	if err != nil {
		return err
	}
	return nil
}

func (postgres *PostgresDataBase) updateArticleInDB(id int, text string) error {
	updateArticleQuery := `UPDATE articles
	                       SET text=$1
	                       WHERE id=$2`
	_, err := postgres.db.Exec(updateArticleQuery, text, id)
	if err != nil {
		return err
	}
	return nil
}

func (postgres *PostgresDataBase) createUserInDB(login, password string) error {
	createUserQuery := `INSERT INTO users
                     (login, password)
                     VALUES($1, $2);`
	hashPassword := sha256.Sum256([]byte(password))
	hexPassword := hex.EncodeToString(hashPassword[:])
	log.Println(hexPassword)
	_, err := postgres.db.Exec(createUserQuery, login, hexPassword)
	if err != nil {
		return err
	}
	return nil
}

func (postgres *PostgresDataBase) VerifyPassword(login, password string) (bool, error) {
	hashPassword := sha256.Sum256([]byte(password))
	hexPassword := hex.EncodeToString(hashPassword[:])

	getUserQuery := `SELECT password FROM users WHERE login=$1`

	rows, err := postgres.db.Query(getUserQuery, login)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var realPassword string

	for rows.Next() {
		err = rows.Scan(&realPassword)
		if err != nil {
			return false, err
		}
	}

	if realPassword == hexPassword {
		return true, nil
	}
	return false, nil
}
