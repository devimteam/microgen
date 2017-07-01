package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"github.com/microgen/util"
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

	fs, err := util.ParseInterface(path, *flagIfaceName)
	fmt.Println(*fs[0].Params[2], err)
}
