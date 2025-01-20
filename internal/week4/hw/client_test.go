package main

import (
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
	Error    *SearchErrorResponse
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
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := &SearchClient{}
	client.URL = ts.URL

	for caseNum, item := range cases {

		fmt.Printf("--------------\nTest case %d\n--------------\n", caseNum)

		result, err := client.FindUsers(item.Request)
		if err != nil {
			t.Errorf("[case %d] unexpected FindUsers error: %s", caseNum, err.Error())
			return
		}

		// fmt.Printf("%#v\n", result)

		if len(result.Users) != len(item.Response.Users) {
			t.Errorf("[case %d] amount of the users: expected %d, got %d", caseNum, len(item.Response.Users), len(result.Users))
		}

		// for _, v := range result.Users {
		// 	fmt.Printf("User Id %d, Name %s\n", v.Id, v.Name)
		// }

		// if err != nil {
		// 	t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		// }
		// if !result.NextPage {
		// 	t.Errorf("[%d] expected next page, got false", caseNum)
		// }
		// if !reflect.DeepEqual(item.Result, result) {
		// 	t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		// }
	}
	ts.Close()
}
