package main

import (
	"fmt"
	vidio "github.com/AlexEidt/Vidio"
	"io/fs"
	"log"
	"os"
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
	//base := "/home/fabien/Videos"
	base := "/media/fabien/exdata/A1_over60"
	result := CMap{value: make(map[uint][]string)}
	keys := CSet{value: make(map[uint]void)}

	ops := 0
	var wg sync.WaitGroup

	files, err := os.ReadDir("/home/fabien/Videos")
	if err != nil {
		log.Fatal(err)
	}

	filesSlices := chunkSlice(files, 4)

	for _, filesSlice := range filesSlices {
		for _, file := range filesSlice {
			ops++
			if !file.IsDir() {
				wg.Add(1)
				go read(base, file, ops, &wg, &result, &keys)
			}
		}
	}

	wg.Wait()

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
				err = os.Rename(f, destPath)
				if err != nil {
					log.Fatal(err)
				}
			}

		}
	}

	//fmt.Printf("result : %v", list)

	//fmt.Printf("result : %v", result.value)
	keys.RUnlock()
	result.RUnlock()
	//err := filepath.Walk("/home/fabien/Videos",
	//	func(path string, info os.FileInfo, err error) error {
	//	)}
	//if err != nil {
	//	log.Fatal(err)
	//}
}

func read(base string, file fs.DirEntry, ops int, wg *sync.WaitGroup, result *CMap, keys *CSet) {
	path := base + "/" + file.Name()
	fmt.Printf("%d, @@@ path : %s\n", ops, path)
	video, err := vidio.NewVideo(path)
	if err != nil {
		fmt.Printf("ERROR : %s", err)
	}

	duration := video.Duration()
	durationAsInt := uint(duration)
	fmt.Printf("lengh : %f or %d \n", duration, durationAsInt)

	keys.Lock()
	result.Lock()

	keys.value[durationAsInt] = member
	result.value[durationAsInt] = append(result.value[durationAsInt], path)

	result.Unlock()
	keys.Unlock()

	wg.Done()
}

func chunkSlice(slice []fs.DirEntry, chunkSize int) [][]fs.DirEntry {
	var chunks [][]fs.DirEntry
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
