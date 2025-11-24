# Минимальная dev-среда для тестирования провайдера

Цель — быстро поднять рабочий стенд 3x-ui локально (Docker), чтобы использовать его для ручного тестирования и Terraform acceptance-тестов.

## 1. Подготовка каталога
Выделите отдельную папку под стенд (например, `~/panel`), чтобы рядом сохранялись база и сертификаты:
```bash
mkdir -p ~/panel
cd ~/panel
```

## 2. Файл `compose.yml`
Создайте `compose.yml` (через `nano`, `vim` или любой редактор) со следующим содержимым:
```yaml
services:
  3xui:
    image: ghcr.io/mhsanaei/3x-ui:latest
    container_name: 3xui_app
    # hostname: yourhostname <- optional
    ports:
      - "2053:2053"   # HTTPS панель
      - "9999:9999"   # HTTP панель (если понадобится)
      - "54321:54321" # XRAY API порт по умолчанию
    volumes:
      - xui-db:/etc/x-ui/
      - xui-cert:/root/cert/
    environment:
      XRAY_VMESS_AEAD_FORCED: "false"
      XUI_ENABLE_FAIL2BAN: "true"
    tty: true
    restart: unless-stopped

volumes:
  xui-db:
  xui-cert:
```
Такой compose-файл сразу использует официальное изображение из `ghcr.io` и прокидывает нужные порты наружу. При необходимости измените порты в секции `ports`.

## 3. Запуск через Docker Compose
Поднимите контейнер командой `task env-up` (внутри репозитория) или вручную:
```bash
docker compose up -d
```
Docker автоматически создаст именованные тома `xui-db` и `xui-cert`, где будет храниться база и TLS-материалы. Для полного сброса выполните `docker compose down -v`.

По умолчанию веб-панель доступна по `https://localhost:2053/panel` (самоподписанный сертификат).

### Настройка окружения
- В секции `environment` уже отключён принудительный VMess AEAD и включён Fail2Ban, при необходимости добавляйте другие переменные окружения.
- Порты проброшены через `ports`. Если нужен нестандартный внешний порт, измените левую часть записи (`<host_port>:<container_port>`).
- Состояние хранится в именованных томах `xui-db` и `xui-cert`; для сброса окружения используйте `docker volume rm` или `docker compose down -v`.

## 3. Доступ и первичная настройка
После запуска откройте `https://localhost:2053/panel` (игнорируя предупреждение TLS).
- Логин/пароль по умолчанию: `admin` / `admin` (попросит сменить при первом входе).
- Настройте двухфакторную аутентификацию при необходимости (для тестов лучше отключить, чтобы не мешала автоматизации).

## 4. Получение cookie для API
Для отладки провайдера можно использовать `curl` или `httpie`:
```bash
curl -k -c cookies.txt -X POST https://localhost:2053/login \
  -d 'username=admin' -d 'password=<новый пароль>'
```
После успешного ответа cookie в `cookies.txt` позволит вызывать `/panel/api/**`.

## 5. Использование в acceptance-тестах
- Экспортируйте переменные окружения, которые будет читать провайдер:
  - `TF_ACC=1` — включает acceptance-тесты.
  - `THREEXUI_BASE_URL`, `THREEXUI_USERNAME`, `THREEXUI_PASSWORD` — подключение к панели.
  - `THREEXUI_TLS_SKIP_VERIFY=true` (опционально) для self-signed стенда.
- Запустите `task acc`, чтобы прогнать сценарий `3xui_inbound` (создание → импорт → удаление). Перед запуском убедитесь, что в панели нет конфликтующих inbound'ов с портом `28000`.

## 6. Smoke-тесты примеров
Чтобы удостовериться, что конфигурации из `examples/` работают целиком через Terraform CLI:

1. Убедитесь, что переменные `THREEXUI_BASE_URL`, `THREEXUI_USERNAME`, `THREEXUI_PASSWORD` заданы.
2. Выполните `task smoke`. Скрипт `scripts/smoke.sh` соберёт провайдер, настроит `dev_overrides` и выполнит `terraform init/apply/destroy` для `examples/resources/3xui_inbound` с использованием локального бинаря.
3. После завершения пример будет удалён автоматически (вызов `terraform destroy` идёт в этом же скрипте).

## 6. Альтернативы
- Можно поднять панель напрямую на хостовой машине (`./x-ui.sh install`), но Docker проще для CI и повторяемости.
- Для быстрой очистки состояния достаточно `task env-down && task env-up` (или вручную `docker compose down -v && docker compose up -d`).

Эти инструкции закрывают пункт 1.4 плана и дают базу для дальнейших автоматизированных тестов.
