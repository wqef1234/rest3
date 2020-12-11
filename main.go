package main

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"sort"
	"strconv"

	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"html/template"
	"log"
	"net/http"
	"os"
	_ "github.com/go-sql-driver/mysql"
	//"strconv"
)


const (
	connPort = "8080"
	connHost = "localhost"

	driverName     = "mysql"
	dataSourceName = "root:root@/restdb"

)

type Book struct {
	Id uint       `json:"id"`
	Title string  `json:"title"`
	Author string `json:"author"`
	Rating int    `json:"rating"`
}

type Journal struct {
	Id uint           `json:"id"`
	Title string      `json:"title"`
	Editor string     `json:"editor"`
	PageAmount  int   `json:"pageAmount"`
}

type BookCollection struct {
	Books []Book
}

type JournalCollection struct {
	Journals []Journal
}

var db *sql.DB

var connectionError error
func GetAllJournalFromDB()JournalCollection{
	books := JournalCollection{}
	var query = "SELECT * FROM journals;"
	rows, err := db.Query(query)
	if err != nil {
		log.Println("error while query SELECT (all) executing:", err)
		return books
	}


	for rows.Next(){
		var book Journal

		err = rows.Scan(&book.Id,&book.Title, &book.Editor,&book.PageAmount,)
		if err != nil {
			log.Println("error while scanning values from SELECT query:", err)
			continue
		}

		books.Journals = append(books.Journals, book)
	}
	fmt.Println(books)
	return books
}
func GetAllBooksFromDB()BookCollection{
	books := BookCollection{}
	var query = "SELECT uid,Title, Author, Rating FROM books;"
	rows, err := db.Query(query)
	if err != nil {
		log.Println("error while query SELECT (all) executing:", err)
		return books
	}


	for rows.Next(){
		var book Book

		err = rows.Scan(&book.Id,&book.Title, &book.Author,&book.Rating,)
		if err != nil {
			log.Println("error while scanning values from SELECT query:", err)
			continue
		}

		books.Books = append(books.Books, book)
	}

	return books
}

func GetBookFromId(id uint)Book{
	var book Book
	var query = "SELECT uid,Title, Author, Rating FROM books where uid = "+
		fmt.Sprintf("%d",id)+";"

	rows, err := db.Query(query)
	if err != nil {
		log.Println("error while query SELECT executing:", err)
		return book
	}



	for rows.Next(){

		err = rows.Scan(&book.Id,&book.Title, &book.Author,&book.Rating,)
		if err != nil {
			log.Println("error while scanning values from SELECT query:", err)
			continue
		}

	}

	return book
}
/////cookies
var cookieHandler = securecookie.New(securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

//настроить сессию с пользователем
func SetsSession(userName string,response http.ResponseWriter){
	value := map[string]string{"username":userName}
	encoded, err := cookieHandler.Encode("session",value)
	if err == nil {
		cookie := &http.Cookie{
			Name: "session",
			Value: encoded,
			Path: "/",
		}
		http.SetCookie(response,cookie)
	}
}
//получить имя пользователя из сессии
func GetUserName(request *http.Request)(userName string){
	cookie,err := request.Cookie("session")
	if err == nil {
		cookieValue := make(map[string]string)
		err = cookieHandler.Decode("session",cookie.Value,&cookieValue)
		if err == nil {
			userName = cookieValue["username"]
		}
	}
	return userName
}


func ClearSession(response http.ResponseWriter){
	cookie := &http.Cookie{
		Name: "session",
		Value: "",
		Path: "/",
		MaxAge: -1,
	}
	http.SetCookie(response,cookie)
}

/////cookies

//handlers

// "/login"
var LoginPageHandler = http.HandlerFunc(
	func(w http.ResponseWriter,r *http.Request){
		if r.Method == "GET" {
			parsedTemplate,_ := template.ParseFiles("templates/loginPage.html")
			parsedTemplate.Execute(w,nil)
		} else {
			username := r.FormValue("username")
			password := r.FormValue("password")
			target   := "/login"
			if username != "" && password != ""{
				SetsSession(username,w)
				target = "/books"
			}
			http.Redirect(w,r,target,302)
		}


	})
// "/books"
var BooksPageHandler = http.HandlerFunc(
	func(w http.ResponseWriter,r *http.Request){
		Username := GetUserName(r)
		if Username != ""{
			books := GetAllBooksFromDB()
			parsedTemplate,_ := template.ParseFiles("templates/books.html")
			parsedTemplate.Execute(w,books)
		} else {
			http.Redirect(w,r,"/login",302)
		}

	})


var CreateBookHandler = http.HandlerFunc(
	func(w http.ResponseWriter,r *http.Request) {

		Username := GetUserName(r)
		if Username == "" {
			http.Redirect(w, r, "/login", 302)
			return
		}

		if r.Method == "GET" {
			parsedTemplate, _ := template.ParseFiles("templates/createBook.html")
			parsedTemplate.Execute(w, nil)
		} else {
			title := r.FormValue("title")
			rating := r.FormValue("rating")
			author := r.FormValue("author")
			detail := r.FormValue("detail")

			_, err := strconv.Atoi(rating)
			if err != nil {
				rating = "0"
			}

			stmt, err :=db.Prepare("INSERT INTO books (Title,Author,Rating) VALUES ('" +
				title+"','"+author+"',"+rating+");")
			if err !=nil {
				log.Println("insert error",err)
			}

			res, err := stmt.Exec()
			if err !=nil {
				log.Println("stmt.Exec insert error",err)
			}
			lid, err := res.LastInsertId()

			newPage,_ := os.Create("content/books/"+fmt.Sprintf("%d.html",lid))
			data,_ := ioutil.ReadFile("content/detailbook.html")
			data = bytes.Replace(data,[]byte("DETAIL"),[]byte(detail),1)
			newPage.Write(data)

			http.Redirect(w, r, "/books/create", 302)

		}
	})



func (a  BookCollection) Len() int           { return len(a.Books) }
func (a  BookCollection) Swap(i, j int)      { a.Books[i].Rating, a.Books[j].Rating = a.Books[j].Rating, a.Books[i].Rating }
func (a  BookCollection) Less(i, j int) bool { return a.Books[i].Rating > a.Books[j].Rating }

var ReverseBookHandler = http.HandlerFunc(
func(w http.ResponseWriter,r *http.Request){
	books := GetAllBooksFromDB()
	sort.Sort(books)
	parsedTemplate,_ := template.ParseFiles("templates/books.html")
	parsedTemplate.Execute(w,books)
})

func (a  JournalCollection) Len() int           { return len(a.Journals) }
func (a  JournalCollection) Swap(i, j int)      { a.Journals[i].PageAmount, a.Journals[j].PageAmount = a.Journals[j].PageAmount, a.Journals[i].PageAmount }
func (a  JournalCollection) Less(i, j int) bool { return a.Journals[i].PageAmount > a.Journals[j].PageAmount }


var ReverseJournalPageHandler= http.HandlerFunc(
	func(w http.ResponseWriter,r *http.Request){
		journals := GetAllJournalFromDB()
		sort.Sort(journals)
		parsedTemplate,_ := template.ParseFiles("templates/journal.html")
		parsedTemplate.Execute(w,journals)
	})

var JournalPageHandler = http.HandlerFunc(
	func(w http.ResponseWriter,r *http.Request){
		Username := GetUserName(r)
		if Username != ""{
			parsedTemplate,_ := template.ParseFiles("templates/journal.html")
			Journal := GetAllJournalFromDB()
			parsedTemplate.Execute(w,Journal)
		} else {
			http.Redirect(w,r,"/login",302)
		}
	})

var CreateJournalHandler = http.HandlerFunc(
	func(w http.ResponseWriter,r *http.Request){
		Username := GetUserName(r)
		if Username == ""{
			http.Redirect(w,r,"/login",302)
			return
		}

		if r.Method == "GET" {
			parsedTemplate,_ := template.ParseFiles("templates/createJournal.html")
			Journal := GetAllJournalFromDB()
			parsedTemplate.Execute(w,Journal)
		} else {
			title := r.FormValue("title")
			editor := r.FormValue("editor")
			pageamount := r.FormValue("pageamount")
			detail := r.FormValue("detail")

			_, err := strconv.Atoi(pageamount)
			if err != nil {
				pageamount = "0"
			}

			stmt, err :=db.Prepare("INSERT INTO journals (Title,Editor,PageAmount) VALUES('" +
				title+"','"+editor+"',"+pageamount+");")
			if err !=nil {
				log.Println("insert error",err)
			}

			res, err := stmt.Exec()
			if err !=nil {
				log.Println("stmt.Exec insert error",err)
			}
			lid, err := res.LastInsertId()

			newPage,_ := os.Create("content/journals/"+fmt.Sprintf("%d.html",lid))
			data,_ := ioutil.ReadFile("content/detailjournal.html")
			data = bytes.Replace(data,[]byte("DETAIL"),[]byte(detail),1)
			newPage.Write(data)

			http.Redirect(w,r,"/journals/create",302)
		}
	})

var LogoutFormPageHandler = func(w http.ResponseWriter,r *http.Request){
	ClearSession(w)
	http.Redirect(w,r,"/login",302)
}

var IndexPageHandler = http.HandlerFunc(
	func(w http.ResponseWriter,r *http.Request){
	Username := GetUserName(r)
	if Username == ""{
		http.Redirect(w,r,"/login",302)
		return
	}
	parsedTemplate,_ := template.ParseFiles("templates/index.html")
	parsedTemplate.Execute(w,nil)

})


//handlers




func init(){
	db, connectionError = sql.Open(driverName, dataSourceName)
	if connectionError != nil {
		log.Fatal("error while connectiong to database:", connectionError)
	}
}


var DetailPageHandler =  http.HandlerFunc(
	func (w http.ResponseWriter,r *http.Request){
	vars := mux.Vars(r)
	id := vars["id"]
	//TODO isn't Correct id

		parsedTemplate,_ := template.ParseFiles("content/books/"+id+".html")
		parsedTemplate.Execute(w,nil)

})

var DetailJournalPageHandler =  http.HandlerFunc(
	func (w http.ResponseWriter,r *http.Request){
		vars := mux.Vars(r)
		id := vars["id"]
		//TODO isn't Correct id

		parsedTemplate,_ := template.ParseFiles("content/journal/"+id+".html")
		parsedTemplate.Execute(w,nil)

	})

func main(){

	defer	db.Close()

	router := mux.NewRouter()

	logFile, err := os.OpenFile("server.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	router.Handle("/",handlers.LoggingHandler(logFile,
		IndexPageHandler)).Methods("GET")

	router.Handle("/login",handlers.LoggingHandler(logFile,
		LoginPageHandler)).Methods("GET","POST")

	router.Handle("/books",handlers.LoggingHandler(logFile,
		BooksPageHandler)).Methods("GET")

	router.Handle("/books/book/{id}",handlers.LoggingHandler(logFile,
		DetailPageHandler)).Methods("GET")

	router.Handle("/books/create",handlers.LoggingHandler(logFile,
		CreateBookHandler)).Methods("GET","POST")

	router.Handle("/books/books_reversed",handlers.LoggingHandler(logFile,
		ReverseBookHandler)).Methods("GET","POST")



	router.Handle("/journals",handlers.LoggingHandler(logFile,
		JournalPageHandler)).Methods("GET")

	router.Handle("/journals/journal/{id}",handlers.LoggingHandler(logFile,
		DetailJournalPageHandler)).Methods("GET")

	router.Handle("/journals_reversed",handlers.LoggingHandler(logFile,
		ReverseJournalPageHandler)).Methods("GET")

	router.Handle("/journals/create",handlers.LoggingHandler(logFile,
		CreateJournalHandler)).Methods("GET","POST")



	router.Handle("/logout",handlers.LoggingHandler(logFile,
		http.HandlerFunc(LogoutFormPageHandler))).Methods("POST")
	fmt.Println("listening:"+connPort)
	err = http.ListenAndServe(connHost+":"+connPort,router)
	if err != nil {
		log.Fatal("error starting server: ",err)
		return
	}

}
