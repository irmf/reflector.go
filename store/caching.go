package store

import (
	"time"

	"github.com/lbryio/lbry.go/v2/extras/errors"
	"github.com/lbryio/lbry.go/v2/stream"

	"github.com/irmf/reflector.go/internal/metrics"
)

// CachingStore combines two stores, typically a local and a remote store, to improve performance.
// Accessed blobs are stored in and retrieved from the cache. If they are not in the cache, they
// are retrieved from the origin and cached. Puts are cached and also forwarded to the origin.
type CachingStore struct {
	origin, cache BlobStore
	component     string
}

// NewCachingStore makes a new caching disk store and returns a pointer to it.
func NewCachingStore(component string, origin, cache BlobStore) *CachingStore {
	return &CachingStore{
		component: component,
		origin:    WithSingleFlight(component, origin),
		cache:     cache,
	}
}

const nameCaching = "caching"

// Name is the cache type name
func (c *CachingStore) Name() string { return nameCaching }

// Has checks the cache and then the origin for a hash. It returns true if either store has it.
func (c *CachingStore) Has(hash string) (bool, error) {
	has, err := c.cache.Has(hash)
	if has || err != nil {
		return has, err
	}
	return c.origin.Has(hash)
}

// Get tries to get the blob from the cache first, falling back to the origin. If the blob comes
// from the origin, it is also stored in the cache.
func (c *CachingStore) Get(hash string) (stream.Blob, error) {
	start := time.Now()
	blob, err := c.cache.Get(hash)
	if err == nil || !errors.Is(err, ErrBlobNotFound) {
		metrics.CacheHitCount.With(metrics.CacheLabels(c.cache.Name(), c.component)).Inc()
		rate := float64(len(blob)) / 1024 / 1024 / time.Since(start).Seconds()
		metrics.CacheRetrievalSpeed.With(map[string]string{
			metrics.LabelCacheType: c.cache.Name(),
			metrics.LabelComponent: c.component,
			metrics.LabelSource:    "cache",
		}).Set(rate)
		return blob, err
	}

	metrics.CacheMissCount.With(metrics.CacheLabels(c.cache.Name(), c.component)).Inc()

	blob, err = c.origin.Get(hash)
	if err != nil {
		return nil, err
	}

	err = c.cache.Put(hash, blob)
	return blob, err
}

// Put stores the blob in the origin and the cache
func (c *CachingStore) Put(hash string, blob stream.Blob) error {
	err := c.origin.Put(hash, blob)
	if err != nil {
		return err
	}
	return c.cache.Put(hash, blob)
}

// PutSD stores the sd blob in the origin and the cache
func (c *CachingStore) PutSD(hash string, blob stream.Blob) error {
	err := c.origin.PutSD(hash, blob)
	if err != nil {
		return err
	}
	return c.cache.PutSD(hash, blob)
}

// Delete deletes the blob from the origin and the cache
func (c *CachingStore) Delete(hash string) error {
	err := c.origin.Delete(hash)
	if err != nil {
		return err
	}
	return c.cache.Delete(hash)
}
