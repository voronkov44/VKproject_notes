package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
)

// type для index for
type Article struct {
	Id                     uint16
	Title, Anons, FullText string
}

var posts = []Article{}
var showPost = Article{}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	//Подключение к базе данных
	db, err := sql.Open("mysql", "root:root@tcp(mariadb:3306)/golang")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	//Выборка данных
	res, err := db.Query("SELECT  * FROM `articles`")
	if err != nil {
		panic(err)
	}

	//цикл
	posts = []Article{}
	for res.Next() {
		var post Article
		err = res.Scan(&post.Id, &post.Title, &post.Anons, &post.FullText)
		if err != nil {
			panic(err)
		}

		posts = append(posts, post)

		//вывод в терминал
		//fmt.Println(fmt.Sprintf("Post: %s with id: %d", post.Title, post.Id))
	}

	//Динамическое подключение
	tmpl.ExecuteTemplate(w, "index", posts)
}

func create(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/create.html", "templates/header2.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	//Динамическое подключение
	tmpl.ExecuteTemplate(w, "create", nil)
}

func save_article(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	anons := r.FormValue("anons")
	full_text := r.FormValue("full_text")

	if title == "" || anons == "" || full_text == "" {
		fmt.Fprintf(w, "Не все данные заполнены")
	} else {

		//Подключение к базе данных
		db, err := sql.Open("mysql", "root:root@tcp(mariadb:3306)/golang")
		if err != nil {
			panic(err)
		}

		defer db.Close()

		// Собираем данные от юзера
		insert, err := db.Query(fmt.Sprintf("INSERT INTO `articles` (`title`, `anons`, `full_text`) VALUES('%s', '%s', '%s')", title, anons, full_text))
		if err != nil {
			panic(err)
		}
		defer insert.Close()

		//Переадрессация на страницу
		http.Redirect(w, r, "/index2", http.StatusSeeOther)
	}
}
func show_post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	tmpl, err := template.ParseFiles("templates/show.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	db, err := sql.Open("mysql", "root:root@tcp(mariadb:3306)/golang")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	//Выборка данных
	res, err := db.Query(fmt.Sprintf("SELECT  * FROM `articles` WHERE `id` = '%s'", vars["post_id"]))
	if err != nil {
		panic(err)
	}

	//цикл
	showPost = Article{}
	for res.Next() {
		var post Article
		err = res.Scan(&post.Id, &post.Title, &post.Anons, &post.FullText)
		if err != nil {
			panic(err)
		}

		showPost = post

		//вывод в терминал
		//fmt.Println(fmt.Sprintf("Post: %s with id: %d", post.Title, post.Id))
	}

	//Динамическое подключение
	tmpl.ExecuteTemplate(w, "show", showPost)

}

func login(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/login.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	tmpl.ExecuteTemplate(w, "login", nil)
}

func reglogin(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/reglogin.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	tmpl.ExecuteTemplate(w, "reglogin", nil)
}

func save_users(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	surname := r.FormValue("surname")
	login := r.FormValue("login")
	password := r.FormValue("password")

	if name == "" || surname == "" || login == "" || password == "" {
		fmt.Fprintf(w, "Не все данные заполнены")
	} else {

		//Подключение к базе данных
		db, err := sql.Open("mysql", "root:root@tcp(mariadb:3306)/golang")
		if err != nil {
			panic(err)
		}

		defer db.Close()

		hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}

		// Собираем данные от юзера
		insert, err := db.Prepare("INSERT INTO users (login, hashed_password, name, surname) VALUES (?, ?, ?, ?)")
		if err != nil {
			panic(err)
		}

		defer insert.Close()

		_, err = insert.Exec(login, hashed_password, name, surname)
		if err != nil {
			log.Fatal(err)
		}

		//Переадрессация на страницу
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func examination(w http.ResponseWriter, r *http.Request) {

	db, err := sql.Open("mysql", "root:root@tcp(mariadb:3306)/golang")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Fatal(err)
		}

		login := r.FormValue("login")
		password := r.FormValue("password")

		var hashedPassword string
		row := db.QueryRow("SELECT hashed_password FROM users WHERE login = ?", login)
		err = row.Scan(&hashedPassword)
		if err != nil {
			fmt.Println("Пользователь с таким логином не найден")
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

		if err != nil {
			fmt.Println("Неверный пароль")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}

		fmt.Println("Вход выполнен успешно")
		http.Redirect(w, r, "/index2", http.StatusSeeOther)

	}

}

func index2(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index2.html", "templates/header2.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	//Подключение к базе данных
	db, err := sql.Open("mysql", "root:root@tcp(mariadb:3306)/golang")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	//Выборка данных
	res, err := db.Query("SELECT  * FROM `articles`")
	if err != nil {
		panic(err)
	}

	//цикл
	posts = []Article{}
	for res.Next() {
		var post Article
		err = res.Scan(&post.Id, &post.Title, &post.Anons, &post.FullText)
		if err != nil {
			panic(err)
		}

		posts = append(posts, post)

		//вывод в терминал
		//fmt.Println(fmt.Sprintf("Post: %s with id: %d", post.Title, post.Id))
	}

	//Динамическое подключение
	tmpl.ExecuteTemplate(w, "index2", posts)
}

// Отслеживание URL адрессов
func handleFunc() {

	//rtr := mux.NewRouter()
	rtr := mux.NewRouter().StrictSlash(true)

	rtr.HandleFunc("/", index).Methods("GET")
	rtr.HandleFunc("/create", create).Methods("GET")
	rtr.HandleFunc("/save_article", save_article).Methods("POST")
	rtr.HandleFunc("/post/{post_id:[0-9]+}", show_post).Methods("GET") //Обработка URL страниц post, осторожно с пробелами при работе в {}
	rtr.HandleFunc("/login", login).Methods("GET")
	rtr.HandleFunc("/reglogin", reglogin).Methods("GET")
	rtr.HandleFunc("/save_users", save_users).Methods("POST")
	rtr.HandleFunc("/examination", examination).Methods("GET", "POST")
	rtr.HandleFunc("/index2", index2).Methods("GET")

	http.Handle("/", rtr)
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func main() {
	handleFunc()
}
