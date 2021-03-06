package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	client = &http.Client{Timeout: time.Second}
)

// User struct describe user
type User struct {
	ID     int
	Name   string
	Age    int
	About  string
	Gender string
}

// SearchResponse contains info about response
type SearchResponse struct {
	Users    []User
	NextPage bool
}

// SearchErrorResponse contains info about response error
type SearchErrorResponse struct {
	Error string
}

// SearchRequest describes request fields
type SearchRequest struct {
	Limit      int
	Offset     int    // Можно учесть после сортировки
	Query      string // подстрока в 1 из полей		Параметр ищет по полям `Name` и `About`
	OrderField string // работает по полям `ID`, `Age`, `Name`, если пустой - то возвращаем по `Name`, если что-то другое - SearchServer ругается ошибкой. `Name` - это first_name + last_name из xml.
	OrderBy    int    // -1 по убыванию, 0 как встретилось, 1 по возрастанию
}

// SearchClient describes struct for search from client
type SearchClient struct {
	// токен, по которому происходит авторизация на внешней системе, уходит туда через хедер
	AccessToken string
	// урл внешней системы, куда идти
	URL string
}

// FindUsers отправляет запрос во внешнюю систему, которая непосредственно ищет пользователей
func (srv *SearchClient) FindUsers(req SearchRequest) (*SearchResponse, error) {

	searcherParams := url.Values{}

	if req.Limit < 0 {
		return nil, fmt.Errorf("limit must be > 0")
	}
	if req.Limit > 25 {
		req.Limit = 25
	}
	if req.Offset < 0 {
		return nil, fmt.Errorf("offset must be > 0")
	}

	// нужно для получения следующей записи, на основе которой мы скажем - можно показать переключатель следующей страницы или нет
	req.Limit++

	searcherParams.Add("limit", strconv.Itoa(req.Limit))
	searcherParams.Add("offset", strconv.Itoa(req.Offset))
	searcherParams.Add("query", req.Query)
	searcherParams.Add("order_field", req.OrderField)
	searcherParams.Add("order_by", strconv.Itoa(req.OrderBy))

	searcherReq, _ := http.NewRequest("GET", srv.URL+"?"+searcherParams.Encode(), nil)
	searcherReq.Header.Add("AccessToken", srv.AccessToken)

	resp, err := client.Do(searcherReq)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil, fmt.Errorf("timeout for %s", searcherParams.Encode())
		}
		return nil, fmt.Errorf("unknown error %s", err)
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			fmt.Println("cannot close response body", err.Error())
			return
		}
	}()
	body, _ := ioutil.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("bad AccessToken") // Wrong AccessToken
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("searchServer fatal error")
	case http.StatusBadRequest:
		errResp := SearchErrorResponse{}
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return nil, fmt.Errorf("cant unpack error json: %s", err)
		}
		if errResp.Error == "ErrorBadOrderField" {
			return nil, fmt.Errorf("OrderFeld %s invalid", req.OrderField)
		}
		return nil, fmt.Errorf("unknown bad request error: %s", errResp.Error)
	}

	var data []User
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("cant unpack result json: %s", err)
	}

	result := SearchResponse{}
	if len(data) == req.Limit {
		result.NextPage = true
		result.Users = data[0 : len(data)-1]
	} else {
		result.Users = data[0:]
	}

	return &result, err
}
