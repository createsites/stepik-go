package main

import (
	"fmt"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	fmt.Printf("SingleHash reads value %#v\n", <-out)

}

func ExecutePipeline(hashSignJobs ...job) {

	defer fmt.Printf("STOP of ExecutePipeline\n")

	var in, out chan interface{}

	wg := &sync.WaitGroup{}

	in = make(chan interface{})

	out = make(chan interface{})
	wg.Add(1)
	go func(in, out chan interface{}) {
		defer wg.Done()
		hashSignJobs[0](in, out)
	}(in, out)
	in = out
	out = make(chan interface{})
	go func(in, out chan interface{}) {
		defer wg.Done()
		hashSignJobs[1](in, out)
	}(in, out)
	in = out
	out = make(chan interface{})
	go func(in, out chan interface{}) {
		defer wg.Done()
		hashSignJobs[2](in, out)
	}(in, out)

	fmt.Printf("TEST EXECUTING\n")

	wg.Wait()
}
