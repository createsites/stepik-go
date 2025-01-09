package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	for rawData := range in {
		data, ok := rawData.(int)
		if !ok {
			panic(fmt.Sprintf("type of the input value (%#v) is not int", rawData))
		}
		// crc32(data)+"~"+crc32(md5(data))
		out <- DataSignerCrc32(strconv.Itoa(data)) + "~" + DataSignerCrc32(DataSignerMd5(strconv.Itoa(data)))
	}
}

func MultiHash(in, out chan interface{}) {
	for rawData := range in {
		data, ok := rawData.(string)
		if !ok {
			panic(fmt.Sprintf("type of the input value (%#v) is not string", rawData))
		}
		results := ""
		for i := 0; i < 6; i++ {
			results += DataSignerCrc32(strconv.Itoa(i) + data)
		}
		out <- results
	}
}

func CombineResults(in, out chan interface{}) {
	results := []string{}
	for v := range in {
		s, ok := v.(string)
		if !ok {
			panic(fmt.Sprintf("type of the input value (%#v) is not int", v))
		}
		results = append(results, s)
	}
	sort.Strings(results)
	out <- strings.Join(results, "_")
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
