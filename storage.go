package api

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"code.google.com/p/google-api-go-client/googleapi"
	"code.google.com/p/google-api-go-client/storage/v1beta2"
)

var (
	BucketNotFoundError = errors.New("bucket not found")
	BlobNotFoundError   = errors.New("blob not found")
)

type Storage interface {
	Upload(bucket, tmp string, r io.Reader, c Context, errs chan error, done chan struct{})
	Delete(bucket, tmp string) error
	Move(srcBucket, src, dstBucket, dst string, c Context) error
	Download(bucket, id string, w io.Writer, c Context) (int64, error)
}

type GoogleCloudStorage struct {
	*storage.Service
}

func (gcs *GoogleCloudStorage) Upload(bucket, tmp string, r io.Reader, c Context, errs chan error, done chan struct{}) {
	if errs != nil {
		defer close(errs)
	}
	if done != nil {
		defer close(done)
	}
	object := &storage.Object{Name: tmp}
	_, err := gcs.Objects.Insert(bucket, object).Media(r).Do()
	if err != nil && errs != nil {
		errs <- err
	}
}

func (gcs *GoogleCloudStorage) Delete(bucket, tmp string) error {
	return gcs.Objects.Delete(bucket, tmp).Do()
}

func (gcs *GoogleCloudStorage) Move(srcBucket, src, dstBucket, dst string, c Context) error {
	_, err := gcs.Objects.Get(dstBucket, dst).Do()
	if err == nil {
		go del(srcBucket, src, c)
		return nil
	}
	if e, ok := err.(*googleapi.Error); !ok || e.Code != 404 {
		return err
	}
	obj, err := gcs.Objects.Copy(srcBucket, src, dstBucket, dst, nil).Do()
	if err != nil {
		return err
	}
	log.Printf("%+v\n", obj.Owner)
	objectAcl := &storage.ObjectAccessControl{
		Bucket: dstBucket, Entity: "allUsers", Object: dst, Role: "READER",
	}
	_, err = gcs.ObjectAccessControls.Insert(dstBucket, dst, objectAcl).Do()
	if err != nil {
		return err
	}
	go del(srcBucket, src, c)
	return nil
}

func (gcs *GoogleCloudStorage) Download(bucket, id string, w io.Writer, c Context) (int64, error) {
	res, err := gcs.Objects.Get(bucket, id).Do()
	if err != nil {
		return 0, err
	}
	resp, err := http.Get(res.MediaLink)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return io.Copy(w, resp.Body)
}

type Memstorage map[string]Bucket

type Bucket map[string][]byte

func (m Memstorage) Upload(bucket, tmp string, r io.Reader, c Context, errs chan error, done chan struct{}) {
	if errs != nil {
		defer close(errs)
	}
	if done != nil {
		defer close(done)
	}
	if _, ok := m[bucket]; !ok {
		m[bucket] = make(Bucket)
	}
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		errs <- err
		return
	}
	m[bucket][tmp] = bytes
}

func (m Memstorage) Delete(bucket, tmp string) error {
	delete(m[bucket], tmp)
	return nil
}

func (m Memstorage) Move(srcBucket, src, dstBucket, dst string, c Context) error {
	if _, ok := m[srcBucket]; !ok {
		return BucketNotFoundError
	}
	if _, ok := m[srcBucket][src]; !ok {
		return BlobNotFoundError
	}
	if _, ok := m[dstBucket]; !ok {
		m[dstBucket] = make(Bucket)
	}
	m[dstBucket][dst] = m[srcBucket][src]
	return m.Delete(srcBucket, src)
}

func (m Memstorage) Download(bucket, id string, w io.Writer, c Context) (int64, error) {
	if _, ok := m[bucket]; !ok {
		return 0, BucketNotFoundError
	}
	if _, ok := m[bucket][id]; !ok {
		return 0, BlobNotFoundError
	}
	n, err := w.Write(m[bucket][id])
	return int64(n), err
}
