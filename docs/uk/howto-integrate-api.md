# Як інтегрувати API go-hbars у програму (шаблони + go:generate)

Покрокова інструкція додавання Handlebars-шаблонів до Go-проекту з використанням go-hbars з GitHub та `go:generate` з компілятором (hbc). Без локального `replace` — залежність береться з GitHub.

## 1. Створити новий проект

```bash
mkdir myapp && cd myapp
go mod init myapp
```

## 2. Додати пакет шаблонів та go:generate

Створіть директорію для шаблонів, наприклад `templates/`, і покладіть туди файли `.hbs` (наприклад `main.hbs`, `header.hbs`, `footer.hbs`).

Додайте файл, який запускає генерацію коду. Наприклад `templates/gen.go`:

```go
//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc@latest -in . -out ./templates_gen.go -pkg templates

package templates
```

- `-in .` — поточна директорія (папка пакету `templates/`) є коренем шаблонів.
- `-out ./templates_gen.go` — згенерований Go-файл у тому ж пакеті.
- `-pkg templates` — ім'я пакету для згенерованого коду.

Використання `go run .../cmd/hbc@latest` запускає компілятор з GitHub; окремо встановлювати `hbc` не потрібно.

## 3. Згенерувати код шаблонів

З кореня проекту:

```bash
go generate ./...
go mod tidy
```

Це за потреби завантажить go-hbars (і hbc), згенерує `templates_gen.go` та оновить `go.mod`/`go.sum`.

## 4. Використати згенерований API у програмі

У `main.go` (або будь-якому пакеті) імпортуйте пакет шаблонів і викликайте згенеровані функції рендеру:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	templates "myapp/templates"
)

func main() {
	dataBytes, err := os.ReadFile("data.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read data: %v\n", err)
		os.Exit(1)
	}
	var data map[string]any
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		fmt.Fprintf(os.Stderr, "parse data: %v\n", err)
		os.Exit(1)
	}

	// Рендер у рядок (ім'я шаблону = ім'я файлу без .hbs, напр. main -> RenderMainString)
	out, err := templates.RenderMainString(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(out)
}
```

Для кожного шаблону `name.hbs` згенерований пакет надає:

- `RenderName(w io.Writer, data any) error`
- `RenderNameString(data any) (string, error)`

(Як імена файлів шаблонів відповідають іменам Go-функцій і що містить згенерований файл: див. [Скомпільований файл шаблонів](compiled-templates.md).)

## 5. Запустити програму

```bash
go run .
```

## Підсумок

| Крок | Дія |
|------|-----|
| 1 | Новий модуль: `go mod init myapp` |
| 2 | Додати `templates/*.hbs` та `templates/gen.go` з `//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc@latest -in . -out ./templates_gen.go -pkg templates` |
| 3 | Виконати `go generate ./...`, потім `go mod tidy` |
| 4 | У main: імпорт templates, завантажити дані, викликати `templates.RenderXxxString(data)` |
| 5 | Запуск: `go run .` |

У `go.mod` не потрібен `replace`; залежність береться з GitHub. Щоб зафіксувати версію, використовуйте конкретний тег замість `@latest` у рядку go:generate (наприклад `@v0.1.0`, коли буде доступний).
