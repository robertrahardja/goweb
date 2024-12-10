package models

type Record struct {
	ID    int32
	Name  string
	Value float32
}

type User struct {
	Email    string
	Name     string
	Picture  string
	IsActive bool
}
