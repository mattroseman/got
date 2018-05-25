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
	Mode        int
	Type        string
	HashPointer string
	Name        string
}

// NewTree creates a new tree given a filepath
func NewTree(filePath, gotRootDir string) (*Tree, error) {
	children := make([]treeChild, 0)

	// Convert filePath and gotRootDir to absolute paths, if they aren't already
	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	gotRootDir, err = filepath.Abs(gotRootDir)
	if err != nil {
		return nil, err
	}

	// Double check that given filePath is withing gotRootDir
	if !strings.HasPrefix(filePath, gotRootDir) {
		return nil, errors.New("can't construct tree, filePath isn't within gotRootDir")
	}

	if fi, err := os.Stat(filePath); err != nil {
		return nil, err
	} else if fi.IsDir() {
		// If given filePath is a directory
		files, err := ioutil.ReadDir(filePath)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if file.IsDir() {
				subtree, err := NewTree(path.Join(filePath, file.Name()), gotRootDir)
				if err != nil {
					return nil, err
				}

				// get subtree hash
				hash, err := subtree.Hash()
				if err != nil {
					return nil, err
				}

				// add child to Tree
				children = append(children, treeChild{
					Mode:        040000,
					Type:        "tree",
					HashPointer: hash,
					Name:        file.Name(),
				})
			} else {
				blob, err := NewBlob(path.Join(filePath, file.Name()), gotRootDir)
				if err != nil {
					return nil, err
				}

				hash, err := blob.Hash()
				if err != nil {
					return nil, err
				}

				children = append(children, treeChild{
					Mode:        100644,
					Type:        "blob",
					HashPointer: hash,
					Name:        file.Name(),
				})
			}
		}
	} else {
		// If given filePath is a file
		blob, err := NewBlob(filePath, gotRootDir)
		if err != nil {
			return nil, err
		}

		// get blob hash
		hash, err := blob.Hash()
		if err != nil {
			return nil, err
		}

		// tree has just one child which is a pointer to the file contents blob
		children = append(children, treeChild{
			Mode:        100644,
			Type:        "blob",
			HashPointer: hash,
			Name:        fi.Name(),
		})
	}

	// save contents of this Tree to the objects directory
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
		content += fmt.Sprintf("%d %s %s %s",
			child.Mode, child.Type, child.HashPointer, child.Name,
		)

		if i != len(children)-1 {
			content += "\n"
		}
	}
	return []byte(content)
}
