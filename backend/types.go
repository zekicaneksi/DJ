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
	TagsIDs []int64
	Amount  int
}
