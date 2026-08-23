package apikey

const (

	// Получение всех API-ключей.
	queryList = `
SELECT
	id,
	api_key,
	owner,
	enabled,
	created_at,
	expires_at
FROM api_keys
ORDER BY id;
`

	// Получение всех активных API-ключей.
	queryListEnabled = `
SELECT
    id,
    api_key,
    owner,
    enabled,
    created_at,
    expires_at
FROM api_keys
WHERE enabled = true
ORDER BY id;
`

	// Проверка существования ключа.
	queryFindByKey = `
SELECT
	id,
	api_key,
	owner,
	enabled,
	created_at,
	expires_at
FROM api_keys
WHERE api_key = $1
LIMIT 1;
`

	// Создание ключа.
	queryCreate = `
INSERT INTO api_keys
(
	api_key,
	owner,
	enabled,
	expires_at
)
VALUES
(
	$1,
	$2,
	$3,
	$4
)
RETURNING id;
`

	// Отключение ключа.
	queryDisable = `
UPDATE api_keys
SET enabled = false
WHERE id = $1;
`

	// Удаление ключа.
	queryDelete = `
DELETE FROM api_keys
WHERE id = $1;
`
)
