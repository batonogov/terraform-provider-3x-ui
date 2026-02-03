# Как запустить пример с dev‑сборкой

Этот пример рассчитан на локальный dev‑бинарь провайдера и включённый dev override.

## 1) Собрать бинарь провайдера

Из корня репозитория:

```
task build-dev
```

После этого в корне репо появится `terraform-provider-threexui`.

## 2) Включить dev overrides

Terraform читает `~/.terraformrc`, OpenTofu — `~/.tofurc`.
Пример:

```
provider_installation {
  dev_overrides {
    "hashicorp/threexui" = "/Users/fedor/Documents/github.com/batonogov/terraform-provider-3x-ui"
  }
  direct {}
}
```

## 3) Запуск (ВАЖНО: без init)

С включёнными dev overrides **не делай** `terraform init`/`tofu init`.
Сразу запускай apply:

```
terraform apply
```

или для OpenTofu:

```
tofu apply
```

Terraform/Tofu подхватит локальный бинарь из dev override.
