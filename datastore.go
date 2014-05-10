package api

import (
	"errors"
)

var (
	CollectionNotFoundError = errors.New("collection not found")
)

type Datastore interface {
	Init(dbName string) error
	CreateCollection(slug, name string) (Collection, error)
	UpdateCollection(slug, name string) error
	GetCollectionData(slug string) (Collection, error)
	GetCollectionItems(slug string) (map[string]Item, error)
	AddItemToCollection(slug string, item Item) error
	GetItemFromCollection(slug, tag string) (Item, error)
}

type Collection struct {
	Items map[string]Item
	Name  string
	Slug  string
}

type Item struct {
	Blob   string
	Bucket string
	Tag    string
}
