package main

import (
	"os/exec"
	"fmt"
	"os"
)

const (
	dirPath     = "/data/"
)

func main() {
	var cmd *exec.Cmd
	filepath := dirPath + "/admin12160c174cbc126e1a/"
	filestore := dirPath + "go1.14.1.linux-amd64.tar.gz"

	cmd = exec.Command("/home/steve/go/src/github.com/HashCell/go-fileserver/script/shell/merge_file_blocks.sh", filepath, filestore)
	// cmd.Run()
	if _, err := cmd.Output(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println(filestore, " has been merge complete")
	}
}