package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Data struct {
	Rows []Row `xml:"row"`
}

type Row struct {
	Name     string `xml:"first_name"`
	LastName string `xml:"last_name"`
	About    string `xml:"about"`
}

func SearchServer(query, orderField string /*, orderBy, limit, offset int*/) ([]Row, *SearchErrorResponse) {
	// todo брать параметры из url query string

	errResult := &SearchErrorResponse{}

	// валидация orderField
	orderField, err := OrderFieldValidate(orderField)
	if err != nil {
		errResult.Error = err.Error()
		return nil, errResult
	}

	// чтение xml
	file, err := os.Open("dataset.xml")
	if err != nil {
		errResult.Error = "unable to open data file: " + err.Error()
		return nil, errResult
	}
	defer file.Close()

	dataRaw, err := io.ReadAll(file)
	if err != nil {
		errResult.Error = "unable to read data file: " + err.Error()
		return nil, errResult
	}

	data := new(Data)
	err = xml.Unmarshal(dataRaw, data)
	if err != nil {
		errResult.Error = "unable to decode xml data: " + err.Error()
		return nil, errResult
	}

	if query == "" {
		return data.Rows, nil
	}

	// поиск подстроки query в name или about
	result := make([]Row, 0, len(data.Rows))
	for _, row := range data.Rows {
		name := row.Name + " " + row.LastName
		if strings.Contains(name+" "+row.About, query) {
			result = append(result, row)
		}
	}
	return result, nil
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
	result, err := SearchServer("Culpa", "")
	if err != nil {
		panic(err)
	}
	for _, v := range result {
		fmt.Printf("%#v\n\n", v)
	}

}
