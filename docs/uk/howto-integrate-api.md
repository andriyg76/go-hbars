# Як інтегрувати API go-hbars у програму (шаблони + go:generate)

Покрокова інструкція додавання Handlebars-шаблонів до Go-проекту з використанням go-hbars з GitHub та `go:generate` з компілятором (hbc). Без локального `replace` — залежність береться з GitHub.

**Альтернатива:** команда [init](init.md) створює новий проєкт або додає шаблони до існуючого модуля: `go run github.com/andriyg76/go-hbars/cmd/init@latest new myapp` або `init add`.

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

У `main.go` (або будь-якому пакеті) імпортуйте пакет шаблонів і викликайте згенеровані функції рендеру. Компілятор генерує **типізовані контексти** (наприклад `MainContext`); коли дані у вигляді `map[string]any` (наприклад з JSON), використовуйте згенерований `XxxContextFromMap`, щоб вони задовольняли інтерфейс контексту:

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

	// Рендер у рядок. Для даних-мапи використовуйте MainContextFromMap, щоб задовольнити MainContext.
	out, err := templates.RenderMainString(templates.MainContextFromMap(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(out)
}
```

Для кожного шаблону `name.hbs` згенерований пакет надає:

- `RenderName(w io.Writer, data NameContext) error`
- `RenderNameString(data NameContext) (string, error)`
- `NameContextFromMap(m map[string]any) NameContext` — використовуйте для даних-мапи (наприклад з JSON).

(Як імена файлів шаблонів відповідають іменам Go-функцій і що містить згенерований файл: див. [Скомпільований файл шаблонів](compiled-templates.md). Повний API: [API шаблонів](api.md).)

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
| 4 | У main: імпорт templates, завантажити дані, викликати `templates.RenderXxxString(templates.XxxContextFromMap(data))` для даних-мапи |
| 5 | Запуск: `go run .` |

У `go.mod` не потрібен `replace`; залежність береться з GitHub. Щоб зафіксувати версію, використовуйте конкретний тег замість `@latest` у рядку go:generate (наприклад `@v0.1.0`, коли буде доступний).

## Робота з локальним чекаутом

Якщо розробляєте go-hbars або тестуєте зміни до релізу:

1. Клонуйте репозиторій локально, наприклад `~/src/go-hbars`.
2. У `go.mod` вашого застосунку додайте `replace`, щоб модуль вказував на локальну копію:

   ```go
   replace github.com/andriyg76/go-hbars => /home/you/src/go-hbars
   ```

3. Залиште той самий рядок `//go:generate` (з `@latest` чи без). Під час `go generate ./...` команда `go run` використовуватиме замінений модуль, тобто ваш локальний hbc.
4. З кореня репо go-hbars можна також запускати компілятор напряму: `go run ./cmd/hbc -in /шлях/до/templates -out /шлях/до/templates_gen.go -pkg templates`.

У згенерованих файлах з’являється коментар `// Generator version: ...`, якщо компілятор зібрано з інформацією про версію (наприклад у CLI hbc).
