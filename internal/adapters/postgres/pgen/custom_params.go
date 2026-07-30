package pgen

// BatchUpsertMetric это кастомный тип параметра для batch upsert операций
// Это не сгенерированный sqlc-код
type BatchUpsertMetric struct {
	Name  string   `json:"name"`
	Type  string   `json:"type"`
	Value *float64 `json:"value"`
	Delta *int64   `json:"delta"`
}

/*
 * BatchUpsertParam это тип который нужно преобразовать в `[]byte` для корректного использования в `BatchUpsert`
 * Создайте и заполните массив `BatchUpsertParam`, затем преобразуйте его в `[]byte` и передайте в `BatchUpsert`
 */
type BatchUpsertParam = []BatchUpsertMetric
