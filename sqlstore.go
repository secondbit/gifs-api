package api

import (
	"database/sql"

	"github.com/secondbit/pan"
)

const (
	collectionTable = "collections"
	itemTable       = "items"
)

type SQLStore sql.DB

func createCollectionTableSQL() *pan.Query {
	query := pan.New(pan.MYSQL, "CREATE TABLE IF NOT EXISTS "+collectionTable)
	query.Include("(slug VARCHAR(32), name VARCHAR(64))")
	return query.FlushExpressions(" ")
}

func createItemTableSQL() *pan.Query {
	query := pan.New(pan.MYSQL, "CREATE TABLE IF NOT EXISTS "+itemTable)
	query.Include("(tag VARCHAR(32), collection VARCHAR(32), sha VARCHAR(64), bucket VARCHAR(64))")
	return query.FlushExpressions(" ")
}

func (s *SQLStore) Init(name string) error {
	tableInits := []*pan.Query{createCollectionTableSQL(), createItemTableSQL()}
	for _, query := range tableInits {
		_, err := (*sql.DB)(s).Exec(query.String(), query.Args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func createCollectionSQL(slug, name string) *pan.Query {
	query := pan.New(pan.MYSQL, "INSERT INTO "+collectionTable+" (slug, name)")
	query.Include("VALUES (?,?)", slug, name)
	return query.FlushExpressions(" ")
}

func (s *SQLStore) CreateCollection(slug, name string) (Collection, error) {
	query := createCollectionSQL(slug, name)
	_, err := (*sql.DB)(s).Exec(query.String(), query.Args...)
	if err != nil {
		return Collection{}, err
	}
	return Collection{Slug: slug, Name: name, Items: make(map[string]Item)}, nil
}

func (s *SQLStore) UpdateCollection(slug, name string) error {
	return nil
}

func (s *SQLStore) GetCollectionData(slug string) (Collection, error) {
	return Collection{}, nil
}

func (s *SQLStore) GetCollectionItems(slug string) (map[string]Item, error) {
	return map[string]Item{}, nil
}

func addItemToCollectionSQL(slug string, item Item) *pan.Query {
	query := pan.New(pan.MYSQL, "INSERT INTO "+itemTable+" (tag, collection, sha, bucket)")
	query.Include("VALUES (?,?,?,?)", item.Tag, slug, item.Blob, item.Bucket)
	return query.FlushExpressions(" ")
}

func (s *SQLStore) AddItemToCollection(slug string, item Item) error {
	query := addItemToCollectionSQL(slug, item)
	_, err := (*sql.DB)(s).Exec(query.String(), query.Args...)
	return err
}

func getItemFromCollectionSQL(slug, tag string) *pan.Query {
	query := pan.New(pan.MYSQL, "SELECT tag, collection, sha, bucket FROM "+itemTable)
	query.IncludeWhere()
	query.Include("collection=?", slug)
	query.Include("tag=?", tag)
	return query.FlushExpressions(" AND ")
}

func (s *SQLStore) GetItemFromCollection(slug, tag string) (Item, error) {
	query := getItemFromCollectionSQL(slug, tag)
	var i Item
	var collection string
	err := (*sql.DB)(s).QueryRow(query.String(), query.Args...).Scan(&i.Tag, &collection, &i.Blob, &i.Bucket)
	return i, err
}
