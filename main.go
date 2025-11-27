package main

import (
    "github.com/nydeg/bd/internal/database"
    "github.com/nydeg/bd/internal/gui"
    "log"
)

func main() {
    db, err := database.OpenDatabase("data/books.db")
    if err != nil {
        log.Fatalf("Ошибка инициализации БД: %v", err)
    }
    defer db.Close()

    gui.Run(db)
}