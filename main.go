package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Livro struct {
	Id     int    `json:"id"`
	Titulo string `json:"titulo"`
	Autor  string `json:"autor"`
}

type ErrorReturn struct {
	Error string `json:"error"`
}

type configdb struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

var db *sql.DB

func configDB() {
	var erroAbertura error
	dat, err := os.ReadFile("config/env.json")

	if err != nil {
		log.Fatal(err)
	}

	var config map[string]string
	err = json.Unmarshal(dat, &config)

	if err != nil {
		log.Fatal(err)
	}

	configdb := configdb{
		Host:     config["DB_HOST_COM_PORTA"],
		User:     config["DB_USUARIO"],
		Password: config["DB_SENHA"],
		Database: config["DB_BANCO_DE_DADOS"],
	}

	db, erroAbertura = sql.Open("mysql", configdb.User+":"+configdb.Password+"@tcp("+configdb.Host+")/"+configdb.Database)

	if erroAbertura != nil {
		log.Fatal(erroAbertura.Error())
	}

	erroPing := db.Ping()

	if erroPing != nil {
		log.Fatal(erroPing.Error())
	}
}

func main() {
	configDB()
	configServer()
}

func configServer() {
	route := mux.NewRouter().StrictSlash(true)
	route.Use(jsonContentType)
	configRouter(route)

	fmt.Println("Listening on port 8081")
	log.Fatal(http.ListenAndServe(":8081", route))
}

func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func configRouter(route *mux.Router) {

	route.HandleFunc("/", RouteMain).Methods("GET")
	route.HandleFunc("/livros", booksHandler).Methods("GET")
	route.HandleFunc("/livros", createBookHandler).Methods("POST")
	route.HandleFunc("/livros/{id}", RouteBook).Methods("GET", "PUT", "DELETE")
}

func GetBookHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, _ := strconv.Atoi(vars["id"])

	result := db.QueryRow("SELECT l.id, l.autor, l.titulo FROM livros l WHERE id = ?", id)
	var livro Livro

	err := result.Scan(&livro.Id, &livro.Autor, &livro.Titulo)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(livro)

}

func RouteMain(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func RouteBook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		GetBookHandler(w, r)
	case "PUT":
		updateBookHandler(w, r)
	case "DELETE":
		deleteBookHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	reg, err := db.Query("SELECT id, autor, titulo FROM livros")

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var livros []Livro = make([]Livro, 0)
	for reg.Next() {
		var livro Livro
		err := reg.Scan(&livro.Id, &livro.Autor, &livro.Titulo)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		livros = append(livros, livro)
	}

	defer reg.Close()

	encoder := json.NewEncoder(w)
	encoder.Encode(livros)
}

func validacaoLivro(novolivro Livro) []string {

	erros := make([]string, 0)

	if len(novolivro.Titulo) == 0 {
		erros = append(erros, "Titulo não pode ser vazio")
	}

	if len(novolivro.Autor) == 0 {
		erros = append(erros, "Autor não pode ser vazio")
	}

	if len(novolivro.Autor) > 100 {
		erros = append(erros, "Autor não pode ter mais de 100 caracteres")
	}

	if len(novolivro.Titulo) > 100 {
		erros = append(erros, "Titulo não pode ter mais de 100 caracteres")
	}

	return erros
}

func createBookHandler(w http.ResponseWriter, r *http.Request) {
	body, erro := ioutil.ReadAll(r.Body)
	if erro != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var novolivro Livro
	erro = json.Unmarshal(body, &novolivro)

	if erro != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	erroValidacao := validacaoLivro(novolivro)

	if len(erroValidacao) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		encoder := json.NewEncoder(w)
		encoder.Encode(ErrorReturn{Error: strings.Join(erroValidacao, "; ")})
		return
	}

	result, errorInsert := db.Exec("INSERT INTO livros (autor, titulo) VALUES (?, ?)", novolivro.Autor, novolivro.Titulo)

	idNovoLivro, errorLastInsertId := result.LastInsertId()

	if errorInsert != nil || errorLastInsertId != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorReturn{Error: "Erro ao criar livro"})
		return
	}

	novolivro.Id = int(idNovoLivro)

	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(novolivro)
}

func updateBookHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, erro := strconv.Atoi(vars["id"])

	if erro != nil {
		fmt.Println(erro)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var livroMod Livro
	erroJson := json.Unmarshal(body, &livroMod)

	if erroJson != nil {
		fmt.Println(erroJson)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reg := db.QueryRow("SELECT l.id, l.autor, l.titulo FROM livros l WHERE id = ?", id)

	var livro Livro

	errScan := reg.Scan(&livro.Id, &livro.Autor, &livro.Titulo)

	if errScan != nil {
		fmt.Println(errScan)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, errExec := db.Exec("UPDATE livros SET autor = ?, titulo = ? WHERE id = ?", livroMod.Autor, livroMod.Titulo, id)

	if errExec != nil {
		fmt.Println(errExec)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(livroMod)
}

func deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, erro := strconv.Atoi(vars["id"])

	if erro != nil {
		fmt.Println(erro)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reg := db.QueryRow("SELECT l.id FROM livros l WHERE id = ?", id)

	var livro_id int

	errScan := reg.Scan(&livro_id)

	if errScan != nil {
		fmt.Println(errScan)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, errExec := db.Exec("DELETE FROM livros WHERE id = ?", livro_id)

	if errExec != nil {
		fmt.Println(errExec)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
