package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	sortFlagName  = "sort"
	syncFlagName  = "sync"
	mergeFlagName = "merge"
)

func main() {
	joinCmd := flag.NewFlagSet(sortFlagName, flag.ExitOnError)
	// joinCmd.Usage = func() {
	// 	fmt.Fprintf(os.Stderr, "This is not helpful.\n")
	// }
	mergeCmd := flag.NewFlagSet(mergeFlagName, flag.ExitOnError)
	syncCmd := flag.NewFlagSet(syncFlagName, flag.ExitOnError)
	helper := fmt.Sprintf("expected on of subcommands: %s, %s, %s", syncFlagName, sortFlagName, mergeFlagName)

	if len(os.Args) < 2 {
		log.Println(helper)
		os.Exit(1)
	}
	switch os.Args[1] {

	case sortFlagName:

		destFolder := joinCmd.String("dst", "out", "destination folder")

		joinCmd.Parse(os.Args[2:])
		paths := joinCmd.Args()
		if len(paths) == 0 {
			log.Fatal("no paths specified")
		}
		log.Println("start")
		join(paths, *destFolder)
	case syncFlagName:
		syncCmd.Parse(os.Args[2:])
		paths := syncCmd.Args()
		if len(paths) == 0 {
			log.Fatal("no paths specified")
		}
		log.Println("start")
		sync(paths)
	case mergeFlagName:
		srcFolder := mergeCmd.String("src", "out", "destination folder")
		mergeCmd.Parse(os.Args[2:])
		log.Println("start")
		merge(*srcFolder)
		log.Println("subcommand 'merge'")
	default:
		log.Println(helper)
		os.Exit(1)
	}
	log.Println("finish")
}

func join(paths []string, destFolder string) {
	allFiles := []string{}
	if destFolder == "out" && len(paths) == 1 {
		destFolder = paths[0] + "-sorted"
	}
	_, err := os.Stat(destFolder)
	if err == nil {
		log.Fatal("dst dir is existing: ", destFolder)
	}
	log.Printf("Current dir: %v, dest: %s", paths, destFolder)
	for _, src := range paths {
		subFiles := getPaths(src)
		log.Println("src: ", src, " len: ", len(subFiles))
		allFiles = append(allFiles, subFiles...)
	}

	for _, format := range []string{"ARW", "JPG", "MP4"} {
		log.Println("Processing: ", format)
		for _, path := range allFiles {
			dstDir := filepath.Join(destFolder, format)
			os.MkdirAll(dstDir, os.ModePerm)
			if strings.Contains(path, format) && !strings.Contains(path, "THMBNL") {
				data, err := os.ReadFile(path)
				if err != nil {
					log.Fatal(err)
				}
				err = os.WriteFile(filepath.Join(dstDir, filepath.Base(path)), data, os.ModePerm)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
	log.Println("len of files:", len(allFiles))
}

func merge(src string) {
	allFiles := getPaths(src)
	namesMap := map[string]string{}
	for _, path := range allFiles {
		baseName := filepath.Base(path)
		name := strings.Split(baseName, ".")[0]
		p, ok := namesMap[name]
		if ok && filepath.Ext(p) == ".ARW" {
			continue
		}
		namesMap[name] = path
	}
	for _, path := range namesMap {
		dstDir := filepath.Join(src, "MERGED")
		os.MkdirAll(dstDir, os.ModePerm)
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(filepath.Join(dstDir, filepath.Base(path)), data, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getPaths(src string) []string {
	allFiles := []string{}
	filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if strings.Contains(path, ".JPG") || strings.Contains(path, ".ARW") || strings.Contains(path, ".MP4") {
			allFiles = append(allFiles, path)
		}
		return nil
	})
	return allFiles
}

func sync(paths []string) {
	for _, path := range paths {
		allJPGFiles := getPaths(filepath.Join(path, "JPG"))
		allRAWFiles := getPaths(filepath.Join(path, "ARW"))
		log.Println("allRAWFiles:", len(allRAWFiles))
		dstDir := filepath.Join(path, "ARW_FILTERED")
		_, err := os.Stat(dstDir)
		if err == nil {
			log.Fatal("dst dir is existing: ", dstDir)
		}
		os.MkdirAll(dstDir, os.ModePerm)
		for _, jpg := range allJPGFiles {
			jpgName := strings.Split(filepath.Base(jpg), ".")[0]
			rawExist := false
			rawFile := ""
			for _, raw := range allRAWFiles {
				rawName := strings.Split(filepath.Base(raw), ".")[0]
				if rawName == jpgName {
					rawExist = true
					rawFile = raw
				}
			}
			if rawExist {
				data, err := os.ReadFile(rawFile)
				if err != nil {
					log.Fatal(err)
				}
				err = os.WriteFile(filepath.Join(dstDir, filepath.Base(rawFile)), data, os.ModePerm)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
