package main

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/vanyason/infozone-gommaraizer/pkg/logger"
)

const (
	numRequests = 200
	execRepeats = 1
	externalUrl = "https://dog.ceo/api/breeds/image/random"
)

var (
	client     = &http.Client{}
	wg         sync.WaitGroup
	numWorkers = runtime.NumCPU() * 2
)

func execAndMeasureTime(f func()) {
	getFuncName := func(f func()) string {
		funcValue := reflect.ValueOf(f)
		if funcValue.Kind() != reflect.Func {
			return "not a function"
		}

		funcPointer := funcValue.Pointer()
		funcName := runtime.FuncForPC(funcPointer).Name()

		return funcName
	}

	funcName := getFuncName(f)
	start := time.Now()
	for i := 0; i < execRepeats; i++ {
		f()
	}
	logger.Info(funcName+" took", "time", time.Since(start)/time.Duration(execRepeats))
}

func fetchURL(client *http.Client, wg *sync.WaitGroup, id int) {
	if wg != nil {
		defer wg.Done()
	}
	resp, err := client.Get(externalUrl)
	if err != nil {
		fmt.Printf("Goroutine %d: Error fetching URL: %v\n", id, err)
		return
	}
	defer resp.Body.Close()
}

func TestUnlimitedGoroutines() {
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go fetchURL(client, &wg, i)
	}

	wg.Wait()
}

func TestWorkerPoolLimitedToCPU() {
	taskChan := make(chan int, numRequests)
	for i := 0; i < numRequests; i++ {
		taskChan <- i
	}
	close(taskChan) // Close the channel to signal workers there are no more tasks

	wg.Add(numWorkers)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for id := range taskChan {
				fetchURL(client, nil, id)
			}
		}()
	}

	// Add tasks to the channel
	wg.Wait() // Wait for all tasks to complete
}

func main() {
	execAndMeasureTime(TestUnlimitedGoroutines)
	execAndMeasureTime(TestWorkerPoolLimitedToCPU)
}
