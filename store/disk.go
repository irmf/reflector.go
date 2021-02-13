package store

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/lbryio/lbry.go/v2/extras/errors"
	"github.com/lbryio/lbry.go/v2/stream"
	"github.com/irmf/reflector.go/store/speedwalk"
)

// DiskStore stores blobs on a local disk
type DiskStore struct {
	// the location of blobs on disk
	blobDir string
	// store files in subdirectories based on the first N chars in the filename. 0 = don't create subdirectories.
	prefixLength int

	// true if initOnce ran, false otherwise
	initialized bool
}

// NewDiskStore returns an initialized file disk store pointer.
func NewDiskStore(dir string, prefixLength int) *DiskStore {
	return &DiskStore{
		blobDir:      dir,
		prefixLength: prefixLength,
	}
}

const nameDisk = "disk"

// Name is the cache type name
func (d *DiskStore) Name() string { return nameDisk }

// Has returns T/F or Error if it the blob stored already. It will error with any IO disk error.
func (d *DiskStore) Has(hash string) (bool, error) {
	err := d.initOnce()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(d.path(hash))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Err(err)
	}
	return true, nil
}

// Get returns the blob or an error if the blob doesn't exist.
func (d *DiskStore) Get(hash string) (stream.Blob, error) {
	err := d.initOnce()
	if err != nil {
		return nil, err
	}

	blob, err := ioutil.ReadFile(d.path(hash))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Err(ErrBlobNotFound)
		}
		return nil, errors.Err(err)
	}

	return blob, nil
}

// Put stores the blob on disk
func (d *DiskStore) Put(hash string, blob stream.Blob) error {
	err := d.initOnce()
	if err != nil {
		return err
	}

	err = d.ensureDirExists(d.dir(hash))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(d.path(hash), blob, 0644)
	return errors.Err(err)
}

// PutSD stores the sd blob on the disk
func (d *DiskStore) PutSD(hash string, blob stream.Blob) error {
	return d.Put(hash, blob)
}

// Delete deletes the blob from the store
func (d *DiskStore) Delete(hash string) error {
	err := d.initOnce()
	if err != nil {
		return err
	}

	has, err := d.Has(hash)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	err = os.Remove(d.path(hash))
	return errors.Err(err)
}

// list returns the hashes of blobs that already exist in the blobDir
func (d *DiskStore) list() ([]string, error) {
	err := d.initOnce()
	if err != nil {
		return nil, err
	}

	return speedwalk.AllFiles(d.blobDir, true)
}

func (d *DiskStore) dir(hash string) string {
	if d.prefixLength <= 0 || len(hash) < d.prefixLength {
		return d.blobDir
	}
	return path.Join(d.blobDir, hash[:d.prefixLength])
}

func (d *DiskStore) path(hash string) string {
	return path.Join(d.dir(hash), hash)
}

func (d *DiskStore) ensureDirExists(dir string) error {
	return errors.Err(os.MkdirAll(dir, 0755))
}

func (d *DiskStore) initOnce() error {
	if d.initialized {
		return nil
	}

	err := d.ensureDirExists(d.blobDir)
	if err != nil {
		return err
	}

	d.initialized = true
	return nil
}
