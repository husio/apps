package webcron

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

var ErrNotFound = errors.New("not found")

type FileStorage struct {
	fd file

	mu    sync.Mutex
	cache map[string]Job
}

type file interface {
	io.ReadWriteSeeker
	io.Closer
}

func NewFileStorage(path string) (*FileStorage, error) {
	cache := make(map[string]Job)

	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return nil, fmt.Errorf("cannot open storage file: %s", err)
	}

	if err := json.NewDecoder(fd).Decode(&cache); err != nil && err != io.EOF {
		return nil, fmt.Errorf("cannot load storage file: %s", err)
	}

	fs := &FileStorage{
		fd:    fd,
		cache: cache,
	}
	return fs, nil
}

func (fs *FileStorage) Close() error {
	return fs.fd.Close()
}

func (fs *FileStorage) Add(job Job) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.cache[job.ID] = job
	if err := fs.persistCache(); err != nil {
		delete(fs.cache, job.ID)
		return fmt.Errorf("cannot persist: %s", err)
	}
	return nil
}

func (fs *FileStorage) Del(jobID string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	job, ok := fs.cache[jobID]
	if !ok {
		return ErrNotFound
	}

	delete(fs.cache, jobID)
	if err := fs.persistCache(); err != nil {
		fs.cache[jobID] = job
		return fmt.Errorf("cannot persist: %s", err)
	}
	return nil
}

func (fs *FileStorage) List(limit, offset int) ([]Job, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if offset > len(fs.cache) {
		return nil, ErrOutOfRange
	}

	if len(fs.cache) < limit+offset {
		limit = len(fs.cache) - offset
	}
	res := make([]Job, 0, limit)
	for _, job := range fs.cache {
		res = append(res, job)
	}
	return res, nil
}

var ErrOutOfRange = errors.New("out of range")

func (fs *FileStorage) persistCache() error {
	b, err := json.MarshalIndent(fs.cache, "", "\t")
	if err != nil {
		return fmt.Errorf("cannot serialize: %s", err)
	}
	if _, err := fs.fd.Seek(0, os.SEEK_SET); err != nil {
		return fmt.Errorf("cannot seek storage file: %s", err)
	}
	if _, err := fs.fd.Write(b); err != nil {
		return fmt.Errorf("cannot write to file: %s", err)
	}
	return nil
}
