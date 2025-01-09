package main

import (
	"fmt"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	fmt.Printf("SingleHash reads value %#v\n", <-out)

}

func ExecutePipeline(hashSignJobs ...job) {
	in := make(chan interface{})
	out := make(chan interface{})

	wg := &sync.WaitGroup{}

	for i, j := range hashSignJobs {
		wg.Add(1)
		go func(in, out chan interface{}, j job, i int) {
			defer wg.Done()
			j(in, out)
			close(out)
		}(in, out, j, i)
		in = out
		out = make(chan interface{})
	}

	wg.Wait()
}
