// Purpose of this file is to test what is better :
// 1000 goroutines for parsing 1000 files or worker pool with 32 goroutines doing the same job
// Program reads all the files, counts the unique symbols, prints out the result

package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/trace"
	"sync"

	"github.com/vanyason/infozone-gommaraizer/pkg/utils"
)

const (
	filesCount = 1000
	folderPath = "/tmp/goroutines_test"
	charset    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLen = len(charset)
)

// GenerateRandomString returns a byte slice containing a random string of
// 50 characters. The string is composed of characters from the charset
// consisting of lowercase and uppercase letters and digits.
func generateRandomString() []byte {
	b := make([]byte, 50)
	for i := range b {
		b[i] = charset[rand.Intn(charsetLen)]
	}
	return b
}

func printResult(m map[byte]uint) {
	fmt.Println("Character counts:")
	for _, c := range []byte(charset) {
		fmt.Printf("'%c': %d\n", c, m[c])
	}
}

// CreateTestData creates folder at folderPath and 1000 files inside
// Files are named 0, 1, 2, ..., 999
// Content of each file is random string of 50 bytes
func CreateTestData() error {
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return fmt.Errorf("error creating folder: %w", err)
	}

	for i := 0; i < filesCount; i++ {
		filePath := filepath.Join(folderPath, fmt.Sprintf("%d", i))
		if err := os.WriteFile(filePath, generateRandomString(), os.ModePerm); err != nil {
			return fmt.Errorf("error writing file %s: %w", filePath, err)
		}
	}

	return nil
}

// Deletes folder at folderPath
func Cleanup() {
	if err := os.RemoveAll(folderPath); err != nil {
		log.Println("Error cleaning up folder:", err)
	}
}

// Straightforward logic : open file, read it, count unique symbols
func TestSingleThread() {
	unique := map[byte]uint{}

	for i := 0; i < filesCount; i++ {
		data, err := os.ReadFile(filepath.Join(folderPath, fmt.Sprintf("%d", i)))
		if err != nil {
			log.Printf("error reading file %s: %v", fmt.Sprintf("%d", i), err)
			continue
		}

		for i := range data {
			unique[data[i]]++
		}
	}

	printResult(unique)
}

// Give each file to a goroutine
func Test1000() {
	var mu sync.Mutex
	var wg sync.WaitGroup
	unique := map[byte]uint{}

	wg.Add(filesCount)
	for i := 0; i < filesCount; i++ {
		go func(i int) {
			defer wg.Done()

			data, err := os.ReadFile(filepath.Join(folderPath, fmt.Sprintf("%d", i)))
			if err != nil {
				log.Printf("error reading file %s: %v", fmt.Sprintf("%d", i), err)
				return
			}

			mu.Lock()
			defer mu.Unlock()
			for i := range data {
				unique[data[i]]++
			}
		}(i)
	}

	wg.Wait()
	printResult(unique)
}

// Create worker pool
func TestWP() {
	var mu sync.Mutex
	var wg sync.WaitGroup
	workers := utils.OptimalGoroutines(false)
	jobsCh := make(chan int, filesCount)
	unique := map[byte]uint{}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := range jobsCh {
				data, err := os.ReadFile(filepath.Join(folderPath, fmt.Sprintf("%d", j)))
				if err != nil {
					log.Printf("error reading file %s: %v", fmt.Sprintf("%d", i), err)
					continue
				}

				mu.Lock()
				for i := range data {
					unique[data[i]]++
				}
				mu.Unlock()
			}
		}()
	}

	for i := 0; i < filesCount; i++ {
		jobsCh <- i
	}

	close(jobsCh)
	wg.Wait()
	printResult(unique)
}

func init() {
	if err := CreateTestData(); err != nil {
		log.Fatal(err)
	}
}

// CMD arguments:
// init 	 - call CreateTestData
// cleanup   - call Cleanup
// single	 - call TestSingleThread
// 1000  	 - call Test1000
// wp 		 - call TestWP
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <init|cleanup|single|1000|wl>")
		return
	}

	traceFile, err := os.Create("/tmp/trace.out")
	if err != nil {
		log.Fatalf("could not create trace file: %v", err)
	}
	defer traceFile.Close()

	if err := trace.Start(traceFile); err != nil {
		log.Fatalf("could not start trace: %v", err)
	}
	defer trace.Stop()

	switch os.Args[1] {
	case "init":
		if err := CreateTestData(); err != nil {
			log.Fatal(err)
		}
	case "cleanup":
		Cleanup()
	case "single":
		TestSingleThread()
	case "1000":
		Test1000()
	case "wp":
		TestWP()
	default:
		fmt.Println("Invalid argument. Use: <init|cleanup|single|1000|wl>")
	}
}
