package object

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Tree represents a got object containing pointers to other got objects
type Tree struct {
	*Object
	Children []treeChild
}

type treeChild struct {
	Mode        string
	Type        string
	HashPointer string
	Name        string
}

// NewTree creates a new tree given a filepath to a directory
func NewTree(dirPath, gotRootDir string) (*Tree, error) {
	children := make([]treeChild, 0)

	// Convert dirPath and gotRootDir to absolute paths, if they aren't already
	dirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, err
	}
	gotRootDir, err = filepath.Abs(gotRootDir)
	if err != nil {
		return nil, err
	}

	// Double check that given dirPath is within gotRootDir
	if !strings.HasPrefix(dirPath, gotRootDir) {
		return nil, errors.New("can't construct tree, dirPath isn't within gotRootDir")
	}

	// Check that dirPath is a directory
	if fi, err := os.Stat(dirPath); err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, errors.New("cannot create tree off of single file")
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	// iterate through all the contents in dirPath
	for _, file := range files {
		if file.IsDir() {
			// for subdirectories, recursively call NewTree on them, and then add the returned
			// subtree as a child to tree
			subtree, err := NewTree(path.Join(dirPath, file.Name()), gotRootDir)
			if err != nil {
				return nil, err
			}

			// get subtree hash
			hash, err := subtree.Hash()
			if err != nil {
				return nil, err
			}

			// add child to tree
			children = append(children, treeChild{
				Mode:        "040000",
				Type:        "tree",
				HashPointer: hash,
				Name:        file.Name(),
			})
		} else {
			// For files in directory, create a new blob object
			blob, err := NewBlob(path.Join(dirPath, file.Name()), gotRootDir)
			if err != nil {
				return nil, err
			}

			hash, err := blob.Hash()
			if err != nil {
				return nil, err
			}

			// add the file's blob object as a child to tree
			children = append(children, treeChild{
				Mode:        "100644",
				Type:        "blob",
				HashPointer: hash,
				Name:        file.Name(),
			})
		}
	}

	content := generateContent(children)
	header := []byte(fmt.Sprintf("tree %d\000", len(content)))

	tree := &Tree{
		&Object{
			Type:    "tree",
			Length:  len(content),
			Content: content,
			Header:  header,
		},
		children,
	}

	if _, err := tree.Save(gotRootDir); err != nil {
		return nil, err
	}

	return tree, nil
}

// generateContent generates a tree's content based on it's children
func generateContent(children []treeChild) []byte {
	content := ""
	for i, child := range children {
		content += fmt.Sprintf("%s %s %s %s",
			child.Mode, child.Type, child.HashPointer, child.Name,
		)

		if i != len(children)-1 {
			content += "\n"
		}
	}
	return []byte(content)
}
