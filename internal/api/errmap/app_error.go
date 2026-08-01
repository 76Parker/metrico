package errmap

// Коды ошибок нужны для того что бы представлять ошибку в машиночитаемом виде (в основном - для Frontend)
const (
	// Ресурс не найден
	CodeNotFound = "not_found"

	// Внутренняя/неизвестная ошибка
	CodeInternal = "internal"

	// Недопустимый тип метрики
	CodeInvalidMetricType = "invalid_metric_type"

	// Пытаемся добавить 2 метрики с одинаковым именем но разными типами
	CodeTypeConflict = "metrics_type_conflict"

	// Недопустимое значение для метрики (не соответствует ожидаемому типу)
	CodeInvalidMetricValue = "invalid_metric_value"

	// Пришло пустое значение для какого-либо типа
	CodeEmpty = "empty"

	// Некорректное тело HTTP-запроса
	CodeInvalidJSON = "invalid_json"

	// Некорректные параметры HTTP-запроса
	CodeInvalidRequest = "invalid_request"
)
