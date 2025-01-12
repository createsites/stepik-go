
# Результаты оптимизации

1. Заменил все regexp аналогичными функциями из пакета strings.
2. Удалил строку foundUsers, вместо нее пишу напрямую в out.
3. Заменил чтение файла с данными в память на сканирование построчно.

### Memory до оптимизации

[png профайла памяти](profile_mem.png)


### CPU

Было `pprof > list SlowSearch`

- 700ms     62:			if ok, err := regexp.MatchString("Android", browser); ok && err == nil
- 530ms     84:			if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
  
Стало `pprof > list FastSearch`

- 60ms     92:			if ok := re.MatchString(browser); ok {
- 20ms    115:			if ok := re.MatchString(browser); ok {

### Memory

Было `pprof > list SlowSearch`

- 585.35MB     62:			if ok, err := regexp.MatchString("Android", browser); ok && err == nil
- 369.21MB     84:			if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {
- 11.53MB    106:		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)

Стало `pprof > list FastSearch`

- .     89:			if strings.Contains(browser, "Android") {
- .    111:			if strings.Contains(browser, "MSIE") {
- 513kB    133:		fmt.Fprintf(out, "[%d] %s <%s>\n", i, user["name"], email)

## Итого

```
pkg: hw3
cpu: AMD Ryzen 7 4800H with Radeon Graphics         
BenchmarkSlow-4   	      33	  35870122 ns/op	20180990 B/op	  182831 allocs/op
BenchmarkFast-4   	      90	  11397899 ns/op	 2718268 B/op	   47247 allocs/op
PASS
ok  	hw3	2.363s
```