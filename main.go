package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type Livro struct {
	Id     int    `json:"id"`
	Titulo string `json:"titulo"`
	Autor  string `json:"autor"`
}

var Livros []Livro = []Livro{
	Livro{Id: 1, Titulo: "Harry Potter e a Ordem da Fênix", Autor: "J.K. Rowling"},
	Livro{Id: 2, Titulo: "O Senhor dos Anéis", Autor: "J.R.R. Tolkien"},
	Livro{Id: 3, Titulo: "O Hobbit", Autor: "J.R.R. Tolkien"},
	Livro{Id: 4, Titulo: "O Senhor dos Anéis", Autor: "J.R.R. Tolkien"},
}

func main() {
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
	fmt.Println(r.URL.Path)

	partes := strings.Split(r.URL.Path, "/")

	id, _ := strconv.Atoi(partes[2])

	for _, livro := range Livros {
		if livro.Id == id {
			json.NewEncoder(w).Encode(livro)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)

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
	encoder := json.NewEncoder(w)
	encoder.Encode(Livros)
}

func createBookHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)

	body, erro := ioutil.ReadAll(r.Body)
	if erro != nil {
		fmt.Println(erro)
		return
	}

	var novolivro Livro
	json.Unmarshal(body, &novolivro)
	novolivro.Id = len(Livros) + 1
	Livros = append(Livros, novolivro)

	encoder := json.NewEncoder(w)
	encoder.Encode(novolivro)
}

func updateBookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	partes := strings.Split(r.URL.Path, "/")
	id, err := strconv.Atoi(partes[2])

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var modBook Livro
	erroJson := json.Unmarshal(body, &modBook)

	if erroJson != nil {
		fmt.Println(erroJson)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key := -1
	for i, livro := range Livros {
		if livro.Id == id {
			key = i
			break
		}
	}

	if key < 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	Livros[key] = modBook

	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(modBook)
}

func deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	partes := strings.Split(r.URL.Path, "/")

	id, erro := strconv.Atoi(partes[2])

	if erro != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key := -1
	for i, livro := range Livros {
		if livro.Id == id {
			key = i
			break
		}
	}

	if key < 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	leftSide := Livros[0:key]
	rightSide := Livros[key+1 : len(Livros)]

	Livros = append(leftSide, rightSide...)
	w.WriteHeader(http.StatusNoContent)
}
