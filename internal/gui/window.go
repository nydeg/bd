package gui

import (
    "github.com/nydeg/bd/internal/database"
    "fmt"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

type App struct {
    database      *database.Database
    window        fyne.Window
    table         *widget.Table
    books         []database.BookView
    // statusLabel   *widget.Label
    updateStatusBar func(string)
}

func Run(db *database.Database) {
    myApp := app.New()
    window := myApp.NewWindow("Book Database - База данных книг")
    
    app := &App{
        database: db,
        window:   window,
        books:    []database.BookView{},
    }
    
    app.createUI()
    window.Resize(fyne.NewSize(1000, 700))
    window.ShowAndRun()
}

func (a *App) createUI() {
    a.table = a.createTable()
    toolbar := a.createToolbar()
    statusBar := a.createStatusBar()
    
    content := container.NewBorder(toolbar, statusBar, nil, nil, a.table)
    a.window.SetContent(content)
    
    a.refreshTable()
}

func (a *App) createToolbar() fyne.CanvasObject {
    addButton := widget.NewButton("➕ Добавить", a.showAddDialog)
    editButton := widget.NewButton("✏️ Редактировать", a.showEditDialog)
    deleteButton := widget.NewButton("🗑️ Удалить", a.showDeleteDialog)
    searchButton := widget.NewButton("🔍 Поиск", a.showSearchDialog)
    refreshButton := widget.NewButton("🔄 Обновить", a.refreshTable)

    importTxtButton := widget.NewButton("📥 Импорт TXT", a.showImportDialog)
    exportTxtButton := widget.NewButton("📤 Экспорт TXT", a.showExportDialog)
    
    importExcelButton := widget.NewButton("📊 Импорт Excel", a.showImportExcelDialog)
    exportExcelButton := widget.NewButton("📈 Экспорт Excel", a.showExportExcelDialog)
    
    clearButton := widget.NewButton("🗑️ Очистить БД", a.showClearDatabaseDialog)
    statsButton := widget.NewButton("📊 Статистика", a.showStatsDialog)
    
    toolbar := container.NewHBox(
        addButton, editButton, deleteButton, searchButton, 
        widget.NewSeparator(),
        refreshButton, 
        widget.NewSeparator(),
        importTxtButton, exportTxtButton,
        widget.NewSeparator(),
        importExcelButton, exportExcelButton,
        widget.NewSeparator(),
        clearButton, statsButton,
    )
    
    return toolbar
}

func (a *App) createStatusBar() fyne.CanvasObject {
    statusLabel := widget.NewLabel("Готов к работе")
    
    a.updateStatusBar = func(text string) {
        statusLabel.SetText(text)
    }
    
    return container.NewHBox(statusLabel)
}

func (a *App) refreshTable() {
    books, err := a.database.GetAllBooks()
    if err != nil {
        fmt.Printf("Ошибка загрузки книг: %v\n", err)
        a.books = []database.BookView{}
    } else {
        a.books = books
    }
    
    if a.table != nil {
        a.table.Refresh()
    }
    
    if a.updateStatusBar != nil {
        a.updateStatusBar(fmt.Sprintf("Загружено книг: %d", len(a.books)))
    }
}

func (a *App) createTable() *widget.Table {
    table := widget.NewTable(
        func() (int, int) {
            return len(a.books) + 1, 5
        },
        func() fyne.CanvasObject {
            return widget.NewLabel("template")
        },
        func(id widget.TableCellID, cell fyne.CanvasObject) {
            label := cell.(*widget.Label)
            if id.Row == 0 {
                headers := []string{"ID", "Название", "Автор", "Год", "Тираж"}
                if id.Col < len(headers) {
                    label.SetText(headers[id.Col])
                }
            } else {
                if id.Row-1 < len(a.books) {
                    book := a.books[id.Row-1]
                    switch id.Col {
                    case 0:
                        label.SetText(fmt.Sprintf("%d", book.ID))
                    case 1:
                        label.SetText(book.Title)
                    case 2:
                        label.SetText(book.Author)
                    case 3:
                        label.SetText(fmt.Sprintf("%d", book.Year))
                    case 4:
                        label.SetText(fmt.Sprintf("%d", book.Copies))
                    }
                }
            }
        },
    )
    
    table.SetColumnWidth(0, 60)
    table.SetColumnWidth(1, 300)
    table.SetColumnWidth(2, 200)
    table.SetColumnWidth(3, 80)
    table.SetColumnWidth(4, 100)
    
    return table
}