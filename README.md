# Учёт расходов на Go

## Этап 1 — простой HTTP-сервер

```bash
cd zadanie1
go run .
```

Маршруты:
- `GET /` — приветствие (неизвестные пути → 404)
- `GET /about` — описание проекта
- `GET /ping` — `pong`; другие методы → `405 Method Not Allowed`

## Этап 2 — веб-интерфейс учёта трат

```bash
cd zadanie2
go run ./cmd/server
```

Открыть http://localhost:8080/expenses

Маршруты:
- `GET /expenses` — список трат (шаблон `list.html`, вывод через `range`)
- `GET /expenses/new` — форма добавления (шаблон `form.html`)
- `POST /expenses` — добавление траты в срез в памяти и redirect на `/expenses`
- `/`, `/about`, `/ping` — маршруты первого этапа

Данные хранятся в памяти и сбрасываются при перезапуске сервера.
Запускать нужно из папки `zadanie2`, так как пути к шаблонам относительные.
