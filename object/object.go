package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// Object represents a got object containing a tree or blob
type Object struct {
	// Type is either blob or tree
	Type string
	// Length is the number of bytes content is
	Length int
	// Content is the actuall content of the object
	Content []byte

	Header []byte
}

// NewBlob creates a new Object with mode blob and content from file at the given path
func NewBlob(filePath string) (*Object, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return &Object{
		Type:    "blob",
		Length:  len(fileBytes),
		Content: fileBytes,
		Header:  []byte(fmt.Sprintf("%s %d\000", "blob", len(fileBytes))),
	}, nil
}

// hash combines this objects type, length, and content and computes the SHA-1 hash
// returning that as a string
func (object Object) hash() (string, error) {
	fileBytes := append(object.Header, object.Content...)

	h := sha1.New()
	if _, err := io.Copy(h, bytes.NewReader(fileBytes)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// compress combines an objects header and content and returns zlib compressed bytes
func (object Object) compress() []byte {
	fileBytes := append(object.Header, object.Content...)

	var buf bytes.Buffer
	wc := zlib.NewWriter(&buf)
	wc.Write(fileBytes)
	wc.Close()

	return buf.Bytes()
}

// uncompress takes an object files compressed bytes, uncompresses them and pareses the
// data into an Object struct
func uncompress(objectFile *os.File) (*Object, error) {
	rc, err := zlib.NewReader(objectFile)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(rc)
	if err != nil {
		return nil, err
	}

	header, err := buf.ReadBytes('\000')
	if err != nil {
		return nil, err
	}
	objectType := string(header[:bytes.IndexByte(header, byte(' '))])
	contentLength, err := strconv.Atoi(
		string(
			header[bytes.IndexByte(header, byte(' '))+1 : bytes.IndexByte(header, byte('\000'))],
		),
	)
	if err != nil {
		return nil, err
	}

	// type mode length
	return &Object{
		Type:    objectType,
		Length:  contentLength,
		Content: buf.Bytes(),
		Header:  header,
	}, nil
}

// Save will save this object to the specified objects directory, and return it's hash
func (object Object) Save(objectsDir string) (string, error) {
	hash, err := object.hash()
	if err != nil {
		return "", err
	}
	compressedBytes := object.compress()

	// make object directory named after first 2 bytes of sha-1 hash if it doesn't already exist
	objectDir := path.Join(objectsDir, hash[:2])
	if _, err := os.Stat(objectDir); os.IsNotExist(err) {
		if err := os.Mkdir(objectDir, 0755); err != nil {
			return "", err
		}
	}

	// TODO what if this file has already been added
	// create file for storing the compressed contents of this object
	objectFile, err := os.Create(path.Join(objectsDir, hash[:2], hash[2:]))
	if err != nil {
		return "", err
	}
	objectFile.Write(compressedBytes)
	objectFile.Close()

	return hash, nil
}

// Load will read in an object with the given hash, and return a pointer to a new
// Object struct
func Load(objectsDir, hash string) (*Object, error) {
	objectFilePath := path.Join(objectsDir, hash[:2], hash[2:])

	// check that the directory for an object with the given hash exists in the given objectsDir
	if _, err := os.Stat(objectFilePath); os.IsNotExist(err) {
		return nil, errors.New("no got object with given hash was found")
	}

	// read in the object file
	objectFile, err := os.Open(objectFilePath)
	if err != nil {
		return nil, err
	}
	defer objectFile.Close()

	return uncompress(objectFile)
}
