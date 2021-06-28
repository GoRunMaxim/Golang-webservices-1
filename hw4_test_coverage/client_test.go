package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

const filePath = "dataset.xml"

type Users struct {
	XMLName xml.Name  `xml:"root"`
	Users   []XMLUser `xml:"row"`
}

// XMLUser describes field that we get from xml file
type XMLUser struct {
	ID        int    `xml:"id"`
	GUID      string `xml:"-"`
	IsActive  bool   `xml:"-"`
	Balance   string `xml:"-"`
	Picture   string `xml:"-"`
	Age       int    `xml:"age"`
	EyeColor  string `xml:"-"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Gender    string `xml:"gender"`
	Company   string `xml:"-"`
	Email     string `xml:"-"`
	Phone     string `xml:"-"`
	Address   string `xml:"-"`
	About     string `xml:"about"`
	RegTime   string `xml:"-"`
	FavFruit  string `xml:"-"`
	Name      string
}

type TestCase struct {
	Request     SearchRequest
	Response    *SearchResponse
	Error       string
	IsError     bool
	AccessToken string
}

type By func(u1, u2 *User) bool
type userSorter struct {
	users []User
	by    func(u1, u2 *User) bool // Closure used in the Less method.
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	var js []byte
	xmlFile, err := os.Open(filePath)

	if err != nil {
		fmt.Println("unable to open file, ", err)
	}

	defer func() {
		if err = xmlFile.Close(); err != nil {
			panic(err)
		}
	}()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	users := Users{}

	err = xml.Unmarshal(byteValue, &users)
	if err != nil {
		logrus.Warn("cannot unmarshal xml file: ", err.Error())
		return
	}

	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		err = fmt.Errorf("format error: cannot convert LIMIT field %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		err = fmt.Errorf("format error: cannot convert OFFSET field %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orderBy, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		err = fmt.Errorf("format error: cannot convert ORDER_BY field %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if (orderBy < -1) || (orderBy > 1) {
		w.WriteHeader(http.StatusBadRequest) // вернуть ошибку в json
		errResp := SearchErrorResponse{Error: "OrderByError"}
		js, err = json.Marshal(errResp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(js)
		if err != nil {
			logrus.Warn("cannot write to writer: ", err.Error())
			return
		}
		return
	}

	req := SearchRequest{
		Limit:      limit,
		Offset:     offset,
		Query:      r.FormValue("query"),
		OrderField: r.FormValue("order_field"),
		OrderBy:    orderBy,
	}

	var finalUsers []User
	for _, value := range users.Users {
		value.Name = value.FirstName + " " + value.LastName
		if strings.Contains(value.Name, req.Query) || strings.Contains(value.About, req.Query) {
			user := User{
				ID:     value.ID,
				Name:   value.Name,
				Age:    value.Age,
				About:  value.About,
				Gender: value.Gender,
			}
			finalUsers = append(finalUsers, user)
		}
	}

	// Sort after founding
	switch req.OrderField {
	case "Id":
		{
			if req.OrderBy == 1 {
				id := func(u1, u2 *User) bool {
					return u1.ID < u2.ID
				}
				By(id).Sort(finalUsers)
			} else if req.OrderBy == -1 {
				id := func(u1, u2 *User) bool {
					return u1.ID > u2.ID
				}
				By(id).Sort(finalUsers)
			}
		}
	case "Age":
		{
			if req.OrderBy == 1 {
				age := func(u1, u2 *User) bool {
					return u1.Age < u2.Age
				}
				By(age).Sort(finalUsers)
			} else if req.OrderBy == -1 {
				age := func(u1, u2 *User) bool {
					return u1.Age > u2.Age
				}
				By(age).Sort(finalUsers)
			}
		}
	case "Name":
		{
			if req.OrderBy == 1 {
				name := func(u1, u2 *User) bool {
					return u1.Name < u2.Name
				}
				By(name).Sort(finalUsers)
			} else if req.OrderBy == -1 {
				name := func(u1, u2 *User) bool {
					return u1.Name > u2.Name
				}
				By(name).Sort(finalUsers)
			}
		}
	case "":
		{
			if req.OrderBy == 1 {
				name := func(u1, u2 *User) bool {
					return u1.Name < u2.Name
				}
				By(name).Sort(finalUsers)
			} else if req.OrderBy == -1 {
				name := func(u1, u2 *User) bool {
					return u1.Name > u2.Name
				}
				By(name).Sort(finalUsers)
			}
		}
	default:
		{
			w.WriteHeader(http.StatusBadRequest)
			errResp := SearchErrorResponse{Error: "ErrorBadOrderField"}
			js, err = json.Marshal(errResp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = w.Write(js)
			if err != nil {
				logrus.Warn("cannot write to writer: ", err.Error())
				return
			}
			return
		}
	}

	accessToken := r.Header.Get("AccessToken")

	switch accessToken {
	case "https://example.auth0.com/":
		w.WriteHeader(http.StatusUnauthorized)
	case "error404":
		w.WriteHeader(http.StatusInternalServerError)
	case "https://example.com/unknownRequest":
		{
			w.WriteHeader(http.StatusBadRequest) // вернуть ошибку в json
			errResp := SearchErrorResponse{Error: "UnknownError"}
			js, err = json.Marshal(errResp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = w.Write(js)
			if err != nil {
				logrus.Warn("cannot write to writer: ", err.Error())
				return
			}
			return
		}
	default:
		w.WriteHeader(http.StatusOK)
	}

	js, err = json.Marshal(finalUsers)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(js)
	if err != nil {
		logrus.Warn("cannot write to writer: ", err.Error())
		return
	}
}

func (by By) Sort(users []User) {
	us := &userSorter{
		users: users,
		by:    by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(us)
}
func (u *userSorter) Len() int {
	return len(u.users)
}
func (u *userSorter) Swap(i, j int) {
	u.users[i], u.users[j] = u.users[j], u.users[i]
}
func (u *userSorter) Less(i, j int) bool {
	return u.by(&u.users[i], &u.users[j])
}

func BadSearch(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	js, err := json.Marshal("{ \"key\": \"<div class=\"coolCSS\">some text</div>\" }")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(js)
	if err != nil {
		logrus.Warn("cannot write to writer: ", err.Error())
		return
	}
}

func BadSearchJSON(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	js, err := json.Marshal("{ \"key\": \"<div class=\"coolCSS\">some text</div>\" }")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(js)
	if err != nil {
		logrus.Warn("cannot write to writer: ", err.Error())
		return
	}
}

func TestSearchClient_FindUsers(t *testing.T) {
	var cases = []TestCase{{
		Request: SearchRequest{
			Limit:      40,
			Offset:     2,
			Query:      "Boyd Wolf",
			OrderField: "",
			OrderBy:    0,
		},
		Response: &SearchResponse{
			Users: []User{{
				ID:     0,
				Name:   "Boyd Wolf",
				Age:    22,
				About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
				Gender: "male",
			},
			},
			NextPage: false,
		},
		Error:   "",
		IsError: false,
	}, {Request: SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "Boyd Wolf",
		OrderField: "",
		OrderBy:    0,
	},
		Response: &SearchResponse{
			Users:    []User{},
			NextPage: true,
		},
		Error:   "",
		IsError: false,
	},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	defer ts.Close()

	for caseNum, item := range cases {
		s := &SearchClient{
			URL: ts.URL,
		}
		result, err := s.FindUsers(item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.Response, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Response, result)
		}
	}
}

func TestErrors_FindUsers(t *testing.T) {
	var cases = []TestCase{{
		Request: SearchRequest{
			Limit:      0,
			Offset:     -1,
			Query:      "",
			OrderField: "",
			OrderBy:    0,
		},
		Response: nil,
		IsError:  true,
	}, // offset error <0
		{Request: SearchRequest{
			Limit:      -1,
			Offset:     0,
			Query:      "",
			OrderField: "",
			OrderBy:    0,
		},
			Response: nil,
			Error:    "",
			IsError:  true,
		}, // Limit error (<0)
		{Request: SearchRequest{
			Limit:      30,
			Offset:     0,
			Query:      "",
			OrderField: "kek",
			OrderBy:    0,
		},
			Response: nil,
			Error:    "",
			IsError:  true,
		}, // Order field error
		{Request: SearchRequest{
			Limit:      10,
			Offset:     2,
			Query:      "Boyd Wolf",
			OrderField: "",
			OrderBy:    10,
		},
			Response: nil,
			Error:    "",
			IsError:  true,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	defer ts.Close()

	for caseNum, item := range cases {
		s := &SearchClient{
			URL: ts.URL,
		}
		result, err := s.FindUsers(item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.Response, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Response, result)
		}
	}
}

func TestStatusCode(t *testing.T) {
	var cases = []TestCase{{
		Request: SearchRequest{
			Limit:      40,
			Offset:     2,
			Query:      "Boyd Wolf",
			OrderField: "",
			OrderBy:    0,
		},
		Response:    nil,
		Error:       "",
		IsError:     true,
		AccessToken: "https://example.auth0.com/",
	}, {Request: SearchRequest{
		Limit:      40,
		Offset:     2,
		Query:      "Boyd Wolf",
		OrderField: "",
		OrderBy:    0,
	},
		Response:    nil,
		Error:       "",
		IsError:     true,
		AccessToken: "error404",
	}, {Request: SearchRequest{
		Limit:      40,
		Offset:     2,
		Query:      "Boyd Wolf",
		OrderField: "",
		OrderBy:    0,
	},
		Response:    nil,
		Error:       "",
		IsError:     true,
		AccessToken: "https://example.com/unknownRequest",
	},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	defer ts.Close()

	for caseNum, item := range cases {
		s := &SearchClient{
			URL:         ts.URL,
			AccessToken: item.AccessToken,
		}
		result, err := s.FindUsers(item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.Response, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Response, result)
		}
	}
}

func TestErrorJson(t *testing.T) {
	var cases = []TestCase{{
		Request: SearchRequest{
			Limit:      10,
			Offset:     2,
			Query:      "Boyd Wolf",
			OrderField: "",
			OrderBy:    10,
		},
		Response: nil,
		Error:    "",
		IsError:  true,
	},
	}
	ts := httptest.NewServer(http.HandlerFunc(BadSearch))

	defer ts.Close()

	for caseNum, item := range cases {
		s := &SearchClient{
			URL:         ts.URL,
			AccessToken: item.AccessToken,
		}
		result, err := s.FindUsers(item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.Response, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Response, result)
		}
	}
}

func TestUnpackJson(t *testing.T) {
	var cases = []TestCase{{
		Request: SearchRequest{
			Limit:      10,
			Offset:     2,
			Query:      "Boyd Wolf",
			OrderField: "",
			OrderBy:    10,
		},
		Response: nil,
		Error:    "",
		IsError:  true,
	}, {Request: SearchRequest{
		Limit:      10,
		Offset:     2,
		Query:      "Boyd Wolf",
		OrderField: "",
		OrderBy:    10,
	},
		Response: nil,
		Error:    "",
		IsError:  true,
	},
	}
	ts := httptest.NewServer(http.HandlerFunc(BadSearchJSON))

	defer ts.Close()

	for caseNum, item := range cases {
		s := &SearchClient{
			URL:         ts.URL,
			AccessToken: item.AccessToken,
		}
		result, err := s.FindUsers(item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.Response, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Response, result)
		}
	}
}

func TestTimeoutError(t *testing.T) {
	var cases = []TestCase{{
		Request: SearchRequest{
			Limit:      10,
			Offset:     2,
			Query:      "Boyd Wolf",
			OrderField: "",
			OrderBy:    10,
		},
		Response: nil,
		Error:    "",
		IsError:  true,
	}, {
		Request: SearchRequest{
			Limit:      10,
			Offset:     2,
			Query:      "Boyd Wolf",
			OrderField: "",
			OrderBy:    10,
		},
		Response: nil,
		Error:    "",
		IsError:  true,
	},
	}
	ts := httptest.NewServer(http.HandlerFunc(SlowSearch))

	defer ts.Close()

	for caseNum, item := range cases {
		s := &SearchClient{
			URL:         ts.URL,
			AccessToken: item.AccessToken,
		}
		result, err := s.FindUsers(item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.Response, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Response, result)
		}
	}
}

func TestUnableServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(UnableServer))
	searchClient := &SearchClient{
		URL: "some bad link",
	}

	_, err := searchClient.FindUsers(SearchRequest{})

	if err == nil {
		t.Error("Unknown error")
	}

	defer ts.Close()
}

func SlowSearch(w http.ResponseWriter, r *http.Request) {
	time.Sleep(3 * time.Second)
	w.WriteHeader(http.StatusOK)
	js, err := json.Marshal("{ \"key\": \"<div class=\"coolCSS\">some text</div>\" }")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(js)
	if err != nil {
		logrus.Warn("cannot write to writer: ", err.Error())
		return
	}
}

func UnableServer(http.ResponseWriter, *http.Request) {}
