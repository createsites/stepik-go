package main

import (
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
		{
			Request: SearchRequest{
				Query: "Culpa", // case sensitive
			},
			Response: SearchResponse{
				Users:    []User{
					{Id: 24},
					{Id: 32},
				},
				NextPage: false,
			},
			Error: nil,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := &SearchClient{}
	client.URL = ts.URL

	for caseNum, item := range cases {

		result, err := client.FindUsers(item.Request)
		if err != nil {
			t.Errorf("[%d] unexpected FindUsers error: %s", caseNum, err.Error())
			return
		}

		if len(result.Users) != len(item.Response.Users) {
			t.Errorf("[%d] expected equal amount of the users, got %d : %d (result : test case))", caseNum, len(result.Users), len(item.Response.Users))
		}

		// for _, v := range result.Users {
		// 	t.Errorf("%#v\n\n", v)
		// 	return
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
