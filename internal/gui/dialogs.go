package gui

import (
    "github.com/nydeg/bd/internal/database"
    "fmt"
    "strconv"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/layout"
    "fyne.io/fyne/v2/storage"
    "fyne.io/fyne/v2/widget"
)

func (a *App) showAddDialog() {
    idEntry := widget.NewEntry()
    titleEntry := widget.NewEntry()
    authorEntry := widget.NewEntry()
    yearEntry := widget.NewEntry()
    copiesEntry := widget.NewEntry()

    form := &widget.Form{
        Items: []*widget.FormItem{
            {Text: "ID", Widget: idEntry},
            {Text: "Название", Widget: titleEntry},
            {Text: "Автор", Widget: authorEntry},
            {Text: "Год издания", Widget: yearEntry},
            {Text: "Тираж", Widget: copiesEntry},
        },
        OnSubmit: func() {
            id, err := strconv.Atoi(idEntry.Text)
            if err != nil {
                dialog.ShowError(err, a.window)
                return
            }

            year, err := strconv.Atoi(yearEntry.Text)
            if err != nil {
                dialog.ShowError(err, a.window)
                return
            }

            copies, err := strconv.Atoi(copiesEntry.Text)
            if err != nil {
                dialog.ShowError(err, a.window)
                return
            }

            book := database.BookView{
                ID:     int32(id),
                Title:  titleEntry.Text,
                Author: authorEntry.Text,
                Year:   int32(year),
                Copies: int32(copies),
            }

            if err := a.database.AddBook(book); err != nil {
                dialog.ShowError(err, a.window)
            } else {
                dialog.ShowInformation("Успех", "Книга добавлена", a.window)
                a.refreshTable()
            }
        },
    }

    customDialog := dialog.NewCustomConfirm("Добавить книгу", "Добавить", "Отмена", 
        container.NewVBox(form), 
        func(b bool) {
            if b {
                form.OnSubmit()
            }
        }, a.window)
    
    customDialog.Resize(fyne.NewSize(600, 400))
    customDialog.Show()
}

func (a *App) showEditDialog() {
    idEntry := widget.NewEntry()
    idEntry.SetPlaceHolder("Введите ID книги")

    infoLabel := widget.NewLabel("")
    infoLabel.Wrapping = fyne.TextWrapWord

    var currentBook *database.BookView

    loadBook := func() {
        if idEntry.Text == "" {
            infoLabel.SetText("Введите ID книги")
            return
        }

        id, err := strconv.Atoi(idEntry.Text)
        if err != nil {
            infoLabel.SetText("❌ ID должен быть числом")
            return
        }

        book, err := a.database.FindByID(int32(id))
        if err != nil {
            infoLabel.SetText(fmt.Sprintf("❌ Книга с ID %d не найдена", id))
            return
        }

        currentBook = &database.BookView{
            ID:     book.ID,
            Title:  database.BytesToString(book.Title[:]),
            Author: database.BytesToString(book.Author[:]),
            Year:   book.Year,
            Copies: book.Copies,
        }

        infoLabel.SetText(fmt.Sprintf(
            "📖 Найдена книга:\nНазвание: %s\nАвтор: %s\nГод: %d\nТираж: %d",
            currentBook.Title, currentBook.Author, currentBook.Year, currentBook.Copies,
        ))
    }

    loadButton := widget.NewButton("Загрузить книгу", loadBook)

    formStep1 := container.NewVBox(
        widget.NewLabel("Введите ID книги для редактирования:"),
        idEntry,
        loadButton,
        infoLabel,
    )

    customDialog := dialog.NewCustomConfirm("Редактирование книги", "Продолжить", "Отмена", 
        formStep1,
        func(continueEditing bool) {
            if continueEditing && currentBook != nil {
                a.showEditForm(*currentBook)
            }
        }, a.window)
    
    customDialog.Resize(fyne.NewSize(500, 300))
    customDialog.Show()
}

func (a *App) showEditForm(book database.BookView) {
    titleEntry := widget.NewEntry()
    titleEntry.SetText(book.Title)
    titleEntry.SetPlaceHolder("Введите название книги")

    authorEntry := widget.NewEntry()
    authorEntry.SetText(book.Author)
    authorEntry.SetPlaceHolder("Введите автора")

    yearEntry := widget.NewEntry()
    yearEntry.SetText(fmt.Sprintf("%d", book.Year))
    yearEntry.SetPlaceHolder("Введите год издания")

    copiesEntry := widget.NewEntry()
    copiesEntry.SetText(fmt.Sprintf("%d", book.Copies))
    copiesEntry.SetPlaceHolder("Введите тираж")

    clearTitle := func() {
        titleEntry.SetText("")
    }

    clearAuthor := func() {
        authorEntry.SetText("")
    }

    clearYear := func() {
        yearEntry.SetText("")
    }

    clearCopies := func() {
        copiesEntry.SetText("")
    }

    titleContainer := container.NewBorder(nil, nil, nil, 
        widget.NewButton("Очистить", clearTitle), titleEntry)

    authorContainer := container.NewBorder(nil, nil, nil, 
        widget.NewButton("Очистить", clearAuthor), authorEntry)

    yearContainer := container.NewBorder(nil, nil, nil, 
        widget.NewButton("Очистить", clearYear), yearEntry)

    copiesContainer := container.NewBorder(nil, nil, nil, 
        widget.NewButton("Очистить", clearCopies), copiesEntry)

    infoText := fmt.Sprintf("Редактирование книги ID: %d\n\nОставьте поле пустым, чтобы сохранить текущее значение\nНажмите 'Очистить', чтобы стереть поле", book.ID)
    infoLabel := widget.NewLabel(infoText)
    infoLabel.Wrapping = fyne.TextWrapWord

    form := &widget.Form{
        Items: []*widget.FormItem{
            {Text: "Информация", Widget: infoLabel},
            {Text: "Название книги", Widget: titleContainer},
            {Text: "Автор", Widget: authorContainer},
            {Text: "Год издания", Widget: yearContainer},
            {Text: "Тираж", Widget: copiesContainer},
        },
        OnSubmit: func() {
            updatedBook := database.BookView{
                ID: book.ID,
            }

            if titleEntry.Text == "" {
                updatedBook.Title = book.Title
            } else {
                updatedBook.Title = titleEntry.Text
            }

            if authorEntry.Text == "" {
                updatedBook.Author = book.Author
            } else {
                updatedBook.Author = authorEntry.Text
            }

            if yearEntry.Text == "" {
                updatedBook.Year = book.Year
            } else {
                year, err := strconv.Atoi(yearEntry.Text)
                if err != nil {
                    dialog.ShowError(fmt.Errorf("год должен быть числом"), a.window)
                    return
                }
                updatedBook.Year = int32(year)
            }

            if copiesEntry.Text == "" {
                updatedBook.Copies = book.Copies
            } else {
                copies, err := strconv.Atoi(copiesEntry.Text)
                if err != nil {
                    dialog.ShowError(fmt.Errorf("тираж должен быть числом"), a.window)
                    return
                }
                updatedBook.Copies = int32(copies)
            }

            if updatedBook.Title == "" {
                dialog.ShowError(fmt.Errorf("название книги не может быть пустым"), a.window)
                return
            }

            if updatedBook.Author == "" {
                dialog.ShowError(fmt.Errorf("автор не может быть пустым"), a.window)
                return
            }

            if err := a.database.UpdateBook(updatedBook); err != nil {
                dialog.ShowError(err, a.window)
            } else {
                dialog.ShowInformation("Успех", "Книга успешно обновлена", a.window)
                a.refreshTable()
            }
        },
    }

    content := container.NewVBox(form)
    customDialog := dialog.NewCustomConfirm("Редактирование книги", "Сохранить", "Отмена", 
        content,
        func(save bool) {
            if save {
                form.OnSubmit()
            }
        }, a.window)
    
    customDialog.Resize(fyne.NewSize(600, 500))
    customDialog.Show()
}

func (a *App) showDeleteDialog() {
    idEntry := widget.NewEntry()

    form := &widget.Form{
        Items: []*widget.FormItem{
            {Text: "ID книги для удаления", Widget: idEntry},
        },
        OnSubmit: func() {
            id, err := strconv.Atoi(idEntry.Text)
            if err != nil {
                dialog.ShowError(err, a.window)
                return
            }

            if err := a.database.DeleteBook(int32(id)); err != nil {
                dialog.ShowError(err, a.window)
            } else {
                dialog.ShowInformation("Успех", "Книга удалена", a.window)
                a.refreshTable()
            }
        },
    }

    customDialog := dialog.NewCustomConfirm("Удалить книгу", "Удалить", "Отмена", 
        form,
        func(b bool) {
            if b {
                form.OnSubmit()
            }
        }, a.window)
    
    customDialog.Resize(fyne.NewSize(400, 200))
    customDialog.Show()
}

func (a *App) showClearDatabaseDialog() {
    confirmDialog := dialog.NewConfirm("Очистка базы данных", 
        "ВНИМАНИЕ! Вы собираетесь полностью очистить базу данных.\nВсе книги будут удалены без возможности восстановления.\n\nПродолжить?",
        func(confirmed bool) {
            if confirmed {
                if err := a.database.ClearDatabase(); err != nil {
                    dialog.ShowError(fmt.Errorf("ошибка очистки БД: %v", err), a.window)
                } else {
                    dialog.ShowInformation("Успех", "База данных полностью очищена", a.window)
                    a.refreshTable()
                }
            }
        }, a.window)
    confirmDialog.Show()
}

func (a *App) showSearchDialog() {
    var searchResults []database.BookView
    var currentSearchField string
    var currentSearchValue string

    searchFieldSelect := widget.NewSelect([]string{
        "ID", 
        "Название", 
        "Автор", 
        "Год издания", 
        "Тираж",
    }, func(value string) {
        currentSearchField = value
    })
    searchFieldSelect.SetSelected("Название")

    searchValueEntry := widget.NewEntry()
    searchValueEntry.SetPlaceHolder("Введите значение для поиска...")
    searchValueEntry.Resize(fyne.NewSize(500, searchValueEntry.MinSize().Height))

    resultsLabel := widget.NewLabel("Результаты не найдены")
    updateResultsLabel := func() {
        if len(searchResults) == 0 {
            resultsLabel.SetText("Результаты не найдены")
        } else {
            resultsLabel.SetText(fmt.Sprintf("Найдено книг: %d", len(searchResults)))
        }
    }

    resultsTable := widget.NewTable(
        func() (int, int) {
            return len(searchResults) + 1, 5
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
                if id.Row-1 < len(searchResults) {
                    book := searchResults[id.Row-1]
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

    resultsTable.SetColumnWidth(0, 80)
    resultsTable.SetColumnWidth(1, 350) // Увеличиваем ширину колонки названия
    resultsTable.SetColumnWidth(2, 250) // Увеличиваем ширину колонки автора
    resultsTable.SetColumnWidth(3, 100)
    resultsTable.SetColumnWidth(4, 100)

    performSearch := func() {
        if currentSearchField == "" || searchValueEntry.Text == "" {
            dialog.ShowInformation("Ошибка", "Выберите поле и введите значение для поиска", a.window)
            return
        }

        currentSearchValue = searchValueEntry.Text
        results, err := a.database.FindBooks(currentSearchField, currentSearchValue)
        if err != nil {
            searchResults = []database.BookView{}
            dialog.ShowError(err, a.window)
        } else {
            searchResults = results
        }
        resultsTable.Refresh()
        updateResultsLabel()
    }

    searchButton := widget.NewButton("Найти", performSearch)
    clearButton := widget.NewButton("Очистить", func() {
        searchValueEntry.SetText("")
        searchResults = []database.BookView{}
        resultsTable.Refresh()
        updateResultsLabel()
    })

    searchValueEntry.OnSubmitted = func(_ string) {
        performSearch()
    }

    searchContent := container.NewVBox(
        widget.NewLabel("Поиск книг:"),
        container.NewHBox(
            widget.NewLabel("Поле поиска:"),
            searchFieldSelect,
        ),
        container.NewHBox(
            widget.NewLabel("Значение:"),
            searchValueEntry,
        ),
        container.NewHBox(
            searchButton,
            clearButton,
        ),
        widget.NewSeparator(),
        resultsLabel,
        container.NewStack(resultsTable),
    )

    scrollContainer := container.NewScroll(searchContent)
    scrollContainer.SetMinSize(fyne.NewSize(900, 600))

    closeButton := widget.NewButton("Закрыть", func() {})
    
    finalContainer := container.NewBorder(
        nil, 
        container.NewHBox(layout.NewSpacer(), closeButton), 
        nil, nil, 
        scrollContainer,
    )

    searchDialog := dialog.NewCustomConfirm("Поиск книг", "Закрыть", "", 
        finalContainer,
        func(close bool) {
        }, a.window)
    
    closeButton.OnTapped = func() {
        searchDialog.Hide()
    }
    
    searchDialog.Resize(fyne.NewSize(920, 650))
    searchDialog.Show()
}

func (a *App) showExportDialog() {
    fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
        if err != nil {
            dialog.ShowError(err, a.window)
            return
        }
        if writer == nil {
            return
        }
        defer writer.Close()

        if err := a.database.ExportToTxt(writer.URI().Path()); err != nil {
            dialog.ShowError(fmt.Errorf("ошибка экспорта: %v", err), a.window)
        } else {
            dialog.ShowInformation("Успех", "Данные успешно экспортированы в файл", a.window)
        }
    }, a.window)

    fileDialog.SetFileName("books_export.txt")
    fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))
    fileDialog.Show()
}

func (a *App) showImportDialog() {
    fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
        if err != nil {
            dialog.ShowError(err, a.window)
            return
        }
        if reader == nil {
            return
        }
        defer reader.Close()

        confirmDialog := dialog.NewConfirm("Импорт данных", 
            "Внимание! При импорте:\n- Новые книги будут добавлены\n- Существующие книги с одинаковым ID будут обновлены\n\nПродолжить?",
            func(confirmed bool) {
                if confirmed {
                    count, err := a.database.ImportFromTxt(reader.URI().Path())
                    if err != nil {
                        dialog.ShowError(fmt.Errorf("ошибка импорта: %v", err), a.window)
                    } else {
                        dialog.ShowInformation("Успех", 
                            fmt.Sprintf("Импорт завершен!\nДобавлено/обновлено книг: %d", count), a.window)
                        a.refreshTable()
                    }
                }
            }, a.window)
        confirmDialog.Show()
    }, a.window)

    fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))
    fileDialog.Show()
}

func (a *App) showStatsDialog() {
    count, size, err := a.database.GetStats()
    if err != nil {
        dialog.ShowError(err, a.window)
        return
    }

    statsText := fmt.Sprintf(
        "Статистика базы данных:\n\n"+
        "📊 Количество книг: %d\n"+
        "💾 Размер файла БД: %.2f КБ\n"+
        "📁 Размер одной записи: %d байт",
        count, float64(size)/1024, 152,
    )

    dialog.ShowInformation("Статистика", statsText, a.window)
}

func (a *App) showExportExcelDialog() {
    fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
        if err != nil {
            dialog.ShowError(err, a.window)
            return
        }
        if writer == nil {
            return
        }
        defer writer.Close()

        if err := a.database.ExportToExcel(writer.URI().Path()); err != nil {
            dialog.ShowError(fmt.Errorf("ошибка экспорта: %v", err), a.window)
        } else {
            dialog.ShowInformation("Успех", "Данные успешно экспортированы в Excel", a.window)
        }
    }, a.window)

    fileDialog.SetFileName("books_export.xlsx")
    fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
    fileDialog.Show()
}

func (a *App) showImportExcelDialog() {
    fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
        if err != nil {
            dialog.ShowError(err, a.window)
            return
        }
        if reader == nil {
            return
        }
        defer reader.Close()

        confirmDialog := dialog.NewConfirm("Импорт данных из Excel", 
            "Внимание! При импорте:\n- Новые книги будут добавлены\n- Существующие книги с одинаковым ID будут обновлены\n\nПродолжить?",
            func(confirmed bool) {
                if confirmed {
                    count, err := a.database.ImportFromExcel(reader.URI().Path())
                    if err != nil {
                        dialog.ShowError(fmt.Errorf("ошибка импорта: %v", err), a.window)
                    } else {
                        dialog.ShowInformation("Успех", 
                            fmt.Sprintf("Импорт завершен!\nДобавлено/обновлено книг: %d", count), a.window)
                        a.refreshTable()
                    }
                }
            }, a.window)
        confirmDialog.Show()
    }, a.window)

    fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
    fileDialog.Show()
}

