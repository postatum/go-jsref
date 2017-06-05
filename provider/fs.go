package provider

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/lestrrat/go-pdebug"
	"github.com/pkg/errors"
)

// NewFS creates a new Provider that looks for JSON documents
// from the local file system. Documents are only searched
// within `root`
func NewFS(root string) *FS {
	return &FS{
		mp:   NewMap(),
		Root: root,
	}
}

// Get fetches the document specified by the `key` argument.
// Everything other than `.Path` is ignored.
// Note that once a document is read, it WILL be cached for the
// duration of this object, unless you call `Reset`
func (fp *FS) Get(key *url.URL) (out interface{}, err error) {
	d, err := fp.GetBytes(key)
	if err != nil {
		return nil, err
	}

	var x interface{}
	if err := json.Unmarshal(d, &x); err != nil {
		return nil, errors.Wrap(err, "failed to parse JSON local resource")
	}

	return x, nil
}

// GetBytes fetches the document specified by the `key` argument and returns its bytes content
func (fp *FS) GetBytes(key *url.URL) ([]byte, error) {
	var err error
	if pdebug.Enabled {
		g := pdebug.Marker("provider.FS.Get(%s)", key.String()).BindError(&err)
		defer g.End()
	}

	if strings.ToLower(key.Scheme) != "file" {
		return nil, errors.New("unsupported scheme '" + key.Scheme + "'")
	}

	// Everything other than "Path" is ignored
	path := filepath.Clean(filepath.Join(fp.Root, key.Path))

	mpkey := &url.URL{Path: path}
	if x, err := fp.mp.Get(mpkey); err == nil {
		return x.([]byte), nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to stat local resource")
	}

	if fi.IsDir() {
		return nil, errors.New("target is not a file")
	}

	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read local file content")
	}

	fp.mp.Set(path, d)

	return d, nil
}

// Reset resets the in memory cache of JSON documents
func (fp *FS) Reset() error {
	return fp.mp.Reset()
}
