package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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

// New creates a new tree/blob depending on the filepath given
func New(filePath, gotRootDir string) (object *Object, err error) {
	// convert filePath, and gotRootDir to absolute paths
	filePath, err = filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	gotRootDir, err = filepath.Abs(gotRootDir)
	if err != nil {
		return nil, err
	}

	// check that filePath is within gotRootDir
	if !strings.HasPrefix(filePath, gotRootDir) {
		return nil, errors.New("filePath is not withing gotRootDir")
	}

	// check that filePath points to an existing file/directory
	if fi, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("filePath does not exist")
		}

		return nil, err
	} else if fi.IsDir() {
		// create tree for directory
		tree, err := NewTree(filePath, gotRootDir)
		if err != nil {
			return nil, err
		}
		object = tree.Object
	} else {
		// create blob for file
		blob, err := NewBlob(filePath, gotRootDir)
		if err != nil {
			return nil, err
		}
		object = blob.Object
	}

	hash, err := object.Hash()
	if err != nil {
		return nil, err
	}

	// add tree objects for each parent directory until gotRootDir is reached
	parentPath := filepath.Dir(filePath)
	childName := filepath.Base(filePath)
	children := make([]treeChild, 0)
	for parentPath != gotRootDir {
		children = []treeChild{
			{
				Mode:        "040000",
				Type:        "tree",
				HashPointer: hash,
				Name:        childName,
			},
		}

		content := generateContent(children)
		header := []byte(fmt.Sprintf("tree %d\000", len(content)))

		object = &Object{
			Type:    "tree",
			Length:  len(content),
			Content: content,
			Header:  header,
		}

		hash, err = object.Hash()
		if err != nil {
			return nil, err
		}

		if _, err := object.Save(gotRootDir); err != nil {
			return nil, err
		}

		childName = filepath.Base(parentPath)
		parentPath = filepath.Dir(parentPath)
	}

	return object, nil
}

// Hash combines this objects type, length, and content and computes the SHA-1 hash
// returning that as a string
func (object Object) Hash() (string, error) {
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

// Save will save this object to the specifiec got repo directory, and return it's hash
func (object Object) Save(gotRootDir string) (string, error) {
	objectsDir := path.Join(gotRootDir, ".got", "objects")

	hash, err := object.Hash()
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
func Load(gotRootDir, hash string) (*Object, error) {
	objectFilePath := path.Join(gotRootDir, ".got", "objects", hash[:2], hash[2:])

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
