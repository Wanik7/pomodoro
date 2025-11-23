# Pomodoro TUI

Минималистичное терминальное приложение Pomodoro + список задач. Написано на Go с использованием [Bubble Tea](https://github.com/charmbracelet/bubbletea) и компонента `textinput` из [Bubbles](https://github.com/charmbracelet/bubbles). 
- Циклические интервалы: Работа ↔ Перерыв (классический Pomodoro).
- Таймер в реальном времени (тик каждую секунду).
- Автопереключение режима при окончании интервала.
- Список задач:
    - Добавление новой задачи (поле ввода).
    - Навигация курсором.
    - Переключение статуса Done.
    - Удаление задачи.
- Persist (JSON): задачи и `next_id` сохраняются в `persist.json`.
- Обработка ошибок загрузки/сохранения без падения приложения.

## Требования

- Go 1.20+ (рекомендуется ≥1.21)
- Наличие TTY (запуск из настоящего терминала, не из «сырого» лог-вывода без /dev/tty)

## Установка и запуск

```bash
git clone https://github.com/Wanik7/pomodoro
cd pomodoro
go mod tidy
go run .
```

Для сборки бинарника:

```bash
go build -o pomodoro-tui .
./pomodoro-tui
```

Основные типы:
- `model` — состояние таймера и списка задач.
- `Task` — элемент задач.
- Сообщения: `tickMsg`, `loadMsg`, `persistMsg`.

Если возникли проблемы с TTY:
```go
p := tea.NewProgram(initialModel()) // без tea.WithAltScreen()
```