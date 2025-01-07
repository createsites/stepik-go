package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

type Entry interface {
	GetName() string
	IsLast() bool
}

// дирректория
type EntryDir struct {
	name   string
	isLast bool
}

func (e EntryDir) GetName() string {
	return e.name
}

func (e EntryDir) IsLast() bool {
	return e.isLast
}

// файл
type EntryFile struct {
	name   string
	isLast bool
	info   fs.FileInfo
}

func (e EntryFile) GetName() string {
	return e.name
}

func (e EntryFile) GetSize() int64 {
	return e.info.Size()
}

func (e EntryFile) IsLast() bool {
	return e.isLast
}

// глубина
var depth int

// мапа содержит [глубина]дошли ли до последнего файла на этой глубине
var lastEntries map[int]bool = make(map[int]bool)

// выводит в поток содержимое директории
// вызывается рекурсивно чтобы пройти по всем вложенным папкам
func dirTree(out io.Writer, path string, printFiles bool) error {
	depth++
	lastEntries[depth] = false
	// открываем дескриптор
	r, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open path: %s", err.Error())
	}
	// читаем все файлы и папки
	entries, err := r.ReadDir(0)
	if err != nil {
		return fmt.Errorf("failed to read dir: %s", err.Error())
	}
	// сортируем
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	// отсеиваем файлы, если не указан флаг -f
	if !printFiles {
		newEntries := make([]os.DirEntry, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				newEntries = append(newEntries, e)
			}
		}
		entries = newEntries
	}
	// цикл по всем вхождениям в директории
	for i := 0; i < len(entries); i++ {
		e := entries[i]

		isLastEntry := false

		if i == len(entries)-1 {
			isLastEntry = true
			lastEntries[depth] = true
		}
		// определяем вхождение (директория или файл)
		var displayEntry Entry
		if e.IsDir() {
			displayEntry = EntryDir{
				name:   e.Name(),
				isLast: isLastEntry,
			}
		} else {
			info, _ := e.Info()
			displayEntry = EntryFile{
				name:   e.Name(),
				isLast: isLastEntry,
				info:   info,
			}
		}
		// выводим
		printEntry(out, displayEntry, depth)
		// для директории запускаем функцию рекурсивно
		if e.IsDir() {
			relPath := filepath.Join(path, e.Name())
			dirTree(out, relPath, printFiles)
		}
	}

	depth--

	return nil
}

// формирует строку для вывода и отправляет в выходной поток
func printEntry(out io.Writer, e Entry, depth int) {
	// префикс вхождения
	prefix := "├───"
	if e.IsLast() {
		prefix = "└───"
	}
	// вывод вертикальной черты
	for i := 1; i < depth; i++ {
		if !lastEntries[i] {
			fmt.Fprint(out, "│")
		}
		fmt.Fprint(out, "\t")
	}
	// вывод в зависимости от типа вхождения
	switch e.(type) {
	// файл
	case EntryFile:
		var sizeStr string
		if size := e.(EntryFile).GetSize(); size > 0 {
			sizeStr = fmt.Sprintf("(%db)", size)
		} else {
			sizeStr = "(empty)"
		}

		fmt.Fprintf(out, "%v%v %s\n", prefix, e.GetName(), sizeStr)
	// директория
	default:
		fmt.Fprintf(out, "%v%v\n", prefix, e.GetName())
	}
}
