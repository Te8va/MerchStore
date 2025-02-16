Приложение работает на порту 8080. 
Запускается через docker-compose up.
Программа имеет следующие энпоинты.
POST /api/auth - авторизация или логирование пользвателя. После успешного запуска выведется jwt токен.
Пример тела запроса.
{
  "username": "string",
  "password": "string"
}
GET /api/buy/{item} - покупка товара за монеты. JWT токен указывается в заголовке.
POST/api/sendCoin Отправить монеты другому пользователю. JWT токен указывается в заголовке.
Пример тела запроса:
{
  "toUser": "string",
  "amount": 0
}
GET /api/info Получить информацию о монетах, инвентаре и истории транзакций. JWT токен указывается в заголовке.

Сервис покрыт югит тестами. Общее тестовое покрытие проекта превышает 40%
Для проверки общего покрытия:
go test -coverprofile cov ./...
go tool cover -func cov

Код покрыт E2E тестами, тест находится в папке e2e_test, также там находятся сценарное покрытие тестами E2E.
Для запуска контейнера с е2е тестами 
docker-compose --profile e2e up db-e2e

 Нагрузочное тестирование:
 Для нагрузочного тестирования использовал K6. Скрипт на JS для этого находится в корне репозитория, называется load.js. У меняе есть предположение, что не очень хорошо проходит нагрузочное тестирование из-за хэширования пароля.

