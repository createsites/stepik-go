package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// возвращает строку вида
// crc32(data)+"~"+crc32(md5(data))
func SingleHash(in, out chan interface{}) {
	// канал квоты для функции DataSignerMd5, чтобы избежать перегрева
	// т.к. функция DataSignerMd5 не может быть распараллелена
	md5QuotaCh := make(chan struct{}, 1)

	wg := &sync.WaitGroup{}
	for rawData := range in {
		data, ok := rawData.(int)
		if !ok {
			panic(fmt.Sprintf("type of the input value (%#v) is not int", rawData))
		}
		wg.Add(1)
		go func(out chan interface{}) {
			defer wg.Done()
			// md5
			// забираем квоту, записывая токен в канал
			md5QuotaCh <- struct{}{}
			// считаем хеш
			md5 := DataSignerMd5(strconv.Itoa(data))
			// читаем наш токен, возвращая квоту для других горутин
			<-md5QuotaCh

			// массив для результатов распараллеливания crc32 функции
			crc := [2]string{}
			wg2 := &sync.WaitGroup{}
			wg2.Add(1)
			go func(data int) {
				defer wg2.Done()
				crc[0] = DataSignerCrc32(strconv.Itoa(data))
			}(data)
			wg2.Add(1)
			go func(data int) {
				defer wg2.Done()
				crc[1] = DataSignerCrc32(md5)
			}(data)
			wg2.Wait()

			out <- crc[0] + "~" + crc[1]
		}(out)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for rawData := range in {
		data, ok := rawData.(string)
		if !ok {
			panic(fmt.Sprintf("type of the input value (%#v) is not string", rawData))
		}
		// поступившие в канал ззначения распараллеливаем сразу
		wg.Add(1)
		go func(out chan interface{}) {
			defer wg.Done()
			wg2 := &sync.WaitGroup{}
			// и внутри job распараллеливаем подсчет crc32
			// записываем в массив, каждая ячейка для одной горутины
			results := [6]string{}
			for i := 0; i < 6; i++ {
				wg2.Add(1)
				go func(i int, data string) {
					defer wg2.Done()
					results[i] = DataSignerCrc32(strconv.Itoa(i) + data)
				}(i, data)
			}
			wg2.Wait()
			out <- strings.Join(results[:], "")
		}(out)
	}
	wg.Wait()
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
		// по горутине на каждую job
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
