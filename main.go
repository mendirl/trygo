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

type CStringList = struct {
	sync.RWMutex
	value []string
}

type Video = struct {
	name string
	path string
	size int64
	duration uint
}

type void struct{}

var member void

func main() {
	//	base := "/media/fabien/exdata/A1_over60"
	//  base := "/home/fabien/Videos"
	//	base := "/run/media/fabien/exdata/O"
	//	base := "/run/media/fabien/data/O"
	dest := "/mnt/share/misc/P/"

	bases := []string{"/home/fabien/Videos", "/run/media/fabien/data"/*, "/mnt/share/misc/P/O"*/}

	process(bases, dest)

	fmt.Printf("#### C'est fini #####")
}

func process(bases []string, dest string) {
	result := CMap{value: make(map[uint][]string)}
	keys := CSet{value: make(map[uint]void)}
	files := CStringList{value: make([]string, 0)}

	ops := 0
	var wg sync.WaitGroup

	// list all files present in folders
	for _, base := range bases {
		wg.Add(1)
		go listFiles(base, &files, &wg)
	}
	wg.Wait()

	// split this list into chuncks to parallize computation
	filesSlices := chunkSlice(files.value, 50)

	// parallize treatment for each chunck
	for _, filesSlice := range filesSlices {
		ops++
		wg.Add(1)
		go reads(filesSlice, ops, &wg, &result, &keys)
	}
	wg.Wait()

	move(&result, &keys, dest)
}

func move(result *CMap, keys *CSet, base string) {
	result.RLock()
	keys.RLock()

	durations := make([]uint, 0, len(keys.value))
	for k := range keys.value {
		durations = append(durations, k)
	}

	slices.Sort(durations)

	for _, duration := range durations {
		nb := len(result.value[duration])
		if nb > 1 {
			fmt.Printf("duration %d has multiples files %d :\n", duration, nb)
			for _, path := range result.value[duration] {
				fmt.Printf("%v\n", path)
				split := strings.Split(path, "/")
				destPath := base + "/verify/" + split[len(split)-1]
				fmt.Printf("will move to %v\n", destPath)
	//				err := os.Rename(path, destPath)
	//				if err != nil {
	//					log.Fatal(err)
	//				}
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

// for each file, compute its size as int
// and group them by the size
func reads(files []string, ops int, wg *sync.WaitGroup, result *CMap, keys *CSet) {
	for _, file := range files {
		video := createVideo(file)
		read(video, ops, result, keys)
	}
	defer wg.Done()
}

func read(video Video, ops int, result *CMap, keys *CSet) {
	keys.Lock()
	result.Lock()

	fmt.Printf("#%d - %s, %db, %ds\n", ops, video.path, video.size, video.duration)
	keys.value[video.duration] = member
	result.value[video.duration] = append(result.value[video.duration], video.path)

	result.Unlock()
	keys.Unlock()
}

func chunkSlice(files []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(files); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond files capacity
		if end > len(files) {
			end = len(files)
		}

		chunks = append(chunks, files[i:end])
	}

	return chunks
}

func listFiles(base string, files *CStringList, wg *sync.WaitGroup) {
	defer HandlePanic()
	defer wg.Done()

	err := filepath.Walk(base,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".mp4") {
				files.Lock()
				files.value = append(files.value, path)
				files.Unlock()
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

func createVideo(path string) Video {
	video, err := vidio.NewVideo(path)
	info, err := os.Stat(path)

	if err != nil {
		fmt.Printf("ERROR : %s", err)
	}

	duration := uint(video.Duration())


	return  Video{path, path, info.Size(), duration}
}
