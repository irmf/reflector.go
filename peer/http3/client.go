package http3

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/lbryio/lbry.go/v2/extras/errors"
	"github.com/lbryio/lbry.go/v2/stream"
	"github.com/irmf/reflector.go/internal/metrics"
	"github.com/irmf/reflector.go/store"

	"github.com/lucas-clemente/quic-go/http3"
)

// Client is an instance of a client connected to a server.
type Client struct {
	Timeout      time.Duration
	conn         *http.Client
	roundTripper *http3.RoundTripper
	ServerAddr   string
}

// Close closes the connection with the client.
func (c *Client) Close() error {
	c.conn.CloseIdleConnections()
	return c.roundTripper.Close()
}

// GetStream gets a stream
func (c *Client) GetStream(sdHash string, blobCache store.BlobStore) (stream.Stream, error) {
	var sd stream.SDBlob

	b, err := c.GetBlob(sdHash)
	if err != nil {
		return nil, err
	}

	err = sd.FromBlob(b)
	if err != nil {
		return nil, err
	}

	s := make(stream.Stream, len(sd.BlobInfos)+1-1) // +1 for sd blob, -1 for last null blob
	s[0] = b

	for i := 0; i < len(sd.BlobInfos)-1; i++ {
		s[i+1], err = c.GetBlob(hex.EncodeToString(sd.BlobInfos[i].BlobHash))
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// HasBlob checks if the blob is available
func (c *Client) HasBlob(hash string) (bool, error) {
	resp, err := c.conn.Get(fmt.Sprintf("https://%s/has/%s", c.ServerAddr, hash))
	if err != nil {
		return false, errors.Err(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, errors.Err("non 200 status code returned: %d", resp.StatusCode)
}

// GetBlob gets a blob
func (c *Client) GetBlob(hash string) (stream.Blob, error) {
	resp, err := c.conn.Get(fmt.Sprintf("https://%s/get/%s", c.ServerAddr, hash))
	if err != nil {
		return nil, errors.Err(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("%s blob not found %d\n", hash, resp.StatusCode)
		return nil, errors.Err(store.ErrBlobNotFound)
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.Err("non 200 status code returned: %d", resp.StatusCode)
	}

	tmp := getBuffer()
	defer putBuffer(tmp)

	written, err := io.Copy(tmp, resp.Body)
	if err != nil {
		return nil, errors.Err(err)
	}

	blob := make([]byte, written)
	copy(blob, tmp.Bytes())

	metrics.MtrInBytesUdp.Add(float64(len(blob)))

	return blob, nil
}

// buffer pool to reduce GC
// https://www.captaincodeman.com/2017/06/02/golang-buffer-pool-gotcha
var buffers = sync.Pool{
	// New is called when a new instance is needed
	New: func() interface{} {
		buf := make([]byte, 0, stream.MaxBlobSize)
		return bytes.NewBuffer(buf)
	},
}

// getBuffer fetches a buffer from the pool
func getBuffer() *bytes.Buffer {
	return buffers.Get().(*bytes.Buffer)
}

// putBuffer returns a buffer to the pool
func putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	buffers.Put(buf)
}
