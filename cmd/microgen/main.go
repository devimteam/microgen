package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	flagFileName  = flag.String("f", "", "File name")
	flagIfaceName = flag.String("i", "", "Interface name")
)

func init() {
	flag.Parse()
}

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(currentDir, *flagFileName)
	fmt.Println(path)
}
