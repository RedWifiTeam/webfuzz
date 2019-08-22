package main

import (
	"encoding/gob"
	"os"
)

const (
	// SpecialFiles 特殊文件名
	SpecialFiles = `,head,config,description,index,entries,.ds_store,.gitignore`
	// SpecialDirs 特殊目录名
	SpecialDirs = `.git,.hg,.svn`
)

// FileNode FileNode
type FileNode struct {
	Path  string               `json:"path"`
	Files []string             `json:"files"`
	Nodes map[string]*FileNode `json:"nodes"`
}

func (n *FileNode) init(path string) {
	n.Path = path
	n.Files = []string{}
	n.Nodes = make(map[string]*FileNode, 100)
}

func (n *FileNode) load(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		n = new(FileNode)
		n.init("/")
	} else {
		dec := gob.NewDecoder(file)
		dec.Decode(&n)
	}
}

func (n *FileNode) save(filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		Error.Println("Create error:", err)
	} else {
		enc := gob.NewEncoder(file)
		if err := enc.Encode(n); err != nil {
			Error.Fatal("Save error:", err)
		}
	}
}

func (n *FileNode) getFiles() []string {
	return n.Files
}

func (n *FileNode) getNodes() map[string]*FileNode {
	return n.Nodes
}

func (n *FileNode) getNode(key string) *FileNode {
	if v, ok := n.Nodes[key]; ok {
		return v
	}
	return nil
}
func (n *FileNode) getNodeKeys() []string {
	keys := []string{}
	for k := range n.Nodes {
		keys = append(keys, k)
	}
	return keys
}

func (n *FileNode) addFile(file string) {
	for _, f := range n.Files {
		if f == file {
			return
		}
	}
	Debug.Println("addFile : ", file)
	n.Files = append(n.Files, file)
}
func (n *FileNode) addNode(path string) {
	if n.Nodes == nil {
		n.Nodes = make(map[string]*FileNode)
	}
	var w *FileNode
	w = new(FileNode)
	w.init(path)
	n.Nodes[path] = w
}
