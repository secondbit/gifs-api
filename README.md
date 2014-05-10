
# api
    import "secondbit.org/gifs/api"




## Constants
``` go
const (
    DefaultTokenURI   = "https://accounts.google.com/o/oauth2/token"
    DefaultListenAddr = ":8080"
    AuthHeader        = "Gifs-Username"
)
```

## Variables
``` go
var (
    BucketNotFoundError = errors.New("bucket not found")
    BlobNotFoundError   = errors.New("blob not found")
)
```
``` go
var (
    CollectionNotFoundError = errors.New("collection not found")
)
```
``` go
var (
    InvalidBearerToken = errors.New("Invalid bearer token")
)
```

## func CollectionList
``` go
func CollectionList(w http.ResponseWriter, r *http.Request, c Context)
```

## func CreateCollection
``` go
func CreateCollection(w http.ResponseWriter, r *http.Request, c Context)
```

## func GetBlob
``` go
func GetBlob(w http.ResponseWriter, r *http.Request, c Context)
```

## func GetDomainMuxer
``` go
func GetDomainMuxer(c Context) *mux.Router
```

## func GetPathMuxer
``` go
func GetPathMuxer(c Context) *mux.Router
```

## func Upload
``` go
func Upload(id, collection, tag string, r io.Reader, c Context) (string, error)
```

## func UploadHandler
``` go
func UploadHandler(w http.ResponseWriter, r *http.Request, c Context)
```


## type Authorizer
``` go
type Authorizer interface {
    Authorize(token string, c Context) (string, error)
}
```










## type Bucket
``` go
type Bucket map[string][]byte
```










## type Collection
``` go
type Collection struct {
    Items map[string]Item
    Name  string
    Slug  string
}
```










## type Context
``` go
type Context struct {
    Storage      Storage
    Datastore    Datastore
    UsageTracker *UsageTracker
    Authorizer   Authorizer
    Bucket       string
    RootDomain   string
}
```










## type Datastore
``` go
type Datastore interface {
    Init(dbName string) error
    CreateCollection(slug, name string) (Collection, error)
    UpdateCollection(slug, name string) error
    GetCollectionData(slug string) (Collection, error)
    GetCollectionItems(slug string) (map[string]Item, error)
    AddItemToCollection(slug string, item Item) error
    GetItemFromCollection(slug, tag string) (Item, error)
}
```








### func NewMemDatastore
``` go
func NewMemDatastore() Datastore
```

### func NewMySQLDatastore
``` go
func NewMySQLDatastore(dsn string) (Datastore, error)
```



## type GoogleCloudStorage
``` go
type GoogleCloudStorage struct {
    *storage.Service
}
```










### func (\*GoogleCloudStorage) Delete
``` go
func (gcs *GoogleCloudStorage) Delete(bucket, tmp string) error
```


### func (\*GoogleCloudStorage) Download
``` go
func (gcs *GoogleCloudStorage) Download(bucket, id string, w io.Writer, c Context) (int64, error)
```


### func (\*GoogleCloudStorage) Move
``` go
func (gcs *GoogleCloudStorage) Move(srcBucket, src, dstBucket, dst string, c Context) error
```


### func (\*GoogleCloudStorage) Upload
``` go
func (gcs *GoogleCloudStorage) Upload(bucket, tmp string, r io.Reader, c Context, errs chan error, done chan struct{})
```


## type GoogleOAuth2Authorizer
``` go
type GoogleOAuth2Authorizer struct {
    ClientID string

    sync.Mutex
    // contains filtered or unexported fields
}
```








### func NewGoogleOAuth2Authorizer
``` go
func NewGoogleOAuth2Authorizer(id string) *GoogleOAuth2Authorizer
```



### func (\*GoogleOAuth2Authorizer) Authorize
``` go
func (g *GoogleOAuth2Authorizer) Authorize(token string, c Context) (string, error)
```


## type Handler
``` go
type Handler func(w http.ResponseWriter, r *http.Request, c Context)
```










## type Item
``` go
type Item struct {
    Blob   string
    Bucket string
    Tag    string
}
```










## type Memstorage
``` go
type Memstorage map[string]Bucket
```










### func (Memstorage) Delete
``` go
func (m Memstorage) Delete(bucket, tmp string) error
```


### func (Memstorage) Download
``` go
func (m Memstorage) Download(bucket, id string, w io.Writer, c Context) (int64, error)
```


### func (Memstorage) Move
``` go
func (m Memstorage) Move(srcBucket, src, dstBucket, dst string, c Context) error
```


### func (Memstorage) Upload
``` go
func (m Memstorage) Upload(bucket, tmp string, r io.Reader, c Context, errs chan error, done chan struct{})
```


## type Memstore
``` go
type Memstore map[string]*Collection
```










### func (Memstore) AddItemToCollection
``` go
func (m Memstore) AddItemToCollection(slug string, item Item) error
```


### func (Memstore) CreateCollection
``` go
func (m Memstore) CreateCollection(slug, name string) (Collection, error)
```


### func (Memstore) GetCollectionData
``` go
func (m Memstore) GetCollectionData(slug string) (Collection, error)
```


### func (Memstore) GetCollectionItems
``` go
func (m Memstore) GetCollectionItems(slug string) (map[string]Item, error)
```


### func (Memstore) GetItemFromCollection
``` go
func (m Memstore) GetItemFromCollection(slug, tag string) (Item, error)
```


### func (Memstore) Init
``` go
func (m Memstore) Init(dbName string) error
```


### func (Memstore) UpdateCollection
``` go
func (m Memstore) UpdateCollection(slug, name string) error
```


## type SQLStore
``` go
type SQLStore sql.DB
```










### func (\*SQLStore) AddItemToCollection
``` go
func (s *SQLStore) AddItemToCollection(slug string, item Item) error
```


### func (\*SQLStore) CreateCollection
``` go
func (s *SQLStore) CreateCollection(slug, name string) (Collection, error)
```


### func (\*SQLStore) GetCollectionData
``` go
func (s *SQLStore) GetCollectionData(slug string) (Collection, error)
```


### func (\*SQLStore) GetCollectionItems
``` go
func (s *SQLStore) GetCollectionItems(slug string) (map[string]Item, error)
```


### func (\*SQLStore) GetItemFromCollection
``` go
func (s *SQLStore) GetItemFromCollection(slug, tag string) (Item, error)
```


### func (\*SQLStore) Init
``` go
func (s *SQLStore) Init(name string) error
```


### func (\*SQLStore) UpdateCollection
``` go
func (s *SQLStore) UpdateCollection(slug, name string) error
```


## type Storage
``` go
type Storage interface {
    Upload(bucket, tmp string, r io.Reader, c Context, errs chan error, done chan struct{})
    Delete(bucket, tmp string) error
    Move(srcBucket, src, dstBucket, dst string, c Context) error
    Download(bucket, id string, w io.Writer, c Context) (int64, error)
}
```








### func NewGCSStorage
``` go
func NewGCSStorage(gcsClientEmail, gcsTokenURI string, gcsPemBytes []byte) (Storage, error)
```

### func NewMemStorage
``` go
func NewMemStorage() Storage
```



## type Usage
``` go
type Usage struct {
    UploadedBytes        int64
    UploadBytesChan      chan int64
    DownloadedBytes      int64
    DownloadBytesChan    chan int64
    UploadRequests       int64
    UploadRequestsChan   chan int64
    DownloadRequests     int64
    DownloadRequestsChan chan int64
}
```










## type UsageTracker
``` go
type UsageTracker struct {
    sync.Mutex
    // contains filtered or unexported fields
}
```








### func NewUsageTracker
``` go
func NewUsageTracker() *UsageTracker
```



### func (\*UsageTracker) TrackDownloads
``` go
func (u *UsageTracker) TrackDownloads(id string) (bytes, requests chan int64)
```


### func (\*UsageTracker) TrackUploads
``` go
func (u *UsageTracker) TrackUploads(id string) (bytes, requests chan int64)
```








- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)