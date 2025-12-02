package database

import (
	"errors"
)

var (
	ErrNoRecord = errors.New("No se encontró la fila en la base de datos.")
)
