package models

type Label struct {
	Id         uint64
	CategoryId uint64
	Name       string
}

type Category struct {
	Id   uint64
	Name string
}
