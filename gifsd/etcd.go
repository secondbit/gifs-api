package main

import (
	"errors"
	"log"

	"github.com/coreos/go-etcd/etcd"
	"secondbit.org/gifs/api"
)

var (
	NoBucketSetError = errors.New("No bucket set.")
	NoAuthIDSetError = errors.New("No auth id set.")
)

func getEtcdContext(resp *etcd.Node) (api.Context, error) {
	var context api.Context
	var gcsEmail, gcsTokenURI string
	var gcsPemBytes []byte
	var authID string
	var dsn string
	var bucket, domain string
	for _, node := range resp.Nodes {
		switch node.Key {
		case "/gcs":
			gcsEmail, gcsTokenURI, gcsPemBytes = gcsFromNode(node)
		case "/dsn":
			dsn = node.Value
		case "/bucket":
			bucket = node.Value
		case "/google_oauth2_id":
			authID = node.Value
		case "/domain":
			domain = node.Value
		}
	}
	if bucket == "" {
		return context, NoBucketSetError
	}
	if authID == "" {
		return context, NoAuthIDSetError
	}
	if gcsEmail != "" && len(gcsPemBytes) > 0 {
		if gcsTokenURI == "" {
			gcsTokenURI = api.DefaultTokenURI
		}
		log.Println("Using Google Cloud Storage as our storage backend.")
		storage, err := api.NewGCSStorage(gcsEmail, gcsTokenURI, gcsPemBytes)
		if err != nil {
			return context, err
		}
		context.Storage = storage
	} else {
		context.Storage = api.NewMemStorage()
	}
	if dsn != "" {
		log.Println("Using MySQL as our datastore.")
		datastore, err := api.NewMySQLDatastore(dsn)
		if err != nil {
			return context, err
		}
		context.Datastore = datastore
	} else {
		context.Datastore = api.NewMemDatastore()
	}
	context.UsageTracker = api.NewUsageTracker()
	context.Authorizer = api.NewGoogleOAuth2Authorizer(authID)
	context.Bucket = bucket
	context.RootDomain = domain
	return context, nil
}

func gcsFromNode(node *etcd.Node) (email, token string, pem []byte) {
	for _, n := range node.Nodes {
		switch n.Key {
		case "/email":
			email = n.Value
		case "/token":
			token = n.Value
		case "/pem":
			pem = []byte(n.Value)
		}
	}
	return
}
