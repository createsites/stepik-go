package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
		// это ошибка 400, а не 500
		{
			Request: SearchRequest{
				OrderField: "Undefined",
			},
			Error: errors.New("OrderFeld Undefined invalid"),
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
		// offset
		{
			Request: SearchRequest{
				OrderField: "Name",
				OrderBy:    OrderByAsc,
				Limit:      1,
				Offset:     2,
			},
			Response: SearchResponse{
				Users: []User{
					{Id: 19},
				},
				NextPage: true,
			},
			Error: nil,
		},
		// error: limit must be > 0
		{
			Request: SearchRequest{
				Limit: -1,
			},
			Error: errors.New("limit must be > 0"),
		},
		// if Limit > 25 then Limit = 25
		{
			Request: SearchRequest{
				Limit: 30,
			},
			Response: SearchResponse{
				// Id == -1 это значит подходит любой юзер
				// если нужно проверить только кол-во, а порядок Id не важен
				Users: []User{
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
					{Id: -1},
				},
				NextPage: true,
			},
			Error: nil,
		},
		// error: Offset < 0
		{
			Request: SearchRequest{
				Offset: -1,
			},
			Error: errors.New("offset must be > 0"),
		},
		// len(data) == req.Limit
		// здесь с учетом фильтра выдается 2 юзера
		// кол-во элементов должно совпадать с лимитом для этого кейса
		{
			Request: SearchRequest{
				Query: "Culpa",
				Limit: 1,
			},
			Response: SearchResponse{
				Users: []User{
					{Id: 24},
				},
				NextPage: true,
			},
		},
		// invalid order by
		{
			Request: SearchRequest{
				OrderBy: -100,
			},
			Error: errors.New("SearchServer fatal error"),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{}
	client.URL = ts.URL
	client.AccessToken = "secret"

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

		// // проверка http статуса
		// if item.HttpStatus != 0 {
		// 	if item.HttpStatus == http.StatusBadRequest {
		// 		if result.
		// 	}
		// 	t.Errorf("[case %d] amount of the users: expected %d, got %d", caseNum, len(item.Response.Users), len(result.Users))
		// }

		if item.Error == nil {
			if testing.Verbose() {
				for _, v := range result.Users {
					fmt.Printf("User Id %d, Name %s\n", v.Id, v.Name)
				}
			}
			// проверка кол-ва элементов
			if len(result.Users) != len(item.Response.Users) {
				t.Errorf("[case %d] amount of the users: expected %d, got %d", caseNum, len(item.Response.Users), len(result.Users))
			}
			// проверка совпадения записей
			for i, user := range result.Users {
				if item.Response.Users[i].Id < 0 {
					continue
				}
				if user.Id != 0 && user.Id != item.Response.Users[i].Id {
					t.Errorf("[case %d] mismatched user ids: expected %d, got %d", caseNum, item.Response.Users[i].Id, user.Id)
				}
			}

		} else {
			// проверка ожидаемых ошибок
			if item.Error.Error() != err.Error() {
				t.Errorf("[case %d] unexpected error: expected '%s', got '%s'", caseNum, item.Error.Error(), err.Error())
			}
		}

		// if !result.NextPage {
		// 	t.Errorf("[%d] expected next page, got false", caseNum)
		// }
		// if !reflect.DeepEqual(item.Result, result) {
		// 	t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		// }
	}
	// проверка авторизации
	client.AccessToken = "wrong"
	_, err := client.FindUsers(cases[0].Request)
	if err == nil {
		t.Errorf("[case unauthorized] expected an unauthorized error, got nil")
		if err.Error() != "Bad AccessToken" {
			t.Errorf("[case unauthorized] expected 'Bad AccessToken error', got '%s'", err.Error())
		}
	}

	// проверка ответа bad request, если не передать токен
	client.AccessToken = ""
	_, err = client.FindUsers(cases[0].Request)
	if err == nil {
		t.Errorf("[case bad request] expected 'token should be passed' error, got nil")
	} else if err.Error() != "unknown bad request error: token should be passed" {
		t.Errorf("[case bad request] expected 'unknown bad request error', got '%s'", err.Error())
	}

	// не удается распарсить json корректного ответа
	// т.к. ожидается корректный, а приходит некорректный ответ
	tsBadJson := httptest.NewServer(http.HandlerFunc(InvalidJsonServer))
	defer tsBadJson.Close()
	client.URL = tsBadJson.URL
	_, err = client.FindUsers(SearchRequest{})
	if err == nil {
		t.Errorf("[case invalid json] expected 'cant unpack result json' error, got nil")
	}
	if !strings.Contains(err.Error(), "cant unpack result json") {
		t.Errorf("[case invalid json] expected 'cant unpack error json' error, got '%s'", err.Error())
	}

	// неправильный формат ошибки при http 400
	tsInvalidError := httptest.NewServer(http.HandlerFunc(InvalidErrorFormatServer))
	defer tsInvalidError.Close()
	client.URL = tsInvalidError.URL
	_, err = client.FindUsers(SearchRequest{})
	if err == nil {
		t.Errorf("[case invalid json] expected 'cant unpack error json' error, got nil")
	}
	if !strings.Contains(err.Error(), "cant unpack error json") {
		t.Errorf("[case invalid json] expected 'cant unpack error json' error, got '%s'", err.Error())
	}
}
