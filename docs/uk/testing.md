# Тестування

## Юніт-тести

Запуск усіх тестів:

```bash
go test ./...
```

Прапорець `-short` пропускає довгі тести (зокрема E2E).

## E2E тести

End-to-end тести знаходяться в `internal/compiler/e2e/`. Вони:

1. Компілюють Handlebars-шаблони компілятором проєкту
2. Записують згенерований Go-код і невеликий драйвер у тимчасовий модуль
3. Запускають `go run` і перевіряють вивід

Ці тести пропускаються при передачі `-short`:

```bash
go test ./... -short                          # без E2E
go test ./internal/compiler/e2e/... -v -count=1   # лише E2E
```

### Список E2E тестів

| Тест | Опис |
|------|------|
| `TestE2E_Compat_IteratorGenerated` | Компілює compat-шаблони; перевіряє згенерований код ітератора |
| `TestE2E_CompatTemplates` | Компілює compat, запускає згенерований код з `data.json`, порівнює з `expected.txt` |
| `TestE2E_IncludeZero` | `{{#if count includeZero=true}}` при `count=0` дає "zero" |
| `TestE2E_Showcase_NilContext` | Showcase з nil/порожнім контекстом; без паніки; помилка динамічного парціалу у виводі |
| `TestE2E_UniversalSection` | Блок-хелпер `date` та умова; перевірка виводу |
| `TestE2E_UserProject_Bootstrap_ServerAndProcessor` | Користувацький проєкт з `-bootstrap`, go generate, `NewQuickProcessor()` |
| `TestE2E_UserProject_GoGenerate_CompatShowcase` | Користувацький проєкт з go:generate (без bootstrap); compat + showcase, RenderCompatString / RenderShowcaseString з `XxxContextFromMap` |

### Контекст і FromMap

Згенеровані шаблони очікують тип контексту (наприклад `MainContext`, `CompatContext`). Якщо дані у вигляді `map[string]any` (наприклад з JSON), використовуйте згенерований `XxxContextFromMap`:

```go
out, err := templates.RenderMainString(templates.MainContextFromMap(data))
```

E2E тести, що передають JSON-дані, використовують цей підхід.
