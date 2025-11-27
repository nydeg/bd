package database

import (
    "bufio"
    "fmt"
    "os"
    "sort"
    "strconv"
    "strings"

    "github.com/xuri/excelize/v2"
    // "unicode/utf8"
)


type Database struct {
    file       *os.File
    filePath   string
    recordSize int64
    
    idIndex     map[int32]int64
    titleIndex  map[string][]int64
    authorIndex map[string][]int64
    yearIndex   map[int32][]int64
    
    freeList []int64
}

// O(n) просто считываем весь файл, зная расположение нужных нам полей
func OpenDatabase(filePath string) (*Database, error) {
    os.MkdirAll("data", 0755)
    
    file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
    if err != nil {
        return nil, fmt.Errorf("ошибка открытия файла: %v", err)
    }
    
    db := &Database{
        file:       file,
        filePath:   filePath,
        recordSize: 152, // байтики сложил просто
        idIndex:    make(map[int32]int64),
        titleIndex: make(map[string][]int64),
        authorIndex: make(map[string][]int64),
        yearIndex:  make(map[int32][]int64),
        freeList:   []int64{},
    }
    
    if err := db.rebuildIndexes(); err != nil {
        file.Close()
        return nil, fmt.Errorf("ошибка восстановления индексов: %v", err)
    }
    
    return db, nil
}

// O(1)
func (db *Database) Close() error {
    if db.file != nil {
        return db.file.Close()
    }
    return nil
}

// O(1), константы небольшие
func (db *Database) ClearDatabase() error {
    if err := db.file.Truncate(0); err != nil {
        return fmt.Errorf("ошибка очистки файла: %v", err)
    }
    
    if _, err := db.file.Seek(0, 0); err != nil {
        return fmt.Errorf("ошибка перемещения в начало файла: %v", err)
    }
    
    db.idIndex = make(map[int32]int64)
    db.titleIndex = make(map[string][]int64)
    db.authorIndex = make(map[string][]int64)
    db.yearIndex = make(map[int32][]int64)
    db.freeList = []int64{}
    
    return nil
}

// O(nlogn)
func (db *Database) GetAllBooks() ([]BookView, error) {
    var views []BookView

    // проходимся по n позициям, но достаем за O(1) (суммарно O(n))
    for _, position := range db.idIndex {
        book, err := db.readRecord(position)
        if err != nil {
            continue
        }
        views = append(views, book.ToView())
    }
    
    // сортирую id => O(nlogn)
    sort.Slice(views, func(i, j int) bool {
        return views[i].ID < views[j].ID
    })
    
    return views, nil
}

// O(1)
func (db *Database) AddBook(bookView BookView) error {
    book := bookView.ToBook()
    
    // благодаря мапам все быренько (проверяем на существование по айди в мапе, потом запись в мапу и обновление индексов)
    if _, exists := db.idIndex[book.ID]; exists {
        return fmt.Errorf("книга с ID %d уже существует", book.ID)
    }
    
    var position int64
    if len(db.freeList) > 0 {
        position = db.freeList[len(db.freeList)-1]
        db.freeList = db.freeList[:len(db.freeList)-1]
    } else {
        stat, err := db.file.Stat()
        if err != nil {
            return err
        }
        position = stat.Size()
    }
    
    if err := db.writeRecord(book, position); err != nil {
        return err
    }
    
    db.updateIndexes(book, position)
    return nil
}


func (db *Database) UpdateBook(bookView BookView) error {
    position, exists := db.idIndex[bookView.ID]
    if !exists {
        return fmt.Errorf("книга с ID %d не найдена", bookView.ID)
    }
    
    oldBook, err := db.readRecord(position)
    if err != nil {
        return err
    }
    
    db.removeFromIndexes(oldBook, position)
    
    newBook := bookView.ToBook()
    if err := db.writeRecord(newBook, position); err != nil {
        return err
    }
    
    db.updateIndexes(newBook, position)
    return nil
}

func (db *Database) DeleteBook(id int32) error {
    position, exists := db.idIndex[id]
    if !exists {
        return fmt.Errorf("книга с ID %d не найдена", id)
    }
    
    book, err := db.readRecord(position)
    if err != nil {
        return err
    }
    
    db.removeFromIndexes(book, position)
    db.freeList = append(db.freeList, position)
    return nil
}

func (db *Database) FindByID(id int32) (*Book, error) {
    position, exists := db.idIndex[id]
    if !exists {
        return nil, fmt.Errorf("книга с ID %d не найдена", id)
    }
    return db.readRecord(position)
}

func (db *Database) FindBooks(field, value string) ([]BookView, error) {
    allBooks, err := db.GetAllBooks()
    if err != nil {
        return nil, err
    }
    
    var result []BookView
    
    searchValue := strings.ToLower(strings.TrimSpace(value))
    
    for _, book := range allBooks {
        var match bool
        
        switch field {
        case "ID":
            searchID, err := strconv.Atoi(value)
            if err == nil && book.ID == int32(searchID) {
                match = true
            }
            
        case "Название":
            bookTitle := strings.ToLower(book.Title)
            if strings.Contains(bookTitle, searchValue) {
                match = true
            }
            
        case "Автор":
            bookAuthor := strings.ToLower(book.Author)
            if strings.Contains(bookAuthor, searchValue) {
                match = true
            }
            
        case "Год издания":
            searchYear, err := strconv.Atoi(value)
            if err == nil && book.Year == int32(searchYear) {
                match = true
            }
            
        case "Тираж":
            searchCopies, err := strconv.Atoi(value)
            if err == nil && book.Copies == int32(searchCopies) {
                match = true
            }
        }
        
        if match {
            result = append(result, book)
        }
    }
    
    if len(result) == 0 {
        return nil, fmt.Errorf("книги по запросу '%s' = '%s' не найдены", field, value)
    }
    
    return result, nil
}

func (db *Database) ExportToTxt(filename string) error {
    books, err := db.GetAllBooks()
    if err != nil {
        return fmt.Errorf("ошибка получения книг: %v", err)
    }

    file, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("ошибка создания файла: %v", err)
    }
    defer file.Close()

    writer := bufio.NewWriter(file)
    
    header := "ID|Название|Автор|Год|Тираж\n"
    if _, err := writer.WriteString(header); err != nil {
        return fmt.Errorf("ошибка записи заголовка: %v", err)
    }

    for _, book := range books {
        line := fmt.Sprintf("%d|%s|%s|%d|%d\n", 
            book.ID, book.Title, book.Author, book.Year, book.Copies)
        if _, err := writer.WriteString(line); err != nil {
            return fmt.Errorf("ошибка записи данных: %v", err)
        }
    }

    if err := writer.Flush(); err != nil {
        return fmt.Errorf("ошибка сохранения файла: %v", err)
    }

    return nil
}

func (db *Database) ImportFromTxt(filename string) (int, error) {
    file, err := os.Open(filename)
    if err != nil {
        return 0, fmt.Errorf("ошибка открытия файла: %v", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    importedCount := 0
    lineNumber := 0

    var booksToImport []BookView

    for scanner.Scan() {
        lineNumber++
        line := strings.TrimSpace(scanner.Text())
        
        if line == "" || line == "ID|Название|Автор|Год|Тираж" {
            continue
        }

        parts := strings.Split(line, "|")
        if len(parts) != 5 {
            return importedCount, fmt.Errorf("ошибка в строке %d: неверный формат данных", lineNumber)
        }

        id, err := strconv.Atoi(parts[0])
        if err != nil {
            return importedCount, fmt.Errorf("ошибка в строке %d: неверный ID", lineNumber)
        }

        year, err := strconv.Atoi(parts[3])
        if err != nil {
            return importedCount, fmt.Errorf("ошибка в строке %d: неверный год", lineNumber)
        }

        copies, err := strconv.Atoi(parts[4])
        if err != nil {
            return importedCount, fmt.Errorf("ошибка в строке %d: неверный тираж", lineNumber)
        }

        book := BookView{
            ID:     int32(id),
            Title:  parts[1],
            Author: parts[2],
            Year:   int32(year),
            Copies: int32(copies),
        }

        booksToImport = append(booksToImport, book)
    }

    if err := scanner.Err(); err != nil {
        return importedCount, fmt.Errorf("ошибка чтения файла: %v", err)
    }

    sort.Slice(booksToImport, func(i, j int) bool {
        return booksToImport[i].ID < booksToImport[j].ID
    })

    for _, book := range booksToImport {
        if err := db.AddBook(book); err != nil {
            if strings.Contains(err.Error(), "уже существует") {
                if err := db.UpdateBook(book); err != nil {
                    return importedCount, fmt.Errorf("ошибка обновления книги в строке %d: %v", lineNumber, err)
                }
            } else {
                return importedCount, fmt.Errorf("ошибка добавления книги в строке %d: %v", lineNumber, err)
            }
        }
        
        importedCount++
    }

    return importedCount, nil
}

func (db *Database) GetStats() (int, int64, error) {
    books, err := db.GetAllBooks()
    if err != nil {
        return 0, 0, err
    }

    stat, err := db.file.Stat()
    if err != nil {
        return 0, 0, err
    }

    return len(books), stat.Size(), nil
}

func BytesToString(data []byte) string {
    return bytesToString(data)
}

func (db *Database) readRecord(position int64) (*Book, error) {
    if _, err := db.file.Seek(position, 0); err != nil {
        return nil, err
    }
    
    buffer := make([]byte, db.recordSize)
    n, err := db.file.Read(buffer)
    if err != nil {
        return nil, err
    }
    if n != int(db.recordSize) {
        return nil, fmt.Errorf("неполная запись")
    }
    
    return bytesToBook(buffer), nil
}

func (db *Database) writeRecord(book *Book, position int64) error {
    if _, err := db.file.Seek(position, 0); err != nil {
        return err
    }
    
    data := bookToBytes(book)
    _, err := db.file.Write(data)
    return err
}

func (db *Database) rebuildIndexes() error {
    stat, err := db.file.Stat()
    if err != nil {
        return err
    }
    
    fileSize := stat.Size()
    var position int64 = 0
    
    for position < fileSize {
        book, err := db.readRecord(position)
        if err != nil {
            db.freeList = append(db.freeList, position)
        } else {
            db.idIndex[book.ID] = position
            
            title := bytesToString(book.Title[:])
            db.titleIndex[title] = append(db.titleIndex[title], position)
            
            author := bytesToString(book.Author[:])
            db.authorIndex[author] = append(db.authorIndex[author], position)
            
            db.yearIndex[book.Year] = append(db.yearIndex[book.Year], position)
        }
        
        position += db.recordSize
    }
    
    return nil
}

func (db *Database) updateIndexes(book *Book, position int64) {
    db.idIndex[book.ID] = position
    
    title := bytesToString(book.Title[:])
    db.titleIndex[title] = append(db.titleIndex[title], position)
    
    author := bytesToString(book.Author[:])
    db.authorIndex[author] = append(db.authorIndex[author], position)
    
    db.yearIndex[book.Year] = append(db.yearIndex[book.Year], position)
}

func (db *Database) removeFromIndexes(book *Book, position int64) {
    delete(db.idIndex, book.ID)
    
    title := bytesToString(book.Title[:])
    db.removeFromSliceIndex(db.titleIndex, title, position)
    
    author := bytesToString(book.Author[:])
    db.removeFromSliceIndex(db.authorIndex, author, position)
    
    db.removeFromSliceIndexInt(db.yearIndex, book.Year, position)
}

func (db *Database) removeFromSliceIndex(index map[string][]int64, key string, position int64) {
    positions := index[key]
    for i, pos := range positions {
        if pos == position {
            index[key] = append(positions[:i], positions[i+1:]...)
            break
        }
    }
    if len(index[key]) == 0 {
        delete(index, key)
    }
}

func (db *Database) removeFromSliceIndexInt(index map[int32][]int64, key int32, position int64) {
    positions := index[key]
    for i, pos := range positions {
        if pos == position {
            index[key] = append(positions[:i], positions[i+1:]...)
            break
        }
    }
    if len(index[key]) == 0 {
        delete(index, key)
    }
}

func (db *Database) ExportToExcel(filename string) error {
    books, err := db.GetAllBooks()
    if err != nil {
        return fmt.Errorf("ошибка получения книг: %v", err)
    }

    f := excelize.NewFile()
    defer f.Close()

    // Создаем новый лист
    index, err := f.NewSheet("Книги")
    if err != nil {
        return fmt.Errorf("ошибка создания листа: %v", err)
    }

    // Устанавливаем заголовки
    headers := []string{"ID", "Название", "Автор", "Год издания", "Тираж"}
    for i, h := range headers {
        cell, _ := excelize.CoordinatesToCellName(i+1, 1)
        f.SetCellValue("Книги", cell, h)
        
        // Устанавливаем стиль для заголовков
        style, _ := f.NewStyle(&excelize.Style{
            Font: &excelize.Font{Bold: true},
            Fill: excelize.Fill{Type: "pattern", Color: []string{"#DDEBF7"}, Pattern: 1},
        })
        f.SetCellStyle("Книги", cell, cell, style)
    }

    // Заполняем данные
    for i, book := range books {
        row := i + 2
        f.SetCellValue("Книги", "A"+strconv.Itoa(row), book.ID)
        f.SetCellValue("Книги", "B"+strconv.Itoa(row), book.Title)
        f.SetCellValue("Книги", "C"+strconv.Itoa(row), book.Author)
        f.SetCellValue("Книги", "D"+strconv.Itoa(row), book.Year)
        f.SetCellValue("Книги", "E"+strconv.Itoa(row), book.Copies)
    }

    // Автоматическая ширина колонок
    f.SetColWidth("Книги", "A", "A", 10)
    f.SetColWidth("Книги", "B", "B", 40)
    f.SetColWidth("Книги", "C", "C", 25)
    f.SetColWidth("Книги", "D", "D", 12)
    f.SetColWidth("Книги", "E", "E", 12)

    // Устанавливаем активный лист
    f.SetActiveSheet(index)

    // Удаляем дефолтный лист
    f.DeleteSheet("Sheet1")

    // Сохраняем файл
    if err := f.SaveAs(filename); err != nil {
        return fmt.Errorf("ошибка сохранения файла: %v", err)
    }

    return nil
}

// ImportFromExcel импортирует книги из Excel файла
func (db *Database) ImportFromExcel(filename string) (int, error) {
    f, err := excelize.OpenFile(filename)
    if err != nil {
        return 0, fmt.Errorf("ошибка открытия файла: %v", err)
    }
    defer f.Close()

    // Получаем строки из листа "Книги"
    rows, err := f.GetRows("Книги")
    if err != nil {
        return 0, fmt.Errorf("ошибка чтения листа: %v", err)
    }

    if len(rows) < 2 {
        return 0, fmt.Errorf("файл не содержит данных")
    }

    importedCount := 0

    // Проходим по строкам, начиная со второй (первая - заголовки)
    for i, row := range rows[1:] {
        // Пропускаем пустые строки
        if len(row) < 5 {
            continue
        }

        // Парсим ID
        id, err := strconv.Atoi(row[0])
        if err != nil {
            return importedCount, fmt.Errorf("ошибка в строке %d: неверный ID", i+2)
        }

        // Парсим год
        year, err := strconv.Atoi(row[3])
        if err != nil {
            return importedCount, fmt.Errorf("ошибка в строке %d: неверный год", i+2)
        }

        // Парсим тираж
        copies, err := strconv.Atoi(row[4])
        if err != nil {
            return importedCount, fmt.Errorf("ошибка в строке %d: неверный тираж", i+2)
        }

        book := BookView{
            ID:     int32(id),
            Title:  row[1],
            Author: row[2],
            Year:   int32(year),
            Copies: int32(copies),
        }

        // Пытаемся добавить книгу
        if err := db.AddBook(book); err != nil {
            // Если книга с таким ID уже существует, обновляем её
            if err := db.UpdateBook(book); err != nil {
                return importedCount, fmt.Errorf("ошибка обновления книги в строке %d: %v", i+2, err)
            }
        }

        importedCount++
    }

    return importedCount, nil
}