package database

import (
    "bytes"
    "encoding/binary"
    "unicode/utf8"
)

func bookToBytes(book *Book) []byte {
    buf := make([]byte, 152)
    
    binary.LittleEndian.PutUint32(buf[0:4], uint32(book.ID))
    copy(buf[4:104], book.Title[:])
    copy(buf[104:144], book.Author[:])
    binary.LittleEndian.PutUint32(buf[144:148], uint32(book.Year))
    binary.LittleEndian.PutUint32(buf[148:152], uint32(book.Copies))
    
    return buf
}

func bytesToBook(data []byte) *Book {
    if len(data) < 152 {
        panic("недостаточно данных для преобразования в Book")
    }
    
    book := &Book{}
    book.ID = int32(binary.LittleEndian.Uint32(data[0:4]))
    copy(book.Title[:], data[4:104])
    copy(book.Author[:], data[104:144])
    book.Year = int32(binary.LittleEndian.Uint32(data[144:148]))
    book.Copies = int32(binary.LittleEndian.Uint32(data[148:152]))
    
    return book
}

func copyStringToBytes(s string, target []byte) {
    bytes := []byte(s)
    if len(bytes) > len(target) {
        bytes = bytes[:len(target)]
    }
    copy(target, bytes)
}

func bytesToString(data []byte) string {
    end := 0
    for i, b := range data {
        if b == 0 {
            end = i
            break
        }
        if i == len(data)-1 {
            end = len(data)
        }
    }
    
    str := string(data[:end])
    if !utf8.ValidString(str) {
        var result bytes.Buffer
        for len(str) > 0 {
            r, size := utf8.DecodeRuneInString(str)
            if r == utf8.RuneError {
                result.WriteRune('?')
                str = str[1:]
            } else {
                result.WriteRune(r)
                str = str[size:]
            }
        }
        return result.String()
    }
    return str
}