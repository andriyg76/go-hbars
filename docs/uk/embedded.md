# Вбудований процесор та веб-сервер

go-hbars надає API для вбудовування генерації сайту та веб-сервера у ваші Go-застосунки.

## Швидкий старт з bootstrap-кодом

Якщо при генерації шаблонів ви використовували прапорець `-bootstrap`, можна використовувати швидкі функції:

### Quick Processor

```go
import "github.com/your/project/templates"

// Швидкий процесор із конфігурацією за замовчуванням
proc, err := templates.NewQuickProcessor()
if err != nil {
    log.Fatal(err)
}

// За потреби налаштуйте конфігурацію
proc.Config().DataPath = "content"
proc.Config().OutputPath = "build"

if err := proc.Process(); err != nil {
    log.Fatal(err)
}
```

### Quick Server

```go
import "github.com/your/project/templates"

// Швидкий сервер із конфігурацією за замовчуванням
srv, err := templates.NewQuickServer()
if err != nil {
    log.Fatal(err)
}

// За потреби налаштуйте конфігурацію
srv.Config().DataPath = "content"
srv.Config().Addr = ":3000"

log.Fatal(srv.Start())
```

## Пряме використання API

### Генерація статичного сайту

```go
import (
    "github.com/andriyg76/go-hbars/pkg/sitegen"
    "github.com/your/project/templates"
)

config := sitegen.DefaultConfig()
config.DataPath = "data"
config.OutputPath = "pages"

// Створити рендерер зі скомпільованих функцій шаблонів
renderer := sitegen.NewRendererFromFunctions(map[string]sitegen.RenderFunc{
    "main":   templates.RenderMain,
    "header": templates.RenderHeader,
    "footer": templates.RenderFooter,
})

proc, err := sitegen.NewProcessor(config, renderer)
if err != nil {
    log.Fatal(err)
}

if err := proc.Process(); err != nil {
    log.Fatal(err)
}
```

### Напівстатичний веб-сервер

```go
import (
    "github.com/andriyg76/go-hbars/pkg/sitegen"
    "github.com/your/project/templates"
)

config := sitegen.DefaultConfig()
config.DataPath = "data"
config.Addr = ":8080"

// Створити рендерер зі скомпільованих функцій шаблонів
renderer := sitegen.NewRendererFromFunctions(map[string]sitegen.RenderFunc{
    "main":   templates.RenderMain,
    "header": templates.RenderHeader,
    "footer": templates.RenderFooter,
})

srv, err := sitegen.NewServer(config, renderer)
if err != nil {
    log.Fatal(err)
}

log.Fatal(srv.Start())
```

## Довідник API

### Конфігурація

```go
type Config struct {
    RootPath      string // Базова директорія для відносних шляхів
    DataPath      string // Шлях до директорії з файлами даних (за замовчуванням: "data")
    SharedPath    string // Шлях до директорії спільних даних (за замовчуванням: "shared")
    OutputPath    string // Шлях до директорії виводу для статичної генерації (за замовчуванням: "pages")
    StaticDir     string // Шлях до директорії статичних файлів для сервера (опційно)
    Addr          string // Адреса прослуховування для сервера (за замовчуванням: ":8080")
}
```

### Процесор

```go
// NewProcessor створює новий процесор із заданою конфігурацією та рендерером
proc, err := sitegen.NewProcessor(config, renderer)

// Process обробляє всі файли даних і генерує вихідні файли
err := proc.Process()

// ProcessFile обробляє один файл даних і повертає шлях виводу та вміст
outputPath, content, err := proc.ProcessFile(dataFilePath)

// Config повертає конфігурацію процесора
config := proc.Config()
```

### Сервер

```go
// NewServer створює новий сервер із заданою конфігурацією та рендерером
srv, err := sitegen.NewServer(config, renderer)

// Start запускає HTTP-сервер
err := srv.Start()

// StartTLS запускає HTTP-сервер з TLS
err := srv.StartTLS(certFile, keyFile)

// Shutdown коректно зупиняє сервер
err := srv.Shutdown()

// Address повертає адресу сервера
addr := srv.Address()

// Config повертає конфігурацію сервера
config := srv.Config()
```

### Рендерер

```go
// NewRendererFromFunctions створює рендерер з мапи імен шаблонів на функції рендеру
renderer := sitegen.NewRendererFromFunctions(map[string]sitegen.RenderFunc{
    "main":   templates.RenderMain,
    "header": templates.RenderHeader,
})

// LoadRenderer намагається автоматично знайти та завантажити функції рендеру
renderer, err := sitegen.LoadRenderer(templatePackage)
```

## Розширене використання

### Власний рендерер

Можна створити власний рендерер, реалізувавши інтерфейс `renderer.TemplateRenderer`:

```go
type TemplateRenderer interface {
    Render(templateName string, w io.Writer, data any) error
}
```

### Обробка окремих файлів

```go
proc, _ := sitegen.NewProcessor(config, renderer)

// Обробити один файл
outputPath, content, err := proc.ProcessFile("data/blog/post.json")
if err != nil {
    log.Fatal(err)
}

// Записати вивід вручну
os.WriteFile(outputPath, content, 0644)
```

### Сервер з власним обробником

Сервер використовує внутрішній обробник, який обробляє файли даних на льоту. Можна розширити це, створивши власний HTTP-обробник на основі процесора:

```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    _, content, err := proc.ProcessFile("data/index.json")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Write(content)
})
```
