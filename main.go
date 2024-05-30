package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	code := run(os.Stdout, os.Args)
	if code != 0 {
		os.Exit(code)
	}
}

const usageText = `usage: sony <command> [<args>...]

sony - manage sony media through CLI

commands:
  sort       Sorts standard sony media file structure to ARW, JPG and MP4 subfolders
  sync       Syncs ARW and JPG folders
  merge      Creates new folder with uniques JPG and ARW files
`

func run(w io.Writer, args []string) (code int) {
	if len(args) < 2 {
		fmt.Fprint(w, usageText)
		return 0
	}
	switch args[1] {
	case "sync":
		return syncCmd(w, args[2:])
	case "merge":
		return mergeCmd(w, args[2:])
	case "sort":
		return sortCmd(w, args[2:])
	}
	fmt.Fprint(w, usageText)
	return 0
}

const syncUsageText = `usage: sony sync [<args>...] PATHS

Syncs JPG and RAW folders. Deletes files from RAW if same JPG doesn't exist. 

Helpful in case if you want to delete unnecessary JPG files and corresponding RAW's.

Args:
  -v
    Set log verbosity level to "debug". (default "info")
  -help
    Print help and exit.
`

func syncCmd(w io.Writer, args []string) (code int) {
	cmd := flag.NewFlagSet("sync", flag.ExitOnError)
	cmd.SetOutput(w)
	helpFlag := cmd.Bool("help", false, "")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		fmt.Fprint(w, syncUsageText)
		return
	}
	paths := cmd.Args()
	if len(paths) == 0 {
		fmt.Fprint(w, syncUsageText)
		return
	}
	if err := sync(w, paths); err != nil {
		fmt.Fprintln(w, "Command failed: "+err.Error())
		return 1
	}
	return 0
}

func sync(w io.Writer, paths []string) error {
	log := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{}))

	for _, path := range paths {
		allJPGFiles := getPaths(filepath.Join(path, "JPG"))
		allRAWFiles := getPaths(filepath.Join(path, "ARW"))
		log.Info("allRAWFiles", "number", len(allRAWFiles))
		dstDir := filepath.Join(path, "ARW_FILTERED")
		_, err := os.Stat(dstDir)
		if err == nil {
			return fmt.Errorf("dst dir is existing: %s", dstDir)
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
					return err
				}
				err = os.WriteFile(filepath.Join(dstDir, filepath.Base(rawFile)), data, os.ModePerm)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

const mergeUsageText = `usage: sony merge [<args>...]

Create new MERGED folder that will contain all unique RAW and JPG files. RAW is prefered over JPG

In case if some of the pictures were taken only in JPG. But all the other pictures are in JPEG and RAW. Then you can merge.

Args:
  -src
    Path to look for JPG and RAW folders. (default "out")
  -help
    Print help and exit.
`

func mergeCmd(w io.Writer, args []string) (code int) {
	cmd := flag.NewFlagSet("merge", flag.ExitOnError)
	cmd.SetOutput(w)

	src := cmd.String("src", "out", "folder")
	helpFlag := cmd.Bool("help", false, "")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		fmt.Fprint(w, syncUsageText)
		return
	}
	paths := cmd.Args()
	if len(paths) == 0 {
		fmt.Fprint(w, syncUsageText)
		return
	}
	if err := merge(w, *src); err != nil {
		fmt.Fprintln(w, "Command failed: "+err.Error())
		return 1
	}
	return 0
}

func merge(_ io.Writer, src string) error {
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
			return err
		}
		err = os.WriteFile(filepath.Join(dstDir, filepath.Base(path)), data, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

const sortUsageText = `usage: sony sort [<args>...]


Args:
  -dst
    Path to look for JPG and RAW folders. (default "out")
  -help
    Print help and exit.
`

func sortCmd(w io.Writer, args []string) (code int) {
	cmd := flag.NewFlagSet("sort", flag.ExitOnError)
	cmd.SetOutput(w)

	dst := cmd.String("dst", "out", "folder")
	helpFlag := cmd.Bool("help", false, "")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		fmt.Fprint(w, syncUsageText)
		return
	}
	paths := cmd.Args()
	if len(paths) == 0 {
		fmt.Fprint(w, syncUsageText)
		return
	}
	if err := sort(w, paths, *dst); err != nil {
		fmt.Fprintln(w, "Command failed: "+err.Error())
		return 1
	}
	return 0
}

func sort(w io.Writer, paths []string, destFolder string) error {
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
	return nil
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
