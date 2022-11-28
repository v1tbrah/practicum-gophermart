# practicum-gophermart

## Installation

`git clone https://github.com/v1tbrah/practicum-gophermart`

## Getting started

### Prepairing

* You need PostgreSQL installed. You can check it: `psql --version`. If it's not installed, visit https://www.postgresql.org/download/
* You need PostgreSQL database "gophermart". If there isn't, do the following:
    ```
    sudo -u postgres psql
    CREATE DATABASE gophermart;
    ```
### Options
The following options are set by default:
```
api server run address: `:8081`
accrual system run address: `:8080`
log level: `info`
order status update interval: `5s`
```
* flag options:
```
   -a string
      api server run address
   -d string
      database connection string
   -r string
      api accrual run address
   -u duration
      order status update interval
   -l string
      log level 
```
For example: `go run cmd/gophermart/main.go -d="host=localhost port=5432 user=postgres password=12345678 dbname=gophermart sslmode=disable"`
* env options can check in internal/parse
      
### Note!

* You definitely need to configure the db connection string

### Starting

* Open first terminal. Go to work project directory. Run binary "accrual loyalty" system. For example:
   ```
   cd ~/go/src/practicum-gophermart
   ./cmd/accrual/accrual_linux_amd64
   ```
* Open second terminal. Go to project working directory. Run "gophermart" app. For example:
   ```
   cd ~/go/src/practicum-gophermart
   go run cmd/gophermart/main.go -d="host=localhost port=5432 user=postgres password=12345678 dbname=gophermart sslmode=disable"
   ```

## Обновление шаблона тестов

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

Затем добавьте полученные изменения в свой репозиторий.
