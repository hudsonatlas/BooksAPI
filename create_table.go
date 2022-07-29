package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func create() {
	db, err := sql.Open("mysql", "root:admin@tcp(localhost:3306)/apiBook")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	db.Exec("CREATE TABLE IF NOT EXISTS livros (id INT NOT NULL AUTO_INCREMENT, titulo VARCHAR(255) NOT NULL, autor VARCHAR(255) NOT NULL, PRIMARY KEY (id))")
	db.Exec("INSERT INTO livros (titulo, autor) VALUES ('O Senhor dos Anéis', 'J.R.R. Tolkien')")
	db.Exec("INSERT INTO livros (titulo, autor) VALUES ('Harry Potter e a Ordem da Fênix', 'J.K. Rowling')")
	db.Exec("INSERT INTO livros (titulo, autor) VALUES ('O Hobbit', 'J.R.R. Tolkien')")
	db.Exec("INSERT INTO livros (titulo, autor) VALUES ('O Senhor dos Anéis', 'J.R.R. Tolkien')")

}
