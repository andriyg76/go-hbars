# Як інтегрувати bootstrap go-hbars (QuickServer + QuickProcessor)

Покрокова інструкція додавання Handlebars-шаблонів з **bootstrap**-кодом: `-bootstrap` генерує `NewQuickServer()` та `NewQuickProcessor()`, щоб можна було запустити напівстатичний HTTP-сервер або генерувати статичний HTML з файлів даних. Використовується go-hbars з GitHub (без локального `replace` у продакшені).

## 1. Створити новий проект

```bash
mkdir myapp && cd myapp
go mod init myapp
```

## 2. Додати шаблони та go:generate з -bootstrap

Створіть директорію для шаблонів. Оскільки Go не дозволяє імпортні шляхи з крапкою на початку (наприклад `.processor/templates`), використовуйте шлях на кшталт `processor/templates/` або `templates/`.

Покладіть туди файли `.hbs` (наприклад `main.hbs`, `header.hbs`, `footer.hbs`).

Додайте файл, який запускає генерацію **з прапорцем `-bootstrap`**. Наприклад `processor/templates/gen.go`:

```go
//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc@latest -in . -out ./templates_gen.go -pkg templates -bootstrap

package templates
```

- `-bootstrap` — додатково до функцій `RenderXxx` генеруються `NewQuickServer()` та `NewQuickProcessor()`.
- Згенерований пакет імпортує лише публічні пакети (`pkg/renderer`, `pkg/sitegen`), тому його можна використовувати з вашого модуля без імпорту internal-пакетів.

## 3. Згенерувати код шаблонів

З кореня проекту:

```bash
go generate ./...
go mod tidy
```

## 4. Додати файли даних з `_page`

Кожен файл даних, з якого має вийти сторінка, повинен містити секцію `_page`:

- `template` — ім'я шаблону (наприклад `main`, відповідає `main.hbs`).
- `output` — шлях виводу відносно директорії виводу (наприклад `index.html`, `blog/post.html`).

Приклад `data/index.json`:

```json
{
  "_page": {
    "template": "main",
    "output": "index.html"
  },
  "title": "Welcome",
  "content": "Hello, world!"
}
```

Створіть директорію `data/` і додайте один або кілька JSON- (або YAML/TOML-) файлів з `_page` та даними для шаблону.

## 5. Використати QuickProcessor (генерація статичного сайту)

У `main.go` (або CLI-команді) використовуйте згенерований `NewQuickProcessor()` та за потреби налаштуйте шляхи:

```go
package main

import (
	"log"

	templates "myapp/processor/templates"
)

func main() {
	proc, err := templates.NewQuickProcessor()
	if err != nil {
		log.Fatal(err)
	}
	proc.Config().DataPath = "data"
	proc.Config().OutputPath = "pages"

	if err := proc.Process(); err != nil {
		log.Fatal(err)
	}
}
```

Запуск: `go run .` (або збірка й запуск). Це читає всі файли з `data/`, підмешовує спільні дані з `shared/` (якщо є) і записує HTML у `pages/` (або ваш `OutputPath`).

## 6. Використати QuickServer (сервер для розробки)

Щоб запустити напівстатичний HTTP-сервер, який рендерить сторінки на вимогу:

```go
	srv, err := templates.NewQuickServer()
	if err != nil {
		log.Fatal(err)
	}
	srv.Config().DataPath = "data"
	srv.Config().Addr = ":8080"

	log.Fatal(srv.Start())
```

Сервер зіставляє URL з файлами даних (наприклад `/` → `data/index.json`, `/about` → `data/about.json`) і рендерить їх указаним шаблоном.

## 7. Спільні дані (опційно)

Створіть директорію `shared/`. Файли JSON/YAML/TOML у ній завантажуються і підмешуються в кожну сторінку під ключем `_shared`. У шаблонах використовуйте `{{_shared.site.name}}` тощо.

## Підсумок

| Крок | Дія |
|------|-----|
| 1 | Новий модуль: `go mod init myapp` |
| 2 | Додати `processor/templates/*.hbs` та `processor/templates/gen.go` з `//go:generate ... hbc@latest ... -bootstrap` |
| 3 | Виконати `go generate ./...`, потім `go mod tidy` |
| 4 | Додати `data/*.json` (або YAML/TOML) з `_page.template` та `_page.output` |
| 5 | У main: `templates.NewQuickProcessor()` та `proc.Process()` для статичної збірки, або `templates.NewQuickServer()` та `srv.Start()` для HTTP |
| 6 | Встановити Config().DataPath, OutputPath/Addr під свою структуру |

Bootstrap використовує лише публічні пакети (`pkg/renderer`, `pkg/sitegen`), тому проект може залежати від go-hbars з GitHub без локального `replace`.

(Що саме генерує bootstrap і інтерфейс для розробника: див. [Згенерований bootstrap](bootstrap-generated.md).)
