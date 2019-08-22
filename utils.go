package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	// DictDB 字典数据库
	DictDB = "dicts/dic.db"
)

var (
	r *rand.Rand
	// Nodes Dict Nodes
	Nodes *FileNode
)

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	Nodes = &FileNode{}
}

// StringStrip StringStrip
func StringStrip(s string, c string) string {
	for {
		s = strings.Replace(s, "//", "/", -1)
		b := true
		if strings.HasPrefix(s, c) {
			b = false
			s = s[len(c):]
		}
		if strings.HasSuffix(s, c) {
			b = false
			s = s[:len(s)-len(c)]
		}
		if b {
			break
		}
	}
	return s
}

// RandString 生成随机字符串
func RandString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

// RemoveHTML 移除 HTML 相关
func RemoveHTML(src string) string {
	//将HTML标签全转换成小写
	re, _ := regexp.Compile(`\<[\S\s]+?\>`)
	src = re.ReplaceAllStringFunc(src, strings.ToLower)
	//去除STYLE
	re, _ = regexp.Compile(`\<style[\S\s]+?\</style\>`)
	src = re.ReplaceAllString(src, "")
	//去除SCRIPT
	re, _ = regexp.Compile(`\<script[\S\s]+?\</script\>`)
	src = re.ReplaceAllString(src, "")
	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile(`\<[\S\s]+?\>`)
	src = re.ReplaceAllString(src, "\n")
	//去除连续的换行符
	re, _ = regexp.Compile(`\s{2,}`)
	src = re.ReplaceAllString(src, "\n")
	return src
}

// AddToNodes AddToNodes
func AddToNodes(node *FileNode, filepath string) {
	paths := strings.Split(filepath, "/")
	Debug.Println("AddToNodes paths : ", paths)
	for _, p := range paths {
		Debug.Println("AddToNodes -> p : ", p)
		if p == "" {
			continue
		}
		isDir := true
		// 特殊目录
		if p == paths[len(paths)-1] {
			if strings.Contains(p, ".") {
				isDir = false
			}
			for _, d := range strings.Split(SpecialFiles, ",") {
				if d == strings.ToLower(p) {
					Debug.Println("AddToNodes special files : ", d)
					isDir = false
				}
			}
			for _, d := range strings.Split(SpecialDirs, ",") {
				if d == strings.ToLower(p) {
					Debug.Println("AddToNodes special dir : ", d)
					isDir = true
				}
			}
		}
		// 判断路径尾部属否文件
		if !isDir {
			Debug.Println("AddToNodes addFile : ", p)
			node.addFile(p)
		} else {
			if _, ok := node.Nodes[p]; !ok {
				Debug.Println("AddToNodes addNode : ", p)
				node.addNode(p)
			} else {
				Debug.Println("AddToNodes getNode : ", p)
				node = node.getNode(p)
			}
		}
	}
}

// UpdateDicts 更新字典
func UpdateDicts(filePath string, dictDB string, jsonStr bool) {
	Debug.Println("Open", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		Error.Println("Open", filePath, "error")
		Error.Fatal(err)
	}
	defer file.Close()
	Debug.Println("Load DictDB : ", dictDB)
	Nodes.load(dictDB)
	Debug.Println("Parse Dict...")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineText := scanner.Text()
		if strings.HasPrefix(lineText, "/") {
			lineText = lineText[1:len(lineText)]
		}
		if strings.HasSuffix(lineText, "/") {
			lineText = lineText[:len(lineText)-1]
		}
		Debug.Println("lineText : ", lineText)
		AddToNodes(Nodes, lineText)
	}
	// PrintVar(nodes, 2)
	Debug.Println("Save DictDB : ", dictDB)
	Nodes.save(dictDB)

	if jsonStr {
		data, err := json.MarshalIndent(&Nodes, "", "\t")
		if err != nil {
			Error.Println("Marshal ResultUrls : ", err)
		}
		fmt.Println(string(data))
	}

}
