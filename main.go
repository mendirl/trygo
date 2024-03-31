package main

import (
	"fmt"
	vidio "github.com/AlexEidt/Vidio"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

type CMap = struct {
	sync.RWMutex
	value map[uint][]string
}
type CSet = struct {
	sync.RWMutex
	value map[uint]void
}

type void struct{}

var member void

func main() {
	base := "/media/fabien/exdata/A1_over60"
	//dest := "/media/fabien/exdata/F"
	result := CMap{value: make(map[uint][]string)}
	keys := CSet{value: make(map[uint]void)}
	ops := 0
	var wg sync.WaitGroup

	files := listFiles(base)
	filesSlices := chunkSlice(files, 4)
	for _, filesSlice := range filesSlices {
		ops++
		wg.Add(1)
		go reads(filesSlice, ops, &wg, &result, &keys)
	}
	wg.Wait()

	move(&result, &keys, base)
}

func move(result *CMap, keys *CSet, base string) {
	result.RLock()
	keys.RLock()

	list := make([]uint, 0, len(keys.value))
	for k := range keys.value {
		list = append(list, k)
	}

	slices.Sort(list)

	for _, k := range list {
		i := len(result.value[k])
		//
		if i > 1 {
			fmt.Printf("%d :\n", k)
			for _, f := range result.value[k] {
				fmt.Printf("%v\n", f)
				split := strings.Split(f, "/")
				destPath := base + "/verify/" + split[len(split)-1]
				fmt.Printf("will mote to %v\n", destPath)
				err := os.Rename(f, destPath)
				if err != nil {
					log.Fatal(err)
				}
			}

		}
	}
	keys.RUnlock()
	result.RUnlock()
}
func HandlePanic() {
	r := recover()

	if r != nil {
		fmt.Println("RECOVER", r)
	}
}
func reads(files []string, ops int, wg *sync.WaitGroup, result *CMap, keys *CSet) {
	for _, file := range files {
		read(file, ops, wg, result, keys)
	}
}
func read(file string, ops int, wg *sync.WaitGroup, result *CMap, keys *CSet) {
	defer wg.Done()

	path := file
	fmt.Printf("%d, @@@ path : %s\n", ops, path)
	video, err := vidio.NewVideo(path)
	if err != nil {
		fmt.Printf("ERROR : %s", err)
	}

	defer HandlePanic()
	duration := video.Duration()
	durationAsInt := uint(duration)
	fmt.Printf("lengh : %f or %d \n", duration, durationAsInt)

	keys.Lock()
	result.Lock()

	keys.value[durationAsInt] = member
	result.value[durationAsInt] = append(result.value[durationAsInt], path)

	result.Unlock()
	keys.Unlock()

}

func chunkSlice(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}
func listFiles(base string) []string {
	files := make([]string, 0)

	err := filepath.Walk(base,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			fmt.Println(path, info.Size())
			if !info.IsDir() && strings.HasSuffix(path, ".mp4") {
				files = append(files, path)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	return files
}
