package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type result struct {
	url    string
	exists bool
}

func checkIfExists(done <-chan struct{}, urls <-chan string) <-chan result {
	// fmt.Println("checkIfExists")
	resultc := make(chan result)
	go func() {
		defer close(resultc)
		for {
			select {
			case <-done:
				return
			case url, ok := <-urls:
				if !ok {
					return
				}
				res, err := http.Get(url)
				if err != nil {
					resultc <- result{url: url, exists: false}
				} else if res.StatusCode == http.StatusOK {
					resultc <- result{url: url, exists: true}
				} else {
					resultc <- result{url: url, exists: false}
				}
			}
		}
	}()
	return resultc
}
func merge[T any](done <-chan struct{}, channels ...<-chan T) <-chan T {
	results := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(channels))
	for _, c := range channels {
		go func(c <-chan T) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				case i, ok := <-c:
					if !ok {
						return
					}
					results <- i
				}
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}

func main() {
	done := make(chan struct{})
	defer close(done)

	urls := make(chan string, 4)
	urls <- "https://google.com"
	urls <- "https://amazon.com"
	urls <- "https://in-valid-url.invalid"
	urls <- "https://facebook.com"
	close(urls)
	c1 := checkIfExists(done, urls)
	c2 := checkIfExists(done, urls)
	c3 := checkIfExists(done, urls)
	now := time.Now()
	for rs := range merge(done, c1, c2, c3) {
		fmt.Printf("url: %v, exists: %v\n", rs.url, rs.exists)
	}
	fmt.Println(time.Since(now))
}
