package database

import (
	// "encoding/binary"
)

type Book struct {
    ID     int32      // 4 байта
    Title  [100]byte  // 100 байт
    Author [40]byte   // 40 байт  
    Year   int32      // 4 байта
    Copies int32      // 4 байта
}

type BookView struct {
    ID     int32  `json:"id"`
    Title  string `json:"title"`
    Author string `json:"author"`
    Year   int32  `json:"year"`
    Copies int32  `json:"copies"`
}


func (b *Book) ToView() BookView {
    return BookView{
        ID:     b.ID,
        Title:  bytesToString(b.Title[:]),
        Author: bytesToString(b.Author[:]),
        Year:   b.Year,
        Copies: b.Copies,
    }
}

func (v *BookView) ToBook() *Book {
    book := &Book{
        ID:     v.ID,
        Year:   v.Year,
        Copies: v.Copies,
    }
    copyStringToBytes(v.Title, book.Title[:])
    copyStringToBytes(v.Author, book.Author[:])
    return book
}