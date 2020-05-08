package main

import (
	"flag"
	"github.com/worldiety/reflectplus"
)

func main() {
	dir := flag.String("dir", ".", "the directory to scan")
	flag.Parse()

	reflectplus.Must(reflectplus.Generate(*dir))
}
