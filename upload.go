package api

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"

	"code.google.com/p/go-uuid/uuid"
)

func Upload(id, collection, tag string, r io.Reader, c Context) (string, error) {
	hashReader, hashWriter := io.Pipe()
	uploadReader, uploadWriter := io.Pipe()
	m := io.MultiWriter(hashWriter, uploadWriter)

	hashChan := make(chan string)
	hashError := make(chan error)
	hashDone := make(chan struct{})

	cpDone := make(chan struct{})
	cpErr := make(chan error)

	uploadDone := make(chan struct{})
	uploadErr := make(chan error)

	n := make(chan int64)

	var bytesWritten int64
	tmp := uuid.NewRandom().String()
	tmp = "tmp/" + tmp

	go hash(hashReader, hashChan, hashError, hashDone)
	go asyncCopy(m, r, n, cpErr, cpDone)
	if c.Storage != nil {
		go c.Storage.Upload(c.Bucket, tmp, uploadReader, c, uploadErr, uploadDone)
	}

	select {
	case err := <-hashError:
		if err != nil {
			hashWriter.CloseWithError(err)
			uploadWriter.CloseWithError(err)
			return "", err
		} else {
			hashWriter.Close()
			uploadWriter.Close()
		}
	case err := <-cpErr:
		if err != nil {
			hashWriter.CloseWithError(err)
			uploadWriter.CloseWithError(err)
			return "", err
		} else {
			hashWriter.Close()
			uploadWriter.Close()
		}
	case err := <-uploadErr:
		if err != nil {
			hashWriter.CloseWithError(err)
			uploadWriter.CloseWithError(err)
			return "", err
		} else {
			hashWriter.Close()
			uploadWriter.Close()
		}
	case <-cpDone:
		hashWriter.Close()
		uploadWriter.Close()
	}
	<-cpDone
	<-hashDone
	if c.Storage != nil {
		<-uploadDone
	}
	bytesWritten = <-n
	finalLocation := <-hashChan
	if c.Storage != nil {
		err := c.Storage.Move(c.Bucket, tmp, c.Bucket, finalLocation, c)
		if err != nil {
			return "", err
		}
	}
	if c.Datastore != nil {
		err := c.Datastore.AddItemToCollection(collection, Item{
			Blob:   finalLocation,
			Bucket: c.Bucket,
			Tag:    tag,
		})
		if err != nil {
			return "", err
		}
	}
	uBytes, uRequests := c.UsageTracker.TrackUploads(id)
	uBytes <- bytesWritten
	uRequests <- 1
	return finalLocation, nil
}

func hash(r io.Reader, resp chan string, errs chan error, done chan struct{}) {
	if resp != nil {
		defer close(resp)
	}
	h := sha1.New()
	go asyncCopy(h, r, nil, errs, done)
	<-done
	resp <- hex.EncodeToString(h.Sum(nil))
}

func del(bucket, tmp string, c Context) {
	err := c.Storage.Delete(bucket, tmp)
	if err != nil {
		log.Printf("Error deleting temporary upload %s in %s: %s\n", tmp, bucket, err)
	}
}

func asyncCopy(dst io.Writer, src io.Reader, n chan int64, errs chan error, done chan struct{}) {
	if errs != nil {
		defer close(errs)
	}
	if done != nil {
		defer close(done)
	}
	buf := make([]byte, 32*1024)
	var written int64
	var err error
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	if errs != nil && err != nil {
		errs <- err
	}
	if n != nil {
		go func(n chan int64, written int64) {
			n <- written
		}(n, written)
	}
}
