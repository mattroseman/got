package object

import (
	"fmt"
	"io/ioutil"
)

// Blob represents a got object containing the contents of some file
type Blob struct {
	*Object
}

// NewBlob creates a new blob object that contains the contents of the given file.
func NewBlob(filePath, gotRootDir string) (*Blob, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	blob := &Blob{
		&Object{
			Type:    "blob",
			Length:  len(fileBytes),
			Content: fileBytes,
			Header:  []byte(fmt.Sprintf("%s %d\000", "blob", len(fileBytes))),
		},
	}

	if _, err := blob.Save(gotRootDir); err != nil {
		return nil, err
	}

	return blob, err
}
