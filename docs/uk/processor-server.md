# Статичний процесор та веб-сервер

go-hbars включає CLI-інструменти для генерації статичних сайтів та запуску напівстатичного веб-сервера.

## Швидкий старт

1. **Структура проекту:**
```
project/
├── processor/
│   ├── templates/
│   │   ├── main.hbs
│   │   ├── header.hbs
│   │   └── footer.hbs
│   └── templates_gen.go  # з директивою go:generate
├── data/
│   └── index.json
└── shared/
    └── site.json
```

2. **Додати директиву go:generate для компіляції шаблонів:**

Створіть `processor/templates/gen.go`:
```go
//go:generate hbc -in . -out ./templates_gen.go -pkg templates -bootstrap

package templates
```

Прапорець `-bootstrap` генерує допоміжні функції для швидкого запуску сервера та процесора.

3. **Згенерувати шаблони:**
```bash
go generate ./...
```

4. **Створити файли даних:**

`data/index.json`:
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

`shared/site.json`:
```json
{
  "name": "My Site",
  "url": "https://example.com"
}
```

## Генерація статичного сайту

Генерація статичних HTML-файлів для хостингу:

```bash
go run ./cmd/build --data-path data --output-path pages
```

### Параметри CLI

**Команда build (`cmd/build`):**
- `--root` — базова директорія для відносних шляхів (за замовчуванням: поточна)
- `--data-path` — директорія з файлами даних (за замовчуванням: `data`)
- `--shared-path` — директорія спільних даних (за замовчуванням: `shared`)
- `--output-path` — директорія виводу (за замовчуванням: `pages`)

## Напівстатичний веб-сервер

Запуск сервера для розробки, який генерує сторінки на льоту:

```bash
go run ./cmd/server --data-path data --addr :8080
```

### Параметри CLI

**Команда server (`cmd/server`):**
- `--root` — базова директорія для відносних шляхів (за замовчуванням: поточна)
- `--data-path` — директорія з файлами даних (за замовчуванням: `data`)
- `--shared-path` — директорія спільних даних (за замовчуванням: `shared`)
- `--static-dir` — директорія статичних файлів (опційно)
- `--addr` — адреса прослуховування (за замовчуванням: `:8080`)

## Формат файлів даних

Кожен файл даних повинен містити секцію `_page`:

**JSON:**
```json
{
  "_page": {
    "template": "blog/post",
    "output": "blog/hello.html"
  },
  "title": "Hello",
  "author": "Ada"
}
```

**YAML:**
```yaml
_page:
  template: blog/post
  output: blog/hello.html
title: Hello
author: Ada
```

**TOML:**
```toml
[_page]
template = "blog/post"
output = "blog/hello.html"

title = "Hello"
author = "Ada"
```

У секції `_page`:
- `template` — ім’я шаблону (без розширення `.hbs`)
- `output` — опційний шлях виводу (відносно директорії виводу). Якщо не вказано, використовується ім’я вхідного файлу з розширенням `.html`.

## Спільні дані

Файли спільних даних завантажуються з директорії `shared/` і підмешуються в усі сторінки під ключем `_shared`:

**Структура:**
```
shared/
  site.json
  navigation/
    menu.yaml
```

**У шаблонах:**
```handlebars
<title>{{_shared.site.name}}</title>
<nav>
  {{#each _shared.navigation.menu.items}}
    <a href="{{href}}">{{label}}</a>
  {{/each}}
</nav>
```

## Використання go:generate

Рекомендований підхід — використовувати `go:generate` для автоматичної компіляції шаблонів:

**У файлі пакету шаблонів (`processor/templates/gen.go`):**
```go
//go:generate hbc -in . -out ./templates_gen.go -pkg templates -bootstrap

package templates
```

**Переваги:**
- Шаблони перекомпільовуються при виконанні `go generate ./...`
- Зручна інтеграція з `go build` та CI/CD
- Не потрібен ручний крок компіляції
- Шаблони перевіряються під час компіляції

**Робочий процес:**
1. Редагувати шаблони в `processor/templates/`
2. Виконати `go generate ./...` для перекомпіляції
3. Зібрати та запустити застосунок

**Інтеграція в CI/CD:**
```yaml
# Приклад GitHub Actions
- name: Generate templates
  run: go generate ./...

- name: Build
  run: go build ./...
```

## Див. також

- [init](init.md) — створити проєкт з `processor/templates`, `data/`, `shared/` через `init new -bootstrap`.
- [Як інтегрувати bootstrap](howto-integrate-bootstrap.md) — покрокова налаштування bootstrap.
- [Вбудований API](embedded.md) — використання процесора та сервера програмно.
