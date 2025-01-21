package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Data struct {
	Rows []Row `xml:"row"`
}

type Row struct {
	Id        int    `xml:"id"`
	Age       int    `xml:"age"`
	Gender    string `xml:"gender"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	About     string `xml:"about"`
}

// для сортировки по имени
type ByName []User

func (u ByName) Len() int           { return len(u) }
func (u ByName) Less(i, j int) bool { return u[i].Name < u[j].Name }
func (u ByName) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }

// для сортировки по Id
type ById []User

func (u ById) Len() int           { return len(u) }
func (u ById) Less(i, j int) bool { return u[i].Id < u[j].Id }
func (u ById) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }

// для сортировки по возрасту
type ByAge []User

func (u ByAge) Len() int           { return len(u) }
func (u ByAge) Less(i, j int) bool { return u[i].Age < u[j].Age }
func (u ByAge) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// параметры из url query string
	queryParams := r.URL.Query()
	query := queryParams.Get("query")
	orderField := queryParams.Get("order_field")
	offset, err := strconv.Atoi(queryParams.Get("offset"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "unable to convert offset from string to int: "+err.Error())
		return
	}
	limit, err := strconv.Atoi(queryParams.Get("limit"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "unable to get limit from request: "+err.Error())
		return
	}
	// в параметрах передается limit + 1, для того чтобы определять следующую страницу
	realLimit := limit - 1

	// для проверки http 400
	if r.Header.Get("AccessToken") == "" {
		w.WriteHeader(http.StatusBadRequest)
		errStr := SearchErrorResponse{"token should be passed"}
		jsonErr, _ := json.Marshal(errStr)
		w.Write(jsonErr)
		return
	}

	// проверка токена авторизации
	if r.Header.Get("AccessToken") != "secret" {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "client is not authorized")
		return
	}

	// валидация orderField
	orderField, err = OrderFieldValidate(orderField)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errStr := SearchErrorResponse{ErrorBadOrderField}
		jsonErr, _ := json.Marshal(errStr)
		w.Write(jsonErr)
		return
	}

	// валидация orderBy
	orderBy, err := OrderByValidate(queryParams.Get("order_by"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	// чтение xml
	file, err := os.Open("dataset.xml")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "unable to open data file: "+err.Error())
		return
	}
	defer file.Close()

	dataRaw, err := io.ReadAll(file)
	if err != nil {
		io.WriteString(w, "unable to read data file: "+err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := new(Data)
	err = xml.Unmarshal(dataRaw, data)
	if err != nil {
		io.WriteString(w, "unable to decode xml data: "+err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := make([]User, 0, len(data.Rows))
	for _, row := range data.Rows {
		result = append(result, *RowToUser(&row))
	}

	// сортировка
	if orderBy == OrderByAsc {
		if orderField == "Id" {
			sort.Sort(ById(result))
		} else if orderField == "Age" {
			sort.Sort(ByAge(result))
		} else {
			sort.Sort(ByName(result))
		}
	}
	if orderBy == OrderByDesc {
		if orderField == "Id" {
			sort.Sort(sort.Reverse(ById(result)))
		} else if orderField == "Age" {
			sort.Sort(sort.Reverse(ByAge(result)))
		} else {
			sort.Sort(sort.Reverse(ByName(result)))
		}
	}

	// фильтрация
	if query != "" {
		// убираем из слайса элементы, сдвигая справа на место не нужных
		// счетчик удаленных эл-ов
		var deleted int
		// указатель на текущий элемент после фильтрации
		// с каждым отфильтрованным элементом он будет отставать на 1 от i
		var posAfterFilter int
		for i := 0; i < len(result); i++ {
			// поиск подстроки query в name или about
			whereSearch := result[i].Name + " " + result[i].About
			// не нашлось
			if query != "" && !strings.Contains(whereSearch, query) {
				deleted++
				continue
			}
			result[posAfterFilter] = result[i]
			posAfterFilter++
		}
		// и потом урезаем слайс справа на кол-во удаленных эл-в
		result = result[:(len(result) - deleted)]
	}

	// лимит и оффсет
	// ограничиваем записи по значению realLimit + 1
	// это нужно для логики клиента, где выбираются записи limit + 1
	if realLimit > 0 && realLimit < len(result) {
		result = result[offset : realLimit+offset]
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

func RowToUser(row *Row) *User {
	return &User{
		Id:     row.Id,
		Name:   row.FirstName + " " + row.LastName,
		Age:    row.Age,
		About:  row.About,
		Gender: row.Gender,
	}
}

func OrderFieldValidate(order string) (string, error) {
	if order == "" {
		return "Name", nil
	}
	if order == "Name" || order == "Age" || order == "Id" {
		return order, nil
	}
	return "", errors.New(ErrorBadOrderField)
}

func OrderByValidate(orderByRaw string) (int, error) {
	orderBy, err := strconv.Atoi(orderByRaw)
	if err != nil {
		return 0, errors.New("can not convert the order by to int")
	}
	if orderBy != OrderByAsIs && orderBy != OrderByAsc && orderBy != OrderByDesc {
		return 0, errors.New("bad order by")
	}
	return orderBy, nil
}
