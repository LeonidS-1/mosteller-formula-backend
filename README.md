#### Лабораторная 3

- **Цель работы**: REST API бэкенда для SPA: препараты (`drugs`), рецепты (`prescriptions`), связь м-м `prescription_drugs` (рост, вес, рассчитанная доза `pediatric_dose_mg`), пользователи, загрузка изображения и короткого видео препарата в MinIO.
- **Порядок показа**: коллекция запросов в Insomnia/Postman — `GET` списка рецептов с фильтрами `from-date`, `to-date`, `status`; `GET` корзины `/api/prescriptions/cart`; удаление черновика рецепта; `GET` списка препаратов с фильтром `Title`; `POST` препарата с файлами; добавление препаратов в черновик; просмотр рецепта; `PUT` правка полей м-м (рост/вес) и рецепта (ФИО врача, примечания); формирование рецепта (пересчёт дозы по Mosteller); попытка завершить черновик (ошибка); вход модератора; завершение/отклонение сформированного рецепта; регистрация пользователя. В БД — `SELECT` для проверки данных; в коде — модели, сериализаторы, singleton создателя (`creatorUserID` / `GetCreatorID`).
- **Контрольные вопросы**: веб-сервис, REST, HTTP, HTTPS, версии HTTP, OSI.
- **Диаграмма классов** бэкенда по URL, модели и таблицы — по курсу.

**Требования к API**

- Префикс всех методов: `/api`.
- Создатель рецепта зафиксирован константой в репозитории; модератор — через `POST /api/users/login` (заглушка пароля).
- Список рецептов не отдаёт статусы `deleted` и `draft`; фильтрация по диапазону даты формирования и статусу на бэкенде.
- Список препаратов — фильтр по строке поиска, query-параметр `Title` (как в методичке по аналогии с первой лабораторной).
- В списке рецептов поле `completed_dose_line_count` — число строк м-м с непустым `pediatric_dose_mg`.

**Эндпоинты**

| Метод | Путь |
|--------|------|
| GET | `/api/drugs?Title=...` |
| GET | `/api/drugs/:id` |
| POST | `/api/drugs` |
| GET | `/api/prescriptions/cart` |
| GET | `/api/prescriptions?from-date=&to-date=&status=` |
| GET | `/api/prescriptions/:id` |
| PUT | `/api/prescriptions/:id` |
| PUT | `/api/prescriptions/:id/form` |
| PUT | `/api/prescriptions/:id/finish` |
| DELETE | `/api/prescriptions/:id` |
| POST | `/api/prescription_drugs/add/:drug_id` |
| DELETE | `/api/prescription_drugs/:drug_id/:prescription_id` |
| PUT | `/api/prescription_drugs/:drug_id/:prescription_id` |
| POST | `/api/users/register` |
| POST | `/api/users/login` |
| POST | `/api/users/logout` |

### Запуск

1. `docker compose up -d` (PostgreSQL, MinIO, Redis, Adminer на http://localhost:8081).
2. В MinIO Console создайте бакет с именем из `MINIO_BUCKET` (по умолчанию `inv-media`) или задайте переменную под существующий бакет.
3. Переменные: `DB_*`, при необходимости `MINIO_HOST`, `MINIO_PORT`, `MINIO_USER`, `MINIO_PASS`, `MINIO_BUCKET`.
4. `make migrate-up` — применить `migrations/init-up.sql`.
5. `make run` или `go run ./cmd/app/` из каталога проекта (после `source .env`).

Статические файлы: `/static`. HTML-шаблоны в репозитории не подключаются к Gin (режим JSON API).

### Ссылки

- [Методические указания Golang (lab3)](https://github.com/iu5git/Networking/blob/main/tutorials/lab3-go/README.md) (при необходимости замените на актуальный репозиторий курса).
- [PostgreSQL в Docker](https://github.com/iu5git/Networking/blob/main/tutorials/lab2-db/README.md).
