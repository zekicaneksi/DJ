package main

type Tag struct {
	ID   int64
	Name string
}

type File struct {
	ID   int64
	Name string
}

type TagGroup struct {
	TagIDs []int64
	Amount int
}
