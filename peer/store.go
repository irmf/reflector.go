package peer

import (
	"time"

	"github.com/lbryio/lbry.go/v2/extras/errors"
	"github.com/lbryio/lbry.go/v2/stream"
)

// Store is a blob store that gets blobs from a peer.
// It satisfies the store.BlobStore interface but cannot put or delete blobs.
type Store struct {
	opts StoreOpts
}

// StoreOpts allows to set options for a new Store.
type StoreOpts struct {
	Address string
	Timeout time.Duration
}

// NewStore makes a new peer store.
func NewStore(opts StoreOpts) *Store {
	return &Store{opts: opts}
}

func (p *Store) getClient() (*Client, error) {
	c := &Client{Timeout: p.opts.Timeout}
	err := c.Connect(p.opts.Address)
	return c, errors.Prefix("connection error", err)
}

func (p *Store) Name() string { return "peer" }

// Has asks the peer if they have a hash
func (p *Store) Has(hash string) (bool, error) {
	c, err := p.getClient()
	if err != nil {
		return false, err
	}
	defer c.Close()
	return c.HasBlob(hash)
}

// Get downloads the blob from the peer
func (p *Store) Get(hash string) (stream.Blob, error) {
	c, err := p.getClient()
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return c.GetBlob(hash)
}

// Put is not supported
func (p *Store) Put(hash string, blob stream.Blob) error {
	panic("PeerStore cannot put or delete blobs")
}

// PutSD is not supported
func (p *Store) PutSD(hash string, blob stream.Blob) error {
	panic("PeerStore cannot put or delete blobs")
}

// Delete is not supported
func (p *Store) Delete(hash string) error {
	panic("PeerStore cannot put or delete blobs")
}
