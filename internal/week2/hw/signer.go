package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	// канал квоты для функции DataSignerMd5, чтобы избежать перегрева
	quotaCh := make(chan struct{}, 1)

	wg := &sync.WaitGroup{}
	for rawData := range in {
		data, ok := rawData.(int)
		if !ok {
			panic(fmt.Sprintf("type of the input value (%#v) is not int", rawData))
		}
		wg.Add(1)
		go func(out chan interface{}) {
			defer wg.Done()
			// crc32(data)+"~"+crc32(md5(data))
			quotaCh <- struct{}{}
			md5 := DataSignerMd5(strconv.Itoa(data))
			<- quotaCh
			out <- DataSignerCrc32(strconv.Itoa(data)) + "~" + DataSignerCrc32(md5)
		}(out)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	for rawData := range in {
		data, ok := rawData.(string)
		if !ok {
			panic(fmt.Sprintf("type of the input value (%#v) is not string", rawData))
		}
		wg := &sync.WaitGroup{}
		results := make([]string, 6)
		for i := 0; i < 6; i++ {
			wg.Add(1)
			go func(i int, data string) {
				defer wg.Done()
				results[i] = DataSignerCrc32(strconv.Itoa(i) + data)
			}(i, data)
		}
		wg.Wait()
		out <- strings.Join(results, "")
	}
}

func CombineResults(in, out chan interface{}) {
	results := []string{}
	for v := range in {
		s, ok := v.(string)
		if !ok {
			panic(fmt.Sprintf("type of the input value (%#v) is not string", v))
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
