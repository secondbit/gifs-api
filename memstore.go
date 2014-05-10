package api

type Memstore map[string]*Collection

func (m Memstore) Init(dbName string) error {
	return nil
}

func (m Memstore) CreateCollection(slug, name string) (Collection, error) {
	m[slug] = &Collection{Name: name,
		Slug:  slug,
		Items: map[string]Item{},
	}
	return *m[slug], nil
}

func (m Memstore) UpdateCollection(slug, name string) error {
	if c, ok := m[slug]; !ok {
		return CollectionNotFoundError
	} else {
		c.Name = name
	}
	return nil
}

func (m Memstore) GetCollectionData(slug string) (Collection, error) {
	if c, ok := m[slug]; ok {
		(*c).Items = nil
		return *c, nil
	}
	return Collection{}, CollectionNotFoundError
}

func (m Memstore) GetCollectionItems(slug string) (map[string]Item, error) {
	if c, ok := m[slug]; ok {
		return c.Items, nil
	}
	return map[string]Item{}, CollectionNotFoundError
}

func (m Memstore) AddItemToCollection(slug string, item Item) error {
	if c, ok := m[slug]; ok {
		c.Items[item.Tag] = item
		return nil
	}
	return CollectionNotFoundError
}

func (m Memstore) GetItemFromCollection(slug, tag string) (Item, error) {
	if _, ok := m[slug]; !ok {
		return Item{}, CollectionNotFoundError
	}
	if _, ok := m[slug].Items[tag]; !ok {
		return Item{}, BlobNotFoundError
	}
	return m[slug].Items[tag], nil
}
