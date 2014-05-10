package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	DefaultTokenURI   = "https://accounts.google.com/o/oauth2/token"
	DefaultListenAddr = ":8080"
	AuthHeader        = "Gifs-Username"
)

type Handler func(w http.ResponseWriter, r *http.Request, c Context)

func GetDomainMuxer(c Context) *mux.Router {
	r := mux.NewRouter()
	domainSuffix := c.RootDomain
	if domainSuffix[0] == '.' {
		domainSuffix = domainSuffix[1:]
	}
	r.HandleFunc("/", timeHandler(wrap(c, authWrapper(UploadHandler)))).Methods("POST").Host("{collection}." + domainSuffix)
	r.HandleFunc("/", timeHandler(wrap(c, authWrapper(CreateCollection)))).Methods("POST").Host(domainSuffix)
	r.Handle("/", timeHandler(wrap(c, CollectionList))).Methods("GET").Host("{collection}." + domainSuffix)
	r.Handle("/{id}", timeHandler(wrap(c, GetBlob))).Methods("GET").Host("{collection}." + domainSuffix)
	return r
}

func GetPathMuxer(c Context) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", timeHandler(wrap(c, authWrapper(CreateCollection)))).Methods("POST")
	r.HandleFunc("/{collection}", timeHandler(wrap(c, UploadHandler))).Methods("POST")
	r.Handle("/{collection}", timeHandler(wrap(c, CollectionList))).Methods("GET")
	r.Handle("/{collection}/{id}", timeHandler(wrap(c, GetBlob))).Methods("GET")
	return r
}

func authWrapper(f Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request, c Context) {
		r.Header.Del(AuthHeader)
		bearer := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if bearer == "" {
			return
		}
		if c.Authorizer == nil {
			return
		}
		user, err := c.Authorizer.Authorize(bearer, c)
		if err != nil && err != InvalidBearerToken {
			log.Println("Error authorizing request: " + err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		} else if err == InvalidBearerToken {
			http.Error(w, "Invalid authorization", http.StatusUnauthorized)
			return
		}
		r.Header.Set(AuthHeader, user)
		f(w, r, c)
	}
}

func wrap(c Context, f func(w http.ResponseWriter, r *http.Request, c Context)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request received.")
		f(w, r, c)
	})
}

func timer(tag string, start time.Time) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", tag, elapsed)
}

func timeHandler(f http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer timer("[timeHandler] "+r.Method+" request to "+r.URL.Path, time.Now())
		f(w, r)
	})
}

func UploadHandler(w http.ResponseWriter, r *http.Request, c Context) {
	user := r.Header.Get(AuthHeader)
	if user == "" {
		http.Error(w, "Must be logged in", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	collection := vars["collection"]
	if collection == "" {
		http.Error(w, "collection doesn't exist", http.StatusNotFound)
		return
	}
	reader, err := r.MultipartReader()
	if err != nil {
		log.Println("Error creating multipart reader: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	ids := []string{}
	for {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			log.Println("Error looping through reader parts: " + err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		name := part.FileName()
		id, err := Upload(user, collection, name, part, c)
		if err != nil {
			log.Println("Error uploading file: " + err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		ids = append(ids, id)
	}
	encoder := json.NewEncoder(w)
	err = encoder.Encode(ids)
	if err != nil {
		log.Println("Error encoding response: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func CollectionList(w http.ResponseWriter, r *http.Request, c Context) {
	vars := mux.Vars(r)
	collectionSlug := vars["collection"]
	if collectionSlug == "" {
		http.Error(w, "collection doesn't exist", http.StatusNotFound)
		return
	}
	collection, err := c.Datastore.GetCollectionItems(collectionSlug)
	if err != nil {
		if err == CollectionNotFoundError {
			http.Error(w, "collection doesn't exist", http.StatusNotFound)
			return
		}
		log.Println("Error retrieving collection items: " + err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	encoder := json.NewEncoder(w)
	err = encoder.Encode(collection)
	if err != nil {
		log.Println("Error encoding response: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetBlob(w http.ResponseWriter, r *http.Request, c Context) {
	vars := mux.Vars(r)
	collection := vars["collection"]
	if collection == "" {
		http.Error(w, "collection doesn't exist", http.StatusNotFound)
		return
	}
	id := vars["id"]
	if id == "" {
		http.Error(w, "id doesn't exist", http.StatusNotFound)
		return
	}
	item, err := c.Datastore.GetItemFromCollection(collection, id)
	if err != nil {
		log.Println("Error getting item: " + err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_, err = c.Storage.Download(item.Bucket, item.Blob, w, c)
	if err != nil {
		log.Println("Error downloading from GCS: " + err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func CreateCollection(w http.ResponseWriter, r *http.Request, c Context) {
	var collection Collection
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&collection)
	if err != nil {
		log.Println("Error decoding request: " + err.Error())
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	collection, err = c.Datastore.CreateCollection(collection.Slug, collection.Name)
	if err != nil {
		log.Println("Error creating collection: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	encoder := json.NewEncoder(w)
	err = encoder.Encode(collection)
	if err != nil {
		log.Println("Error encoding response: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
