package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

func main() {
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("go run main.go . [--files]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "--files"
	err := dirTree(os.Stdout, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

type osFileInfoSlice []os.FileInfo

func (s osFileInfoSlice) Len() int      { return len(s) }
func (s osFileInfoSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s osFileInfoSlice) Less(i, j int) bool {
	var first os.FileInfo = s[i]
	var second os.FileInfo = s[j]

	//var firstIsDir bool = first.IsDir()
	//var secondIsDir bool = second.IsDir()

	//if firstIsDir && secondIsDir {
	return sort.StringsAreSorted([]string{first.Name(), second.Name()})
	//}

	//if firstIsDir {
	//	return true
	//} else {
	//	return false
	//}
}

func getFormattedFileSize(fileSize int64) string {
	if fileSize > 0 {
		return fmt.Sprintf("(%db)", fileSize)
	} else {
		return "(empty)"
	}
}

type Node struct {
	parent *Node
	childs map[int]*Node

	index    int
	info     os.FileInfo
	fullPath string
}

func (n *Node) GetParent() *Node {
	return n.parent
}

func (n *Node) GetIndex() int {
	return n.index
}

func (n *Node) GetInfo() os.FileInfo {
	return n.info
}

func (n *Node) IsRootNode() bool {
	return n.index == -1
}

func CreateNodeFromFileInfo(parent *Node, index int, fileInfo os.FileInfo) Node {
	var returnNode Node = Node{
		parent:   parent,
		childs:   map[int]*Node{},
		index:    index,
		info:     fileInfo,
		fullPath: parent.fullPath + "/" + fileInfo.Name(),
	}

	returnNode.childs = make(map[int]*Node)

	return returnNode
}

func CreateNodeFromPath(parent *Node, index int, path string) (Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return Node{}, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return Node{}, err
	}

	return CreateNodeFromFileInfo(parent, index, fileInfo), nil

}

func CreateRootNodeFromFileInfo(fileInfo os.FileInfo) Node {
	var returnNode Node = Node{
		parent:   &Node{},
		childs:   map[int]*Node{},
		index:    -1,
		info:     fileInfo,
		fullPath: fileInfo.Name(),
	}

	returnNode.childs = make(map[int]*Node)

	return returnNode
}

func CreateRootNodeFromPath(path string) (Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return Node{}, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return Node{}, err
	}

	return CreateRootNodeFromFileInfo(fileInfo), nil
}

func (n *Node) GetDrawChar() string {
	var drawChar string
	if n.index == len(n.parent.childs)-1 {
		drawChar = "└"
	} else {
		drawChar = "├"
	}

	return drawChar
}

func (n *Node) GetDrawableString() string {
	var drawString string

	var drawChar string = n.GetDrawChar()

	var depthLevel int = n.GetDepthLevel()

	var parentNode *Node = n.parent
	for i := depthLevel*4 - 4; i > 0; i -= 4 {
		if parentNode.index != len(parentNode.parent.childs)-1 {
			drawString = "│\t" + drawString
		} else {
			drawString = "\t" + drawString
		}

		parentNode = parentNode.parent
	}

	var formattedSize string
	if !n.info.IsDir() {
		formattedSize = " " + getFormattedFileSize(n.info.Size())
	}

	drawString += fmt.Sprintf("%s───%s%s\n", drawChar, n.info.Name(), formattedSize)

	return drawString
}

func (n *Node) GetDepthLevel() int {
	return strings.Count(n.fullPath, "/")
}

func (n *Node) Draw(writer io.Writer) {
	if !n.IsRootNode() {
		fmt.Fprintf(writer, "%s", n.GetDrawableString())
	}

	for i := 0; i < len(n.childs); i++ {
		var node *Node = n.childs[i]
		node.Draw(writer)
	}
}

func dirTreeWorker(parentNode *Node, writer io.Writer, path string, printFiles bool, depthLevel int) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	var newFiles []os.FileInfo
	for _, fileInfo := range files {
		if !fileInfo.IsDir() && !printFiles {
			continue
		}

		newFiles = append(newFiles, fileInfo)
	}

	var wraperedFiles osFileInfoSlice = osFileInfoSlice(newFiles)
	sort.Sort(wraperedFiles)
	for index, fileInfo := range wraperedFiles {
		if !fileInfo.IsDir() && !printFiles {
			continue
		}

		var node Node = CreateNodeFromFileInfo(parentNode, index, fileInfo)
		parentNode.childs[index] = &node
	}

	for i := 0; i < len(parentNode.childs); i++ {
		var nodePtr *Node = parentNode.childs[i]

		nodePtr.Draw(writer)

		if nodePtr.info.IsDir() {
			dirTreeWorker(nodePtr, writer, path+"/"+nodePtr.info.Name(), printFiles, depthLevel+1)
		}

		parentNode.childs[i] = nil
	}

	return nil
}

func dirTree(writer io.Writer, path string, printFiles bool) error {
	rootNode, err := CreateRootNodeFromPath(path)
	if err != nil {
		return err
	}

	dirTreeWorker(&rootNode, writer, path, printFiles, 1)

	return nil
}
