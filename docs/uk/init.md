# go-hbars init: створити проект або додати до нього

Команда `init` створює новий проєкт на основі go-hbars або додає шаблони та bootstrap до існуючого Go-модуля. Запуск:

```bash
go run github.com/andriyg76/go-hbars/cmd/init@latest <підкоманда> [опції]
```

Або встановити й викликати:

```bash
go install github.com/andriyg76/go-hbars/cmd/init@latest
init new myapp -bootstrap
```

## Підкоманди

### init new [path]

Створює новий проєкт. **path** за замовчуванням — поточна директорія (`.`).

| Прапорець | Опис |
|-----------|------|
| `-bootstrap` | Додати `processor/templates/`, `data/`, `shared/` та main з `NewQuickServer()` / `NewQuickProcessor()`. Без нього створюється простий API-проєкт з `templates/` та `RenderXxxString`. |
| `-module` | Шлях модуля для `go mod init` (за замовчуванням — ім’я директорії). |

**Приклади:**

```bash
# Новий проєкт лише з API
init new myapp

# Новий bootstrap-проєкт (сервер + статична генерація)
init new myapp -bootstrap

# Шлях і прапорець у будь-якому порядку
init new -bootstrap myapp

# Поточна директорія, явна назва модуля
init new . -module example.com/myapp
```

Після створення `init` автоматично виконує `go generate ./...` та `go mod tidy` у директорії проєкту, щоб згенерувати код шаблонів і підтягнути залежності.

### init add

Додає шаблони (та за бажанням bootstrap) у **поточну** директорію. Поточна директорія має бути Go-модулем (наявний `go.mod`).

| Прапорець | Опис |
|-----------|------|
| `-bootstrap` | Додати `processor/templates/`, `data/`, `shared/` та приклад main. Якщо `main.go` вже є, створюється `main_hbars_example.go` для вбудовування. |

**Приклади:**

```bash
cd /шлях/до/вашого/модуля
init add                # templates/ + gen.go + приклад .hbs
init add -bootstrap     # processor/templates, data/, shared/, приклад main
```

Існуючі файли (наприклад `gen.go`, `main.go`) не перезаписуються; створюються лише нові. Після додавання файлів `init` виконує `go generate ./...` та `go mod tidy` у директорії модуля.

## Локальний чекаут

При роботі з локальною копією go-hbars використовуйте `replace` у `go.mod` проєкту та запускайте init з репо:

```bash
cd /шлях/до/go-hbars
go run ./cmd/init new /шлях/до/myapp -bootstrap
```

Детальніше: [Робота з локальним чекаутом](howto-integrate-api.md#робота-з-локальним-чекаутом) в гайдах по інтеграції.

## Див. також

- [Як інтегрувати API](howto-integrate-api.md) — ручне налаштування без init (шаблони + go:generate).
- [Як інтегрувати bootstrap](howto-integrate-bootstrap.md) — ручне налаштування bootstrap (QuickServer, QuickProcessor).
- [Процесор та веб-сервер](processor-server.md) — CLI-інструменти та формат файлів даних.
