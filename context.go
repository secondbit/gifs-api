package api

import (
	"database/sql"

	"code.google.com/p/goauth2/oauth/jwt"
	"code.google.com/p/google-api-go-client/storage/v1beta2"
	_ "github.com/go-sql-driver/mysql"
)

type Context struct {
	Storage      Storage
	Datastore    Datastore
	UsageTracker *UsageTracker
	Authorizer   Authorizer
	Bucket       string
	RootDomain   string
}

func NewMemStorage() Storage {
	return make(Memstorage)
}

func NewMemDatastore() Datastore {
	return make(Memstore)
}

func NewGCSStorage(gcsClientEmail, gcsTokenURI string, gcsPemBytes []byte) (Storage, error) {
	t := jwt.NewToken(gcsClientEmail, storage.DevstorageFull_controlScope, gcsPemBytes)
	t.ClaimSet.Aud = gcsTokenURI
	transport, err := jwt.NewTransport(t)
	if err != nil {
		return nil, err
	}
	client := transport.Client()
	gcsService, err := storage.New(client)
	if err != nil {
		return nil, err
	}
	return &GoogleCloudStorage{gcsService}, nil
}

func NewMySQLDatastore(dsn string) (Datastore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return (*SQLStore)(db), nil
}
