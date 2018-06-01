package object

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
)

// Blob represents a got object whose contents are the bytes of a file
type Blob struct {
	*Object
}

// newBlob reads in the file at filePath, and stores it as a new object.
func newBlob(filePath string) (*Blob, error) {
	// check that filePath is a file and not a directory
	if fi, err := os.Stat(filePath); err != nil {
		return nil, err
	} else if fi.IsDir() {
		return nil, errors.New("cannot create a blob out of a directory")
	}

	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return &Blob{
		&Object{
			Type:    "blob",
			Content: fileBytes,
		},
	}, nil
}

// saveBlob compresses a blob, and saves it to the objects directory
// for got repository specified by the given gotRootDir
func (b Blob) saveBlob(gotRootDir string) error {
	hash, err := b.hash()
	if err != nil {
		return err
	}

	blobPath := path.Join(gotRootDir, ".got/objects", hash[:2], hash[2:])

	if err := b.Object.save(blobPath); err != nil {
		return err
	}

	return nil
}

// LoadBlob uncompresses an object file with the given hash,
// and loads the data into a Blob struct
func LoadBlob(gotRootDir, hash string) (*Blob, error) {
	blobPath := path.Join(gotRootDir, ".got/objects", hash[:2], hash[2:])

	object, err := load(blobPath)
	if err != nil {
		return nil, err
	}

	return &Blob{
		object,
	}, nil
}
