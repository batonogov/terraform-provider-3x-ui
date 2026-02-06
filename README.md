# Terraform/OpenTofu провайдер для 3x-ui

Провайдер для управления инбаундами и клиентами 3x-ui через HTTP API панели.

## Конфигурация провайдера

```hcl
provider "threexui" {
  endpoint            = "http://localhost:2053"
  username            = "admin"
  password            = "admin"
  # base_path           = "/"           # опционально
  # insecure_skip_verify = true          # для self-signed HTTPS
  # request_timeout      = "30s"
}
```

## Ресурсы

### `threexui_inbound`

```hcl
resource "threexui_inbound" "example" {
  remark   = "Example Inbound"
  port     = 8443
  protocol = "vless"

  # Опциональные настройки (пример для VLESS)
  # settings {
  #   decryption = "none"
  #   encryption = "none"
  # }
}
```

Основные поля:
- `remark` — описание инбаунда.
- `port` — порт прослушивания.
- `protocol` — протокол (`vless`, `vmess`, `trojan`, `shadowsocks`, ...).
- `settings` — JSON‑настройки инбаунда (без клиентов).

Полезно знать:
- `settings` инбаунда больше не управляет клиентами.
- Клиенты создаются только через `threexui_inbound_client`.

### `threexui_inbound_client`

```hcl
resource "threexui_inbound_client" "client_a" {
  inbound_id = threexui_inbound.example.id
  email      = "client-a@example.com"
  enable     = true
  flow       = "xtls-rprx-vision"
}
```

Основные поля:
- `inbound_id` — ID инбаунда.
- `email` — идентификатор клиента.
- `enable` — включён ли клиент.
- `flow` — flow для VLESS (`xtls-rprx-vision` и т.д.).
- `expiry_time` — время истечения в миллисекундах Unix epoch (число).
- `limit_ip` — лимит IP.
- `total_gb` — лимит трафика.
- `security` / `password` — используются для некоторых протоколов (чувствительные).

## Outputs (пример)

```hcl
output "inbound_clients" {
  value = {
    client_a = {
      id          = threexui_inbound_client.client_a.id
      client_id   = threexui_inbound_client.client_a.client_id
      email       = threexui_inbound_client.client_a.email
      enable      = threexui_inbound_client.client_a.enable
      flow        = threexui_inbound_client.client_a.flow
      limit_ip    = threexui_inbound_client.client_a.limit_ip
      total_gb    = threexui_inbound_client.client_a.total_gb
      expiry_time = threexui_inbound_client.client_a.expiry_time
      tg_id       = threexui_inbound_client.client_a.tg_id
      sub_id      = threexui_inbound_client.client_a.sub_id
      comment     = threexui_inbound_client.client_a.comment
      reset       = threexui_inbound_client.client_a.reset
      security    = threexui_inbound_client.client_a.security
    }
  }
}
```

## Импорт

```bash
# inbound
terraform import threexui_inbound.example 123

# inbound client: <inbound_id>:<client_id>
terraform import threexui_inbound_client.client_a 123:client-id
```

## Разработка

### Требования

- Go 1.21+
- [Task](https://taskfile.dev/) - task runner
- [golangci-lint](https://golangci-lint.run/welcome/install/) - линтер
- [pre-commit](https://pre-commit.com/) - git hooks фреймворк
- Docker - для локального окружения 3x-ui

### Установка pre-commit hooks

```bash
# Установить pre-commit (если ещё не установлен)
pip install pre-commit
# или через brew на macOS
brew install pre-commit

# Установить git hooks
pre-commit install

# Запустить проверки вручную на всех файлах
pre-commit run --all-files
```

### Команды для разработки

```bash
task build        # Собрать провайдер
task fmt          # Форматировать код (gofmt)
task vet          # Запустить go vet
task lint         # Запустить golangci-lint
task pre-commit   # Запустить все проверки вручную (fmt, vet, lint, build)
task test         # Запустить acceptance-тесты (запускает docker compose)
```

### Pre-commit проверки

При каждом коммите автоматически запускаются:
- **gofmt** - форматирование кода
- **go vet** - статический анализ
- **go build** - проверка компиляции
- Проверки YAML/JSON файлов
- Проверка trailing whitespace

Если проверки не проходят, коммит будет отклонён. Исправьте ошибки и попробуйте снова.

**Важно:** `golangci-lint` не запускается автоматически при коммите (он медленный), но рекомендуется запускать вручную перед PR:
```bash
task lint
```

### Локальное окружение

```bash
# Запустить 3x-ui v2.8.9 на localhost:2053
docker compose up -d

# Логин: admin / admin
# webBasePath по умолчанию: /panel/

# Остановить
docker compose down
```
