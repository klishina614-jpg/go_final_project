package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/anaklisina/go_final_project/pkg/api"
	"github.com/anaklisina/go_final_project/pkg/db"
)

func main() {
	port := "7540"
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		port = envPort
	}

	dbFile := "scheduler.db"
	if envDBFile := os.Getenv("TODO_DBFILE"); envDBFile != "" {
		dbFile = envDBFile
	}

	err := db.Init(dbFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.DB.Close()

	api.Init()

	http.Handle("/", http.FileServer(http.Dir("./web")))

	fmt.Println("Сервер запущен на порту", port)

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("ошибка запуска сервера:", err)
	}
}
