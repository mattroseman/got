package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Object represents a got object containing a tree or blob.
type Object struct {
	Type    string
	Content []byte
}

// New creates a new blob for file at filePath, and saves the blob to the
// got repositories objects direcotry.
func New(gotRootDir, filePath string) error {
	// TODO if path is directory, walk directory and add all blobs

	blob, err := newBlob(filePath)
	if err != nil {
		return err
	}

	if err := blob.saveBlob(gotRootDir); err != nil {
		return err
	}

	// TODO add blob to index

	return nil
}

// header computes the header for an object file based on its content and type.
func (o Object) header() []byte {
	return []byte(fmt.Sprintf("%s %d\000", o.Type, len(o.Content)))
}

// hash computes the SHA-1 hash of an object's header + content.
// returns hash as a hexadecimal string
func (o Object) hash() (string, error) {
	// combine an objects header and it's content
	objectBytes := append(o.header(), o.Content...)

	// compute the SHA-1 hash of objectBytes
	h := sha1.New()
	if _, err := io.Copy(h, bytes.NewReader(objectBytes)); err != nil {
		return "", err
	}

	// convert hash to hexadecimal string and return
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// compress calculates an objects zlib compressed bytes, which includes content + header
func (o Object) compress() []byte {
	// combine an objects header and it's content
	objectBytes := append(o.header(), o.Content...)

	var buf bytes.Buffer
	wc := zlib.NewWriter(&buf)
	wc.Write(objectBytes)
	wc.Close()

	return buf.Bytes()
}

// uncompress calculates the zlib uncompressed bytes, and puts the data in an Object struct
func uncompress(data []byte) (*Object, error) {
	// create zlib reader off of data bytes
	r := bytes.NewReader(data)
	rc, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	// read the zlib uncompressed bytes into a buffer
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(rc); err != nil {
		return nil, err
	}

	// read the bytes from the buffer up until the first null bytes, indicating end of
	// the header
	header, err := buf.ReadBytes('\000')
	if err != nil {
		return nil, err
	}

	// extract the type bytes from header, and convert to a string
	objectType := string(header[:bytes.IndexByte(header, byte(' '))])

	// read the rest of the bytes in buffer, which is the object content
	content := buf.Bytes()

	return &Object{
		Type:    objectType,
		Content: content,
	}, nil
}

// save compresses an object, and saves it to the given filePath
func (o Object) save(filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Write(o.compress())

	return nil
}

// load reads in an object file and returns a new Object struct witth the loaded content
func load(filePath string) (*Object, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return uncompress(fileBytes)
}
