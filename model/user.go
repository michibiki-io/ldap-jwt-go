package model

type User struct {
	DN     string
	Id     string
	Groups []string
}

type StoreType int8

const (
	StoreTypeAccess StoreType = iota
	StoreTypeRefresh
)

type StoredAuth struct {
	Type       StoreType
	UserId     string
	LinkedUuid string
}
