package utils

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"runtime"
)

// OptimalGoroutines returns the optimal number of goroutines to use based on the nature of the tasks.
// isIOBound should be set to true for I/O-bound tasks, and false for CPU-bound tasks (in that case amount of CPUs is x2).
func OptimalGoroutines(isIOBound bool) int {
	numCPU := runtime.NumCPU()

	if isIOBound {
		return numCPU * 2
	}

	return numCPU
}

// CopyJar returns a new cookie jar with the same cookies as the given jar.
func CopyJar(jar *cookiejar.Jar, u *url.URL) (*cookiejar.Jar, error) {
	if jar == nil {
		return nil, fmt.Errorf("nil cookie jar")
	}

	newJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	for _, cookie := range jar.Cookies(u) {
		newJar.SetCookies(u, []*http.Cookie{cookie})
	}

	return newJar, nil
}
