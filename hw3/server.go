package main

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type XMLUser struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// Проверка валидности токена (на случай прямых тестов сервера)
	if r.Header.Get("AccessToken") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	params := r.URL.Query()
	limitStr := params.Get("limit")
	offsetStr := params.Get("offset")
	query := strings.ToLower(params.Get("query"))
	orderField := params.Get("order_field")
	orderByStr := params.Get("order_by")

	// Вспомогательная функция для отправки JSON-ошибок формата {"Error": "..."}
	sendJSONError := func(w http.ResponseWriter, msg string, status int) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"Error": msg})
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		sendJSONError(w, "Invalid limit", http.StatusBadRequest)
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		sendJSONError(w, "Invalid offset", http.StatusBadRequest)
		return
	}
	orderBy, err := strconv.Atoi(orderByStr)
	if err != nil || (orderBy != -1 && orderBy != 0 && orderBy != 1) {
		sendJSONError(w, "Invalid order_by", http.StatusBadRequest)
		return
	}
	if orderField == "" {
		orderField = "Name"
	}

	if orderField != "Id" && orderField != "Age" && orderField != "Name" {
		sendJSONError(w, "OrderField invalid", http.StatusBadRequest)
		return
	}
	file, err := os.Open("dataset.xml")
	if err != nil {
		http.Error(w, "Cannot open dataset", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var dataset struct {
		Rows []XMLUser `xml:"row"`
	}
	if err := xml.NewDecoder(file).Decode(&dataset); err != nil {
		http.Error(w, "Cannot decode dataset", http.StatusInternalServerError)
		return
	}
	var users []User
	for _, xmlUser := range dataset.Rows {
		user := User{
			ID:     xmlUser.ID,
			Name:   xmlUser.FirstName + " " + xmlUser.LastName,
			Age:    xmlUser.Age,
			About:  xmlUser.About,
			Gender: xmlUser.Gender,
		}
		if query == "" || strings.Contains(strings.ToLower(user.Name), query) ||
			strings.Contains(strings.ToLower(user.About), query) {
			users = append(users, user)
		}
	}
	if orderBy != 0 {
		sort.SliceStable(users, func(i, j int) bool {
			var less bool

			switch orderField {
			case "Id":
				less = users[i].ID < users[j].ID
			case "Age":
				less = users[i].Age < users[j].Age
			case "Name":
				nameI := users[i].Name
				nameJ := users[j].Name
				less = nameI < nameJ
			}

			if orderBy == -1 {
				return !less && !equalUsers(users[i], users[j], orderField)
			}
			return less
		})
	}

	if offset >= len(users) {
		users = []User{}
	} else {
		users = users[offset:]
	}

	if limit < len(users) {
		users = users[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Cannot encode response", http.StatusInternalServerError)
		return
	}
}

func equalUsers(a, b User, field string) bool {
	switch field {
	case "Id":
		return a.ID == b.ID
	case "Age":
		return a.Age == b.Age
	case "Name":
		return a.Name == b.Name
	}
	return false
}
