package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONB: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, j)
}

// JSONRaw is a byte-slice type for JSONB columns that may hold any JSON value
// (objects, arrays, strings, etc.). Using a pointer receiver so that GORM's
// reflect.New(fieldType).Elem() produces a value whose method set includes Scan.
type JSONRaw []byte

func (j *JSONRaw) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONRaw: expected []byte, got %T", value)
	}
	*j = make(JSONRaw, len(bytes))
	copy(*j, bytes)
	return nil
}

func (j JSONRaw) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return []byte(j), nil
}
