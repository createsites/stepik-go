package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

/*
запись бенчмарка в файл
go test -bench . -benchmem -cpuprofile=report/cpu.out -memprofile=report/mem.out

профайлинг записанного бенчмарка
go tool pprof report/hw3.test report/cpu.out

hw3.test - скомпилированный бинарник
cpu.out - файл бенчмарка из предыдущей команды
Если добавить флаг -http=:8083 то выведется сразу граф в браузер

Профайлим так:
go tool pprof hw3.test cpu.out (это по cpu, для профайла по памяти меняем на mem.out)
в консоли откроется интерактивный режим pprof>
в нем команды:
top - выводит топ потребления
list <название функции> - показывает исходный код с отметками на какой строке потребление
web - вывести в браузер
png > <адрес сохраняемого файла> - сохраняет граф картинкой
*/

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	/*
		!!! !!! !!!
		обратите внимание - в задании обязательно нужен отчет
		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
		так же обратите внимание на команду в параметром -http
		перечитайте еще раз задание
		!!! !!! !!!
	*/
	// SlowSearch(out)

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

    scanner := bufio.NewScanner(file)

	seenBrowsers := []string{}
	uniqueBrowsers := 0
	fmt.Fprintf(out, "found users:\n")
	
	users := make([]map[string]interface{}, 0)
	// for _, line := range lines {
	for scanner.Scan() {
		line := scanner.Text()
		
		user := make(map[string]interface{})
		// fmt.Printf("%v %v\n", err, line)
		err := json.Unmarshal([]byte(line), &user)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	for i, user := range users {

		isAndroid := false
		isMSIE := false

		browsers, ok := user["browsers"].([]interface{})
		if !ok {
			// log.Println("cant cast browsers")
			continue
		}

		for _, browserRaw := range browsers {
			browser, ok := browserRaw.(string)
			if !ok {
				// log.Println("cant cast browser to string")
				continue
			}
			if strings.Contains(browser, "Android") {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		for _, browserRaw := range browsers {
			browser, ok := browserRaw.(string)
			if !ok {
				// log.Println("cant cast browser to string")
				continue
			}
			if strings.Contains(browser, "MSIE") {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := strings.ReplaceAll(user["email"].(string), "@", " [at] ")
		fmt.Fprintf(out, "[%d] %s <%s>\n", i, user["name"], email)
	}
	fmt.Fprintf(out, "\n")
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
