package utils

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"runtime"
)

// OptimalGoroutines returns the optimal number of goroutines to use based on the nature of the tasks.
func OptimalGoroutines(isIOBound bool) int {
	numCPU := runtime.NumCPU()

	if isIOBound {
		// For I/O-bound tasks, you can afford to have more goroutines than CPU cores.
		// A good heuristic is to multiply the number of CPU cores by 2 or more.
		return numCPU * 2
	}

	// For CPU-bound tasks, it's best to match the number of goroutines to the number of CPU cores.
	return numCPU
}

// CopyJar returns a new cookie jar with the same cookies as the given jar.
// This is useful for concurrent use of the same cookie jar, since the
// standard library's cookie jar is not safe for concurrent use.
//
// The new cookie jar is created with the same policy as the original jar.
//
// The cookies are copied from the original jar to the new jar using the
// SetCookies method. This means that any changes made to the original jar
// after calling CopyJar will not be reflected in the new jar.
func CopyJar(jar *cookiejar.Jar, u *url.URL) (*cookiejar.Jar, error) { //< copy jar for concurrent use
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
