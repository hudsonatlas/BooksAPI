package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	configRouter()

	fmt.Println("Listening on port 1337")
	log.Fatal(http.ListenAndServe(":1337", nil))
}

func configRouter() {
	http.HandleFunc("/", RouteMain)
	http.HandleFunc("/livros", RouteBook)
	http.HandleFunc("/livros/", RouteBook)
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
	http.HandleFunc("/livros", RouteBook)
}

func RouteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	partes := strings.Split(r.URL.Path, "/")

	if len(partes) == 2 || len(partes) == 3 && partes[2] == "" {
		if r.Method == "GET" {
			booksHandler(w, r)
		} else if r.Method == "POST" {
			createBookHandler(w, r)
		}
	} else if len(partes) == 3 || len(partes) == 4 && partes[3] == "" {
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
	} else {
		w.WriteHeader(http.StatusNotFound)
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
