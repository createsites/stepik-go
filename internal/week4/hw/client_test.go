package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func CheckoutDummy(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `[{"Id": 1, "Name": "Tester", "Age": 21, "About": "Something", "Gender": "male"}]`)
}

type TestCase struct {
	Request  SearchRequest
	Response SearchResponse
	Error    error
}

func TestFindUsers(t *testing.T) {
	cases := []TestCase{
		// находит 2 юзера по подстроке Culpa
		{
			Request: SearchRequest{
				Query: "Culpa", // case sensitive
			},
			Response: SearchResponse{
				Users: []User{
					{Id: 24},
					{Id: 32},
				},
				NextPage: false,
			},
			Error: nil,
		},
		// limit 1, должен выдать постраничную навигацию
		{
			Request: SearchRequest{
				Query: "",
				Limit: 1,
			},
			Response: SearchResponse{
				Users: []User{
					{Id: 24},
				},
				NextPage: true,
			},
			Error: nil,
		},
		// не корректный order_field
		{
			Request: SearchRequest{
				OrderField: "Undefined",
			},
			Error: errors.New("SearchServer fatal error"),
		},
		// сортировка по id по убыванию
		{
			Request: SearchRequest{
				Query:      "Culpa",
				OrderField: "Id",
				OrderBy:    OrderByDesc,
			},
			Response: SearchResponse{
				Users: []User{
					{Id: 32},
					{Id: 24},
				},
				NextPage: false,
			},
			Error: nil,
		},
		// сортировка по name по возрастанию
		{
			Request: SearchRequest{
				OrderField: "Name",
				OrderBy:    OrderByAsc,
				Limit:      4,
			},
			Response: SearchResponse{
				Users: []User{
					{Id: 15},
					{Id: 16},
					{Id: 19},
					{Id: 22},
				},
				NextPage: true,
			},
			Error: nil,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := &SearchClient{}
	client.URL = ts.URL

	for caseNum, item := range cases {
		if testing.Verbose() {
			fmt.Printf("--------------\nTest case %d\n--------------\n", caseNum)
		}

		result, err := client.FindUsers(item.Request)
		// неожиданная ошибка
		if err != nil && item.Error == nil {
			t.Errorf("[case %d] unexpected FindUsers error: %s", caseNum, err.Error())
			return
		}

		if item.Error == nil {

			for _, v := range result.Users {
				fmt.Printf("User Id %d, Name %s\n", v.Id, v.Name)
			}

			// проверка кол-ва элементов
			if len(result.Users) != len(item.Response.Users) {

				t.Errorf("[case %d] amount of the users: expected %d, got %d", caseNum, len(item.Response.Users), len(result.Users))
			}
			// проверка совпадения записей
			for i, user := range result.Users {
				if user.Id != 0 && user.Id != item.Response.Users[i].Id {
					t.Errorf("[case %d] mismatched user ids: expected %d, got %d", caseNum, item.Response.Users[i].Id, user.Id)
				}
			}

		} else {
			// проверка ожидаемых ошибок
			if item.Error.Error() != err.Error() {
				t.Errorf("[case %d] unexpected error: expected %s, got %s", caseNum, item.Error.Error(), err.Error())
			}
		}

		// if !result.NextPage {
		// 	t.Errorf("[%d] expected next page, got false", caseNum)
		// }
		// if !reflect.DeepEqual(item.Result, result) {
		// 	t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		// }
	}
	ts.Close()
}
