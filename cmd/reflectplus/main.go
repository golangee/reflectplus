// Copyright 2020 Torben Schinke
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"github.com/golangee/reflectplus"
	"github.com/golangee/reflectplus/golang"
	"log"
	"strings"
)

func main() {
	dir := flag.String("dir", "", "the directory to scan")
	patterns := flag.String("patterns", "", "the path patterns to parse, e.g. github.com/myproject/mypath/...;github.com/other/path/...")
	help := flag.Bool("help", false, "shows this help.")
	flag.Parse()

	if *help {
		fmt.Println("reflectplus parses the go code at your fingertips and represents a subset of it in json form.")
		flag.PrintDefaults()
		return
	}

	var prj *golang.Project
	var err error

	if *dir == "" && *patterns == "" {
		prj, err = reflectplus.ParseModule()
	} else {
		prj, err = reflectplus.Parse(golang.Options{
			Dir:      *dir,
			Patterns: strings.Split(*patterns, ";"),
		})
	}

	if err != nil{
		log.Fatal(err)
	}

	fmt.Println(prj.String())
}
