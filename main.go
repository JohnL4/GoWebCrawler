package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type FetchResult struct {
	body string
	err  error
}

// var __urls SyncMap[string,bool] = SyncMap[string,bool]{ make(map[string]bool)}

// Map from url to the body that url returns.
var __urls = &SyncMap[string, FetchResult]{make(map[string]FetchResult), sync.Mutex{}}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	defer waitGrp.Done()

	fmt.Println("__urls has", len(__urls._map), "elements")

	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		return
	}
	if _, err := __urls.Get(url); err == nil {
		// Already fetched
		return
	}
	body, urls, err := fetcher.Fetch(url)
	__urls.Put(url, FetchResult{body, err})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		waitGrp.Add(1)
		go Crawl(u, depth-1, fetcher)
	}
	return
}

//----------------------------------------------------------------------------------------------------------------------

// Synchronized map
type SyncMap[T comparable, U any] struct {
	_map   map[T]U
	_mutex sync.Mutex // Prolly should use something like RWMutex?  Eh.  Another day.
}

func (sm *SyncMap[T, U]) Put(key T, val U) {
	sm._mutex.Lock()         // acquire lock
	defer sm._mutex.Unlock() // release lock at end, guaranteed
	sm._map[key] = val       // store value
}

func (sm *SyncMap[T, U]) Get(key T) (U, error) {
	sm._mutex.Lock()
	defer sm._mutex.Unlock()
	if v, ok := sm._map[key]; ok {
		return v, nil
	} else {
		return v, fmt.Errorf("key not found: %v", key)
	}
}

var waitGrp sync.WaitGroup

// =====================================================================================================================
func main() {
	waitGrp.Add(1)
	Crawl("https://golang.org/", 4, fetcher)
	waitGrp.Wait()
	fmt.Println("Final results:")
	for key, entry := range __urls._map {
		if entry.err == nil {
			fmt.Printf("%v --> %v\n", key, entry.body)
		} else {
			fmt.Printf("%v --> ERROR: %v\n", key, entry.err)
		}
	}
}

// =====================================================================================================================

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
