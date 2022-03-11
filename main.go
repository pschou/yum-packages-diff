// Written by Paul Schou (paulschou.com) March 2022
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
	"io"
	"log"
	"os"
	"path"
	"strings"
)

var version = "test"
var repoPath string

// HelloGet is an HTTP Cloud Function.
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Yum Package Diff,  Version: %s\n\nUsage: %s [options...]\n\n", version, os.Args[0])
		flag.PrintDefaults()
	}

	var newFile = flag.String("new", "NewPrimary.xml.gz", "Package list for comparison")
	var oldFile = flag.String("old", "OldPrimary.xml.gz", "Package list for comparison")
	var inRepoPath = flag.String("repo", "/7/os/x86_64", "Repo path to use in file list")
	var outputFile = flag.String("output", "-", "Output for comparison result")
	var showNew = flag.Bool("showAdded", false, "Display packages only in the new list")
	var showOld = flag.Bool("showRemoved", false, "Display packages only in the old list")
	var showCommon = flag.Bool("showCommon", false, "Display packages in both the new and old lists")
	flag.Parse()

	newPackages := readFile(*newFile)
	oldPackages := readFile(*oldFile)
	repoPath = strings.TrimSuffix(strings.TrimPrefix(*inRepoPath, "/"), "/")

	out := os.Stdout
	if *outputFile != "-" {
		f, err := os.Create(*outputFile)
		check(err)
		defer f.Close()
		out = f
	}

	// initialized with zeros
	newMatched := make([]int8, len(newPackages))
	oldMatched := make([]int8, len(oldPackages))

	log.Println("doing matchups")
matchups:
	for iNew, pNew := range newPackages {
		for iOld, pOld := range oldPackages {
			//if reflect.DeepEqual(pNew, pOld) {
			if pNew.Checksum.Text == pOld.Checksum.Text &&
				pNew.Checksum.Type == pOld.Checksum.Type &&
				pNew.Size.Package == pOld.Size.Package &&
				pNew.Location.Href == pOld.Location.Href {
				newMatched[iNew] = 1
				oldMatched[iOld] = 1
				continue matchups
			}
		}
	}

	fmt.Fprintln(out, "# Yum-diff matchup, version:", version)
	fmt.Fprintln(out, "# new:", *newFile, "old:", *oldFile)

	if *showNew {
		for iNew, pNew := range newPackages {
			if newMatched[iNew] == 0 {
				// This package was not seen in OLD
				printPackage(out, pNew)
			}
		}
	}

	if *showCommon {
		for iNew, pNew := range newPackages {
			if newMatched[iNew] == 1 {
				// This package was seen in BOTH
				printPackage(out, pNew)
			}
		}
	}

	if *showOld {
		for iOld, pOld := range oldPackages {
			if oldMatched[iOld] == 0 {
				// This package was not seen in NEW
				printPackage(out, pOld)
			}
		}
	}
}

func printPackage(out io.Writer, p Package) {
	fmt.Fprintf(out, "{%s}%s %s %s\n", p.Checksum.Type, p.Checksum.Text, p.Size.Package, path.Join(repoPath, p.Location.Href))
}

func check(e error) {
	if e != nil {
		//panic(e)
		log.Fatal(e)
	}
}
