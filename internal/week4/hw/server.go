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
func (u ByName) Less(i, j int) bool { return u[i].Name > u[j].Name }
func (u ByName) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// параметры из url query string
	queryParams := r.URL.Query()
	query := queryParams.Get("query")
	orderField := queryParams.Get("order_field")
	limit, err := strconv.Atoi(queryParams.Get("limit"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "unable to get limit from request: "+err.Error())
		return
	}
	// в параметрах передается limit + 1, для того чтобы определять следующую страницу
	realLimit := limit - 1

	// валидация orderField
	orderField, err = OrderFieldValidate(orderField)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
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
		// ограничиваем записи по значению realLimit + 1
		// это нужно для логики клиента, где выбираются записи limit + 1
		if realLimit > 0 && realLimit < len(result) {
			break
		}

		// поиск подстроки query в name или about
		whereSearch := row.FirstName + " " + row.LastName + " " + row.About
		if query != "" && !strings.Contains(whereSearch, query) {
			continue
		}
		// если query пустой - возвращаем все результаты
		result = append(result, *RowToUser(&row))
	}

	// сортировка
	if orderBy == OrderByAsc {
		sort.Sort(ByName(result))
	}
	if orderBy == OrderByDesc {
		sort.Sort(sort.Reverse(ByName(result)))
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
