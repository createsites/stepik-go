package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"os"
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
	realLimit := limit-1

	// валидация orderField
	orderField, err = OrderFieldValidate(orderField)
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
