package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

var ignor = ".DS_Store"

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}

}
func dirTree(out io.Writer, path string, key bool) error {
	pref := ""
	if key{
		walkDirTree(out, path, key, pref)
	}else{
		walkPartTree(out, path, key, pref)
	}
	return nil
}

func walkPartTree(out io.Writer, path string, key bool, pref string) error{
	var files []string
	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, file := range fileInfo{
		if file.IsDir(){
			files = append(files, file.Name())
		}
	}
	for _, file := range files {
		if file != files[len(files)-1] {
			fmt.Fprintln(out, pref + "├───" + file)
			walkDirTree(out, path + "/" + file, key, pref + "│" + "\t")
		}else{
			fmt.Fprintln(out, pref + "└───" + file)
			path += "/" + file
			walkDirTree(out, path, key, pref+"\t")
		}
	}
	return nil
}

func walkDirTree(out io.Writer, path string, key bool, pref string) error {
	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, file := range fileInfo {
		if file.Name() != fileInfo[len(fileInfo)-1].Name() {
			if file.IsDir(){
				fmt.Fprintln(out, pref + "├───" + file.Name())
				walkDirTree(out, path + "/" + file.Name(), key, pref + "│" + "\t")
			} else if file.Name() != ignor && key{
				if file.Size() > 0 {
					fmt.Fprintln(out, pref + "├───" + file.Name() + " (" + strconv.FormatInt(int64(file.Size()), 10) + "b)") // + file.Size() + "b)")// + " (" + file.Size() + "b)")
				}else{
					fmt.Fprintln(out, pref + "├───" + file.Name() + " (empty)")
				}
			}
		}else{
			if file.IsDir(){
				fmt.Fprintln(out, pref + "└───" + file.Name())
				path += "/" + file.Name()
				walkDirTree(out, path, key, pref + "\t")
			} else if file.Name() != ignor && key{
				if file.Size()>0{
					fmt.Fprintln(out, pref + "└───" + file.Name() + " (" + strconv.FormatInt(int64(file.Size()), 10) + "b)")
				}else{
					fmt.Fprintln(out, pref + "└───" + file.Name() + " (empty)")
				}
			}
		}
	}
	return nil
}