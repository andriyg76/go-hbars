---
name: HBS Unsupported Patterns Doc
overview: Створити документацію з прикладами Handlebars патернів, які виглядають валідними, але не підтримуються go-hbars через compile-time type inference, та тести що фіксують ці обмеження.
todos:
  - id: create-limitations-doc-uk
    content: Створити docs/uk/limitations.md з прикладами непідтримуваних патернів (українською)
    status: pending
  - id: create-limitations-doc-en
    content: Створити docs/limitations.md з прикладами непідтримуваних патернів (англійською)
    status: pending
  - id: create-limitations-tests
    content: Створити internal/compiler/limitations_test.go з тестами що фіксують поведінку компілятора
    status: pending
  - id: update-readme-uk
    content: Додати посилання на limitations.md в docs/uk/README.md
    status: pending
  - id: update-readme-en
    content: Додати посилання на limitations.md в docs/README.md
    status: pending
isProject: false
---

# Документація непідтримуваних Handlebars патернів

## Контекст проблеми

go-hbars - це compile-time Handlebars компілятор, який генерує type-safe Go код. На відміну від runtime-інтерпретаторів (як Handlebars.js), go-hbars виводить типи контексту з шаблонів під час компіляції. Це означає, що деякі легітимні Handlebars патерни неможливо підтримати через необхідність статичної типізації.

## Категорії непідтримуваних патернів

### 1. Поліморфні partials - різні типи контексту

Partial викликається з різними типами контексту:

```handlebars
{{! card.hbs - partial що очікує різні типи }}
<div>{{title}}</div>
<div>{{description}}</div>

{{! page1.hbs }}
{{> card user}}  {{! user: {title: "John", description: "Developer"} }}

{{! page2.hbs }}  
{{> card product}}  {{! product: {title: "Widget", description: "A thing", price: 99} }}
```

**Проблема:** Компілятор не може згенерувати один інтерфейс `CardContext` що задовольняє обидва виклики - `user` і `product` мають різну структуру.

### 2. Змішане використання шляху - each та with

Один шлях використовується і як колекція, і як об'єкт:

```handlebars
{{! template1.hbs }}
{{#each items}}
  <li>{{name}}</li>
{{/each}}

{{! template2.hbs (або в іншому місці template1) }}
{{#with items}}
  Total: {{count}}
{{/with}}
```

**Проблема:** `items` має бути одночасно `[]ItemContext` (для each) і об'єктом з полем `count` (для with). Це взаємовиключні вимоги.

### 3. Навігація `../` в shared partials

Partial що використовує `../` і викликається з різних глибин:

```handlebars
{{! userBadge.hbs }}
<span>{{name}} - {{../siteName}}</span>

{{! page1.hbs }}
{{#with user}}
  {{> userBadge}}  {{! ../siteName = root.siteName }}
{{/with}}

{{! page2.hbs }}
{{#with company}}
  {{#with admin}}
    {{> userBadge}}  {{! ../siteName = company.siteName, не root! }}
  {{/with}}
{{/with}}
```

**Проблема:** Partial очікує `siteName` в батьківському контексті, але батьківський контекст різний залежно від глибини виклику.

### 4. Динамічні partial імена з різними контекстами

```handlebars
{{> (lookup . "partialName") data}}
```

**Проблема:** Якщо `partialName` може бути "userCard" або "productCard", і кожен partial очікує різний тип `data`, компілятор не може вивести правильний тип.

### 5. Conditional context paths

```handlebars
{{#if isAdmin}}
  {{adminSettings.theme}}
{{else}}
  {{userSettings.theme}}
{{/if}}
```

**Проблема:** Обидва `adminSettings.theme` і `userSettings.theme` додаються до типу контексту, навіть якщо тільки один буде використаний.

### 6. Рекурсивні partials з різною глибиною

```handlebars
{{! tree.hbs }}
<div>{{name}}</div>
{{#each children}}
  {{> tree}}  {{! рекурсивний виклик }}
{{/each}}
```

**Проблема:** Кожен рівень рекурсії створює новий контекстний тип (`TreeContext`, `TreeChildrenItemContext`), що може призвести до нескінченної генерації типів.

### 7. Cross-partial `@root` з різними root контекстами

```handlebars
{{! footer.hbs - shared partial }}
<footer>{{@root.copyright}}</footer>

{{! page1.hbs - root має copyright }}
{{> footer}}

{{! email.hbs - root НЕ має copyright }}
{{> footer}}
```

**Проблема:** `@root.copyright` вимагає щоб всі caller-и мали поле `copyright` в root контексті. Канонічний тип partial не включає `@root` поля.

## Файли для створення

### Документація (дві мовні версії)

1. **[docs/uk/limitations.md](docs/uk/limitations.md)** - українською
2. **[docs/limitations.md](docs/limitations.md)** - англійською

Обидва документи містять:

- Вступ про природу обмежень (compile-time vs runtime)
- Детальні приклади кожного патерну
- Пояснення чому це не працює
- Можливі обхідні шляхи де вони існують

### Тести

3. **[internal/compiler/limitations_test.go](internal/compiler/limitations_test.go)** - тести що фіксують поведінку

## Структура документа

1. **Вступ** - пояснення різниці compile-time vs runtime
2. **Поліморфні partials** - один partial, різні контексти
3. **Змішане each/with** - один шлях як масив і об'єкт
4. **Parent navigation в shared partials** - `../` з різних глибин
5. **Рекурсивні partials** - нескінченна генерація типів
6. **@root в shared partials** - різні root контексти
7. **Динамічні partial імена** - runtime визначення partial

---

## Тести для фіксації обмежень

Файл: [internal/compiler/limitations_test.go](internal/compiler/limitations_test.go)

### Підхід до тестування

Тести повинні фіксувати поточну поведінку компілятора для непідтримуваних патернів:

1. **Помилка компіляції** - `CompileTemplates` повертає error
2. **Код не компілюється** - Go code генерується, але `go build` падає
3. **Panic/timeout** - компілятор зависає або панікує

### Структура тестів

```go
func TestLimitations_PolymorphicPartials(t *testing.T) {
    // Partial викликається з різними типами контексту
    _, err := CompileTemplates(map[string]string{
        "card":  "<div>{{title}}</div>",
        "page1": "{{> card user}}",
        "page2": "{{> card product}}",
    }, Options{PackageName: "templates"})
    // Очікуємо: помилка або код що не компілюється
}

func TestLimitations_MixedEachWith(t *testing.T) {
    // Один шлях як колекція і об'єкт
    code, err := CompileTemplates(map[string]string{
        "main": `{{#each items}}{{name}}{{/each}}{{#with items}}{{count}}{{/with}}`,
    }, Options{PackageName: "templates"})
    // Перевіряємо що код не компілюється
}

func TestLimitations_ParentNavigationInSharedPartial(t *testing.T) {
    // Partial з ../ викликається з різних глибин
    _, err := CompileTemplates(map[string]string{
        "badge": "<span>{{../siteName}}</span>",
        "page1": "{{#with user}}{{> badge}}{{/with}}",
        "page2": "{{#with company}}{{#with admin}}{{> badge}}{{/with}}{{/with}}",
    }, Options{PackageName: "templates"})
    // Очікуємо: різні типи для ../siteName
}

func TestLimitations_RecursivePartials(t *testing.T) {
    // Рекурсивний partial - може зависнути або panic
    done := make(chan struct{})
    go func() {
        defer close(done)
        CompileTemplates(map[string]string{
            "tree": "<div>{{name}}{{#each children}}{{> tree}}{{/each}}</div>",
            "main": "{{> tree}}",
        }, Options{PackageName: "templates"})
    }()
    select {
    case <-done:
        // OK - завершився (можливо з помилкою)
    case <-time.After(5 * time.Second):
        t.Fatal("compiler hung on recursive partial")
    }
}
```

### Категорії тестів

| Патерн | Очікувана поведінка | Тест |

|--------|---------------------|------|

| Поліморфні partials | Помилка або некомпільований код | `TestLimitations_PolymorphicPartials` |

| Змішане each/with | Некомпільований код (конфлікт типів) | `TestLimitations_MixedEachWith` |

| `../` в shared partial | Некомпільований код або помилка | `TestLimitations_ParentNavigation` |

| Рекурсивні partials | Timeout або panic | `TestLimitations_RecursivePartials` |

| `@root` в shared partial | Некомпільований код | `TestLimitations_RootInSharedPartial` |

| Динамічні partial імена | Залежить від реалізації | `TestLimitations_DynamicPartialNames` |