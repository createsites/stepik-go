package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Data struct {
	Rows []Row `xml:"row"`
}

type Row struct {
	Id        int `xml:"id"`
	Age       int `xml:"age"`
	Gender    string `xml:"gender"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	About     string `xml:"about"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// параметры из url query string
	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")

	// валидация orderField
	orderField, err := OrderFieldValidate(orderField)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	// чтение xml
	file, err := os.Open("dataset.xml")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "unable to open data file: " + err.Error())
		return
	}
	defer file.Close()

	dataRaw, err := io.ReadAll(file)
	if err != nil {
		io.WriteString(w, "unable to read data file: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := new(Data)
	err = xml.Unmarshal(dataRaw, data)
	if err != nil {
		io.WriteString(w, "unable to decode xml data: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if query == "" {
		// todo в объектах нет Name
		fmt.Fprintf(w, "%#v", data.Rows)
		return
	}

	// поиск подстроки query в name или about
	result := make([]User, 0, len(data.Rows))
	for _, row := range data.Rows {
		name := row.FirstName + " " + row.LastName
		if strings.Contains(name+" "+row.About, query) {
			result = append(result, User{
				Id:     row.Id,
				Name:   name,
				Age:    row.Age,
				About:  row.About,
				Gender: row.Gender,
			})
		}
	}
	jsonData, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
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

func main() {
	// result, err := SearchServer("Culpa", "")
	// if err != nil {
	// 	panic(err)
	// }
	// for _, v := range result {
	// 	fmt.Printf("%#v\n\n", v)
	// }

}
