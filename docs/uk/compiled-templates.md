# Скомпільований файл шаблонів (деталі реалізації)

Цей документ описує **згенерований Go-файл**, який створює `hbc` з шаблонів `.hbs`: як імена файлів шаблонів відповідають Go-символам і що саме випромінює компілятор.

## Ім'я шаблону → Go-ідентифікатор

Компілятор використовує **ім'я шаблону** = ім'я файлу без `.hbs` (наприклад `main.hbs` → `main`, `blog/post.hbs` → `blog/post`). З цього імені він отримує **Go-ідентифікатор** для імен функцій:

1. Розбити ім'я по будь-якому символу, що **не** є літерою або цифрою (`/`, `_`, `-`, пробіл тощо).
2. Зробити першу літеру кожної частини великою.
3. Склеїти частини.
4. Якщо результат порожній або починається з цифри — додати префікс `"Template"`.

| Файл шаблону   | Ім'я шаблону    | Go-ідентифікатор |
|----------------|-----------------|------------------|
| `main.hbs`     | `main`          | `Main`           |
| `header.hbs`   | `header`         | `Header`         |
| `blog/post.hbs`| `blog/post`     | `BlogPost`       |
| `compat_footer.hbs` | `compat_footer` | `CompatFooter`   |
| `404.hbs`      | `404`           | `Template404`    |
| `userCard.hbs` | `userCard`      | `UserCard`       |

Два різні імена шаблонів, що зводяться до одного Go-ідентифікатора (наприклад `blog-post` і `blog_post` → `BlogPost`), призводять до помилки компіляції: компілятор повідомляє про конфлікт.

## Згенерований API на шаблон

Для кожного імені шаблону згенерований пакет надає:

| Go-символ     | Сигнатура | Опис |
|---------------|-----------|------|
| `renderXxx`  | `func(ctx *runtime.Context, w io.Writer) error` | Внутрішня: використовується партіалами та `RenderXxx`. Не призначена для прямого виклику. |
| `RenderXxx`   | `func(w io.Writer, data any) error` | Рендерить шаблон з `data` у `w`. |
| `RenderXxxString` | `func(data any) (string, error)` | Рендерить шаблон з `data` і повертає результат як рядок. |

Приклад для `main.hbs` (Go-ім'я `Main`):

```go
func renderMain(ctx *runtime.Context, w io.Writer) error { ... }
func RenderMain(w io.Writer, data any) error { ... }
func RenderMainString(data any) (string, error) { ... }
```

## Структура згенерованого файлу

1. **Пакет та імпорти**  
   Ім'я пакету (з `-pkg`), імпорти для `io`, `strings`, `runtime` та пакетів хелперів.

2. **Контекстні інтерфейси** (опційно)  
   Типобезпечні аксесори для шляхів контексту шаблону (виводяться з виразів у шаблоні). Використовуються рантаймом; імена похідні від Go-ідентифікатора шаблону та шляху (наприклад `MainContextUser`, `MainContextItems`).

3. **Мапа partials**  
   ```go
   var partials map[string]func(*runtime.Context, io.Writer) error
   func init() {
       partials = map[string]func(*runtime.Context, io.Writer) error{
           "main":   renderMain,
           "header": renderHeader,
           ...
       }
   }
   ```  
   Ключі — імена шаблонів (як у файлах без `.hbs`). Використовується внутрішньо, коли шаблон містить `{{> partialName }}`.

4. **Функції**  
   Для кожного шаблону: `renderXxx`, `RenderXxx`, `RenderXxxString` як вище.

5. **Блок bootstrap** (лише з `-bootstrap`)  
   Див. [Згенерований bootstrap](bootstrap-generated.md).

## Підсумок

- **Ім'я шаблону** = ім'я файлу без `.hbs`.
- **Go-ім'я** = розбити по не-буквоцифровим, з великої літери кожну частину, склеїти; якщо порожньо або починається з цифри — префікс `Template`.
- **Публічний API**: `RenderXxx(w, data)` та `RenderXxxString(data)`; внутрішні `renderXxx` та `partials` — для компілятора/рантайму.
