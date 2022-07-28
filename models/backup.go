package models

type backup struct {
	name string
	size int
	createdTime string
}

type snapshot struct {
	backup
}
