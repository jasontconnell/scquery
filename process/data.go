package process

import "github.com/google/uuid"

type Comparison string

const (
	Equals   Comparison = "=="
	Contains Comparison = "[]"
)

type FieldQuery struct {
	Name  string
	Value string
}

type FieldResult struct {
	Name string
}

type ItemFieldResult struct {
	Name  string
	Value string
}

type ItemResult struct {
	Id     uuid.UUID
	Fields []ItemFieldResult
}
