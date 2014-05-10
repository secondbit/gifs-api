package api

import (
	"sync"
)

func NewUsageTracker() *UsageTracker {
	return &UsageTracker{
		usages: make(map[string]*Usage),
	}
}

type UsageTracker struct {
	usages map[string]*Usage
	sync.Mutex
}

func (u *UsageTracker) TrackUploads(id string) (bytes, requests chan int64) {
	u.Lock()
	defer u.Unlock()
	if _, ok := u.usages[id]; !ok {
		u.usages[id] = &Usage{
			UploadBytesChan:      make(chan int64),
			DownloadBytesChan:    make(chan int64),
			UploadRequestsChan:   make(chan int64),
			DownloadRequestsChan: make(chan int64),
		}
	}
	return u.usages[id].UploadBytesChan, u.usages[id].UploadRequestsChan
}

func (u *UsageTracker) TrackDownloads(id string) (bytes, requests chan int64) {
	u.Lock()
	defer u.Unlock()
	if _, ok := u.usages[id]; !ok {
		u.usages[id] = &Usage{
			UploadBytesChan:      make(chan int64),
			DownloadBytesChan:    make(chan int64),
			UploadRequestsChan:   make(chan int64),
			DownloadRequestsChan: make(chan int64),
		}
	}
	return u.usages[id].DownloadBytesChan, u.usages[id].DownloadRequestsChan
}

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

func (u *Usage) collect() {
	for {
		select {
		case b, ok := <-u.UploadBytesChan:
			if !ok {
				u.UploadBytesChan = nil
			}
			u.UploadedBytes += b
		case r, ok := <-u.UploadRequestsChan:
			if !ok {
				u.UploadRequestsChan = nil
			}
			u.UploadRequests += r
		case b, ok := <-u.DownloadBytesChan:
			if !ok {
				u.DownloadBytesChan = nil
			}
			u.DownloadedBytes += b
		case r, ok := <-u.DownloadRequestsChan:
			if !ok {
				u.DownloadRequestsChan = nil
			}
			u.DownloadRequests += r
		}
		if u.UploadBytesChan == nil && u.UploadRequestsChan == nil && u.DownloadBytesChan == nil && u.DownloadRequestsChan == nil {
			break
		}
	}
}
