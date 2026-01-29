# Документація go-hbars (українською)

## Інтеграція

- [init: створити проєкт або додати до нього](init.md) — scaffold нового проєкту або додати шаблони до існуючого модуля
- [Як інтегрувати API](howto-integrate-api.md) — шаблони + go:generate
- [Як інтегрувати bootstrap](howto-integrate-bootstrap.md) — QuickServer + QuickProcessor

## Деталі реалізації

- [Скомпільований файл шаблонів](compiled-templates.md) — згенерований код, імена файлів → функції
- [Згенерований bootstrap](bootstrap-generated.md) — інтерфейс для розробника

## Довідники

- [API шаблонів](api.md) — рантайм API, контекст, хелпери, партіали
- [Вбудований процесор та сервер](embedded.md) — QuickProcessor, QuickServer, sitegen API
- [Процесор та веб-сервер](processor-server.md) — CLI build/server, формат даних, спільні дані
- [Вбудовані хелпери](helpers.md) — рядкові, порівняння, дати, колекції, власні хелпери
- [Синтаксис Handlebars](syntax.md) — вирази, партіали, блоки, шляхи, істинність
- [Власні розширення](extensions.md) — includeZero, універсальна секція
