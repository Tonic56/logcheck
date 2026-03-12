# logcheck

Go-линтер, совместимый с golangci-lint, который проверяет лог-сообщения на стиль, язык и безопасность.

## Правила

| Правило | ID | Описание |
|---|---|---|
| Строчная буква | `lowercase` | Лог-сообщение должно начинаться со строчной буквы |
| Только английский | `english-only` | Нельзя использовать не-ASCII / не-латинские символы |
| Без спецсимволов | `no-special-chars` | Нельзя использовать эмодзи и шумные спецсимволы |
| Без чувствительных данных | `no-sensitive-data` | Нельзя логировать пароли, токены, ключи и т.д. |

## Примеры

```go
// ❌ Правило 1 — строчная буква
slog.Info("Starting server on port 8080")
// ✅
slog.Info("starting server on port 8080")

// ❌ Правило 2 — только английский
slog.Error("ошибка подключения к базе данных")
// ✅
slog.Error("failed to connect to database")

// ❌ Правило 3 — без спецсимволов / эмодзи
slog.Warn("connection failed!!!")
slog.Info("server started 🚀")
// ✅
slog.Warn("connection failed")
slog.Info("server started")

// ❌ Правило 4 — без чувствительных данных
slog.Info("user password: " + password)
slog.Debug("api_key=" + apiKey)
// ✅
slog.Info("user authenticated successfully")
slog.Debug("api request completed")
```

## Поддерживаемые логгеры

- `log/slog` — стандартная библиотека Go 1.21+
- `go.uber.org/zap` — включая методы SugaredLogger (`Infof`, `Infow` и т.д.)
- `log` — стандартная библиотека

## Установка

### Отдельный бинарный файл

```bash
go install github.com/Tonic56/logcheck/cmd/logcheck@latest
```

Запуск:

```bash
logcheck ./...
# с указанием конфига:
logcheck -config path/to/logcheck.yaml ./...
```

### Плагин для golangci-lint

```bash
go build -buildmode=plugin -o logcheck.so ./plugin
```

Добавьте в `.golangci.yml`:

```yaml
linters-settings:
  custom:
    logcheck:
      path: ./logcheck.so
      description: Проверяет лог-сообщения на стиль и чувствительные данные
      original-url: github.com/Tonic56/logcheck

linters:
  enable:
    - logcheck
```

## Конфигурация

Создайте `logcheck.yaml` в корне проекта:

```yaml
# Отключить отдельные правила (по умолчанию все включены)
disabled:
  - no-sensitive-data

# Дополнительные ключевые слова для правила no-sensitive-data
extra_keywords:
  - internal_secret
  - db_master_pass

# Разрешить определённые спецсимволы в правиле no-special-chars
allowed_chars: ""

# Пользовательские regexp-паттерны для правила no-sensitive-data
custom_patterns:
  - "sk_live_[a-zA-Z0-9]+"
```

### Встроенные чувствительные ключевые слова

`password`, `passwd`, `secret`, `token`, `apikey`, `api_key`, `authkey`, `auth_key`, `credential`, `private_key`, `access_key`, `session`, `jwt`, `bearer`, `ssn`, `credit_card`

## Авто-исправление

Правила `lowercase` и `no-special-chars` поддерживают авто-исправление через механизм `SuggestedFixes`:

- **lowercase** — первая буква строкового литерала приводится к нижнему регистру
- **no-special-chars** — из строкового литерала удаляются эмодзи и шумные спецсимволы

Авто-исправление работает только для простых строковых литералов, не для конкатенаций с переменными.

## Сборка и тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск с детектором гонок
go test -race ./...

# Сборка бинарного файла
go build -o logcheck ./cmd/logcheck

# Сборка плагина
go build -buildmode=plugin -o logcheck.so ./plugin
```

## Структура проекта

```
logcheck/
├── cmd/logcheck/           # CLI-бинарный файл
├── pkg/
│   ├── analyzer/           # Основной go/analysis проход
│   ├── config/             # Загрузка YAML-конфигурации
│   ├── logmatch/           # Определение и разбор вызовов логгеров
│   └── rules/              # Реализации правил
│       ├── rule.go             Интерфейс Rule и типы Diagnostic/Fix
│       ├── lowercase.go        Правило 1: строчная буква
│       ├── english.go          Правило 2: только английский
│       ├── nospecial.go        Правило 3: без спецсимволов
│       └── sensitive.go        Правило 4: без чувствительных данных
├── plugin/                 # Точка входа плагина golangci-lint
├── testdata/src/           # Фикстуры для analysistest
├── .github/workflows/      # GitHub Actions CI
├── logcheck.yaml           # Пример конфигурации
└── .golangci.yml
```