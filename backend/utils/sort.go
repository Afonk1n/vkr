package utils

import "strings"

// SafeOrderClause builds a safe "column direction" ORDER BY clause.
// Защита от SQL-инъекции: и колонка, и направление берутся только из
// белого списка, всё остальное откатывается к значениям по умолчанию.
//
// allowed — множество разрешённых колонок (ключ = имя из query, значение =
// реальное выражение для ORDER BY, что позволяет при желании квалифицировать
// имя таблицы). defaultCol должен присутствовать в allowed.
func SafeOrderClause(sortBy, sortOrder string, allowed map[string]string, defaultCol string) string {
	column, ok := allowed[strings.ToLower(strings.TrimSpace(sortBy))]
	if !ok {
		column = allowed[defaultCol]
	}

	direction := "DESC"
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		direction = "ASC"
	}

	return column + " " + direction
}
