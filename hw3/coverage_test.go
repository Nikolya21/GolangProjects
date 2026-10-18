package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// Тестовый XML из 5 записей, чтобы проверить пагинацию, equalUsers и все сортировки
const testXMLData = `<?xml version="1.0" encoding="utf-8"?>
<root>
    <row>
        <id>1</id>
        <first_name>Boyd</first_name>
        <last_name>Wolf</last_name>
        <age>22</age>
        <about>Nulla cillum...</about>
        <gender>male</gender>
    </row>
    <row>
        <id>2</id>
        <first_name>Asher</first_name>
        <last_name>Raymond</last_name>
        <age>35</age>
        <about>Laborum non...</about>
        <gender>male</gender>
    </row>
    <row>
        <id>3</id>
        <first_name>Boyd</first_name>
        <last_name>Wolf</last_name>
        <age>22</age>
        <about>Same name and age for equalUsers check</about>
        <gender>male</gender>
    </row>
    <row>
        <id>4</id>
        <first_name>Zack</first_name>
        <last_name>Fair</last_name>
        <age>24</age>
        <about>Solidity engineer</about>
        <gender>male</gender>
    </row>
</root>`

func TestFindUsers_AllScenarios(t *testing.T) {
	err := os.WriteFile("dataset.xml", []byte(testXMLData), 0644)
	if err != nil {
		t.Fatal("Не удалось создать тестовый dataset.xml:", err)
	}
	defer os.Remove("dataset.xml")

	// Переопределяем глобальный клиент с маленьким таймаутом для теста таймаута
	oldClient := client
	client = &http.Client{Timeout: 50 * time.Millisecond}
	defer func() { client = oldClient }()

	tests := []struct {
		name           string
		accessToken    string
		req            SearchRequest
		corruptXML     bool
		deleteXML      bool
		serverError    bool
		invalidJSON    bool
		badRequestText bool // Флаг для отправки текста вместо JSON при 400 ошибке
		timeoutError   bool
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:        "1. Успешный запрос, NextPage = true (запросили 2, в базе больше)",
			accessToken: "valid_token",
			req: SearchRequest{
				Limit:      2,
				Offset:     0,
				OrderField: "Name",
				OrderBy:    0,
			},
			expectedError: false,
		},
		{
			name:        "2. Успешный запрос, NextPage = false (запросили много, вернулось все)",
			accessToken: "valid_token",
			req: SearchRequest{
				Limit:      25,
				Offset:     0,
				OrderField: "Name",
				OrderBy:    0,
			},
			expectedError: false,
		},
		{
			name:        "3. Фильтрация по Query и автоматический сброс лимита > 25",
			accessToken: "valid_token",
			req: SearchRequest{
				Query: "Asher",
				Limit: 30, // Сбросится до 25 внутри клиента
			},
			expectedError: false,
		},
		{
			name:        "4. Клиентская валидация: Отрицательный Limit",
			accessToken: "valid_token",
			req: SearchRequest{
				Limit: -1,
			},
			expectedError:  true,
			expectedErrMsg: "limit must be > 0",
		},
		{
			name:        "5. Клиентская валидация: Отрицательный Offset",
			accessToken: "valid_token",
			req: SearchRequest{
				Offset: -1,
			},
			expectedError:  true,
			expectedErrMsg: "offset must be > 0",
		},
		{
			name:           "6. Ошибка авторизации (StatusUnauthorized)",
			accessToken:    "bad_token",
			req:            SearchRequest{},
			expectedError:  true,
			expectedErrMsg: "bad AccessToken",
		},
		{
			name:        "7. Ошибка сервера: Некорректный OrderField",
			accessToken: "valid_token",
			req: SearchRequest{
				OrderField: "InvalidField",
			},
			expectedError:  true,
			expectedErrMsg: "OrderFeld InvalidField invalid",
		},
		{
			name:        "8. Ошибка сервера: Недопустимый OrderBy",
			accessToken: "valid_token",
			req: SearchRequest{
				OrderBy: 5,
			},
			expectedError:  true,
			expectedErrMsg: "unknown bad request error",
		},
		{
			name:        "9. Ошибка сервера: 400 Bad Request возвращает невалидный JSON",
			accessToken: "valid_token",
			req: SearchRequest{
				Limit: 10,
			},
			badRequestText: true,
			expectedError:  true,
			expectedErrMsg: "cant unpack error json",
		},
		{
			name:           "10. Ошибка сервера: StatusInternalServerError",
			accessToken:    "valid_token",
			req:            SearchRequest{},
			serverError:    true,
			expectedError:  true,
			expectedErrMsg: "SearchServer fatal error",
		},
		{
			name:           "11. Ошибка сервера: Возврат невалидного JSON результата",
			accessToken:    "valid_token",
			req:            SearchRequest{},
			invalidJSON:    true,
			expectedError:  true,
			expectedErrMsg: "cant unpack result json",
		},
		{
			name:           "12. Ошибка клиента: Превышение таймаута",
			accessToken:    "valid_token",
			req:            SearchRequest{},
			timeoutError:   true,
			expectedError:  true,
			expectedErrMsg: "timeout for",
		},
		{
			name:        "13. Сортировка по Name (Убывание)",
			accessToken: "valid_token",
			req: SearchRequest{
				OrderField: "Name",
				OrderBy:    -1,
			},
			expectedError: false,
		},
		{
			name:        "14. Сортировка по Age (Возрастание)",
			accessToken: "valid_token",
			req: SearchRequest{
				OrderField: "Age",
				OrderBy:    1,
			},
			expectedError: false,
		},
		{
			name:        "15. Сортировка по Age (Убывание)",
			accessToken: "valid_token",
			req: SearchRequest{
				OrderField: "Age",
				OrderBy:    -1,
			},
			expectedError: false,
		},
		{
			name:        "16. Сортировка по Id (Возрастание)",
			accessToken: "valid_token",
			req: SearchRequest{
				OrderField: "Id",
				OrderBy:    1,
			},
			expectedError: false,
		},
		{
			name:        "17. Сортировка по Id (Убывание)",
			accessToken: "valid_token",
			req: SearchRequest{
				OrderField: "Id",
				OrderBy:    -1,
			},
			expectedError: false,
		},
		{
			name:        "18. Офсет превышает длину данных",
			accessToken: "valid_token",
			req: SearchRequest{
				Offset: 100,
			},
			expectedError: false,
		},
		{
			name:           "19. Ошибка файловой системы: XML отсутствует",
			accessToken:    "valid_token",
			req:            SearchRequest{},
			deleteXML:      true,
			expectedError:  true,
			expectedErrMsg: "SearchServer fatal error",
		},
		{
			name:           "20. Ошибка парсинга: Битый XML файл",
			accessToken:    "valid_token",
			req:            SearchRequest{},
			corruptXML:     true,
			expectedError:  true,
			expectedErrMsg: "SearchServer fatal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.deleteXML {
				os.Remove("dataset.xml")
			} else if tt.corruptXML {
				os.WriteFile("dataset.xml", []byte("<root><row>broken"), 0644)
			} else {
				os.WriteFile("dataset.xml", []byte(testXMLData), 0644)
			}

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("AccessToken") != "valid_token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if tt.serverError {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if tt.invalidJSON {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{broken-json`))
					return
				}
				if tt.badRequestText {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`plain text error instead of JSON`))
					return
				}
				if tt.timeoutError {
					time.Sleep(60 * time.Millisecond)
					return
				}
				SearchServer(w, r)
			}))
			defer ts.Close()

			clientStruct := &SearchClient{
				AccessToken: tt.accessToken,
				URL:         ts.URL,
			}

			_, err := clientStruct.FindUsers(tt.req)
			if (err != nil) != tt.expectedError {
				t.Fatalf("Ожидалась ошибка: %v, получили: %v", tt.expectedError, err)
			}
			if tt.expectedError && err != nil {
				if !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Ожидали ошибку с текстом %q, получили %q", tt.expectedErrMsg, err.Error())
				}
			}
		})
	}
}
