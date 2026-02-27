# Вбудовані хелпери

go-hbars включає набір хелперів, узгоджений із Handlebars.js core та handlebars-helpers 7.4. **Базові хелпери підключаються за замовчуванням** — вказувати їх не потрібно, якщо не хочете перевизначити або вимкнути.

## Використання хелперів

**Використання базових хелперів за замовчуванням (найпростіше):**
```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates
// Усі базові хелпери доступні автоматично
```

**Вибір окремих базових хелперів:**
```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -no-core-helpers \
//  -import github.com/andriyg76/go-hbars/helpers/handlebars \
//  -helpers Upper,Lower,FormatDate
```

**Вимкнення базових хелперів і використання власних:**
```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -no-core-helpers \
//  -import github.com/you/custom-helpers \
//  -helpers MyHelper,AnotherHelper
```

**Простий хелпер (локальна функція):**
```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates -helper upper=Upper

func Upper(args runtime.HelperArgs) (any, error) {
	if len(args.Args) == 0 {
		return "", nil
	}
	return strings.ToUpper(runtime.Stringify(args.Args[0])), nil
}
```

**Рекомендований скорочений синтаксис:**
```go
// Імпорт пакету та реєстрація кількох хелперів
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -import github.com/andriyg76/go-hbars/helpers/handlebars \
//  -helpers Upper,Lower,FormatDate

// З аліасами імпортів
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -import github.com/andriyg76/go-hbars/helpers/handlebars \
//  -import extra:github.com/you/extra-helpers \
//  -helpers Upper,Lower \
//  -helpers extra:CustomHelper,extra:AnotherHelper

// Перевизначення імен хелперів
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -import github.com/andriyg76/go-hbars/helpers/handlebars \
//  -helpers myUpper=Upper,myLower=Lower
```

**Застарілий синтаксис (ще підтримується):**
```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates -helper upper=github.com/you/helpers:Upper

// Кілька хелперів
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -helper upper=Upper -helper lower=github.com/you/helpers:Lower
```

**Програмний доступ (для складних випадків):**
```go
import (
	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/internal/compiler"
)

registry := helpers.Registry()
compilerHelpers := make(map[string]compiler.HelperRef)
for name, ref := range registry {
	compilerHelpers[name] = compiler.HelperRef{
		ImportPath: ref.ImportPath,
		Ident:      ref.Ident,
	}
}
opts := compiler.Options{
	PackageName: "templates",
	Helpers:     compilerHelpers,
}
```

## Доступні хелпери

### Рядкові

- `upper`, `lower` — зміна регістру
- `capitalize`, `capitalizeAll` — велика літера в словах
- `truncate` — обрізання рядків з опційним суфіксом
- `reverse` — реверс рядка
- `replace` — заміна підрядків
- `stripTags`, `stripQuotes` — видалення HTML-тегів або лапок
- `join`, `split` — об’єднання/розбиття масивів з роздільником

### Порівняння

- `eq`, `ne` — перевірки рівності
- `lt`, `lte`, `gt`, `gte` — числові порівняння
- `and`, `or`, `not` — логічні операції

### Дати

- `formatDate` — форматування дат (формат Go time)
- `now` — поточний час
- `ago` — «скільки часу тому» у людському вигляді

### Колекції

- `lookup` — пошук значення за ключем
- `default` — значення за замовчуванням для порожніх
- `length` — довжина рядків/масивів/об’єктів
- `first`, `last` — перший/останній елемент масиву
- `inArray` — перевірка наявності значення в масиві

### Математика

- `add`, `subtract`, `multiply`, `divide`, `modulo` — арифметика
- `floor`, `ceil`, `round`, `abs` — округлення та модуль
- `min`, `max` — мінімум/максимум двох чисел

### Числа

- `formatNumber` — форматування з точністю та роздільником
- `toInt`, `toFloat`, `toNumber` — перетворення типів
- `toFixed` — фіксована кількість знаків після коми
- `toString` — перетворення в рядок

### Об’єкти

- `has` — перевірка наявності властивості
- `keys`, `values` — ключі та значення об’єкта
- `size` — розмір об’єкта/масиву
- `isEmpty`, `isNotEmpty` — перевірки на порожність

### URL

- `encodeURI`, `decodeURI` — кодування/декодування URI
- `stripProtocol`, `stripQuerystring` — маніпуляції з URL

## Власні хелпери

Власні хелпери можна реалізувати як звичайні Go-функції та зіставити їх через `-helper name=Ident`. Сигнатура:

```go
func MyHelper(args runtime.HelperArgs) (any, error)
```

Аргументи обчислюються компілятором перед передачею; ви отримуєте `runtime.HelperArgs` з полями `Args` (позиційні) та `HashArgs` (іменовані). Для блокових викликів викликайте `args.BlockFn()` та `args.InverseFn()` для рендеру контенту блоку (writer у замиканні) (вони повертають `error`).

```go
func FormatCurrency(args runtime.HelperArgs) (any, error) {
	if len(args.Args) == 0 {
		return "", nil
	}
	amount := runtime.Stringify(args.Args[0])
	symbol := "$"
	if args.HashArgs != nil {
		if s, ok := args.HashArgs["symbol"].(string); ok {
			symbol = s
		}
	}
	return fmt.Sprintf("%s%s", symbol, amount), nil
}
```

### Блокові хелпери

Блокові хелпери мають ту саму сигнатуру `func(args runtime.HelperArgs) (any, error)`. У блоці `args.IsBlock` дорівнює true і `args.Writer` — writer виводу шаблону. Викликайте `args.BlockFn()` або `args.InverseFn()` без аргументів; кожен є замиканням з уже захопленим writer. Повертають `error`. Повернене значення хелпера для блокових викликів ігнорується.

```go
func MyBlockHelper(args runtime.HelperArgs) (any, error) {
	if !args.IsBlock {
		return nil, nil
	}
	if args.BlockFn != nil {
		if err := args.BlockFn(); err != nil {
			return nil, err
		}
	}
	if args.InverseFn != nil {
		if err := args.InverseFn(); err != nil {
			return nil, err
		}
	}
	return nil, nil
}
```

Приклад умовного блокового хелпера (див. [API шаблонів](api.md)):

```go
func IfHelper(args runtime.HelperArgs) (any, error) {
	if !args.IsBlock || len(args.Args) == 0 {
		return nil, fmt.Errorf("if потребує умову та блок")
	}
	condition := args.Args[0]
	if ok, _ := runtime.IsTruthy(condition); ok {
		if args.BlockFn != nil {
			return nil, args.BlockFn()
		}
	} else if args.InverseFn != nil {
		return nil, args.InverseFn()
	}
	return nil, nil
}
```
