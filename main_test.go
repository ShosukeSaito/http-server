package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCharacterHandler(t *testing.T) {
	// ① HTTPリクエストを作成
	req := httptest.NewRequest(
		http.MethodGet,
		"/characters/2001",
		nil,
	)

	// ② レスポンスを記録するオブジェクトを作成
	recorder := httptest.NewRecorder()

	// ③ パスパラメータを設定
	req.SetPathValue("id", "2001")

	// ④ テスト対象のハンドラーを実行
	characterHandler(recorder, req)

	// ⑤ レスポンスのステータスコードを確認
	if recorder.Code != http.StatusOK {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	// レスポンスボディを取得する
	body := recorder.Body.String()

	// ログを出力
	t.Logf("response body: %s", body)

	var character Character

	err := json.Unmarshal(recorder.Body.Bytes(), &character)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if character.ID != 2001 {
		t.Errorf("expected ID 2001, got %d", character.ID)
	}

	if character.Name != "Hero" {
		t.Errorf("expected name Hero, got %s", character.Name)
	}

	if character.Level != 10 {
		t.Errorf("expected level 10, got %d", character.Level)
	}
}

func TestCharacterHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/characters/9999",
		nil,
	)

	req.SetPathValue("id", "9999")

	recorder := httptest.NewRecorder()

	characterHandler(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}

	expected := "Character not found\n"
	actual := recorder.Body.String()

	if expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestCharacterHandlerBadRequest(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/characters/abc",
		nil,
	)

	req.SetPathValue("id", "abc")

	recorder := httptest.NewRecorder()

	characterHandler(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	expected := "Invalid character ID\n"
	actual := recorder.Body.String()

	if expected != actual {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestCharacterHandlerTable(t *testing.T) {
	mux := newRouter()

	tests := []struct {
		// テストケースの識別情報
		name string

		// 入力値
		method string
		path   string

		// 期待値
		expectedStatus      int
		expectedBody        string
		expectedCharacter   Character
		expectedContentType string
	}{
		{
			name:           "existing character 2001",
			method:         http.MethodGet,
			path:           "/characters/2001",
			expectedStatus: http.StatusOK,
			expectedCharacter: Character{
				ID:    2001,
				Name:  "Hero",
				Level: 10,
			},
			expectedContentType: "application/json",
		},
		{
			name:           "existing character 2002",
			method:         http.MethodGet,
			path:           "/characters/2002",
			expectedStatus: http.StatusOK,
			expectedCharacter: Character{
				ID:    2002,
				Name:  "Mage",
				Level: 8,
			},
			expectedContentType: "application/json",
		},
		{
			name:                "character not found",
			method:              http.MethodGet,
			path:                "/characters/9999",
			expectedStatus:      http.StatusNotFound,
			expectedBody:        "Character not found\n",
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "invalid character ID",
			method:              http.MethodGet,
			path:                "/characters/abc",
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        "Invalid character ID\n",
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "404 page not found",
			method:              http.MethodGet,
			path:                "/unknown",
			expectedStatus:      http.StatusNotFound,
			expectedBody:        "404 page not found\n",
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "Method Not Allowed",
			method:              http.MethodPost,
			path:                "/characters/2001",
			expectedStatus:      http.StatusMethodNotAllowed,
			expectedBody:        "Method Not Allowed\n",
			expectedContentType: "text/plain; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(
				tt.method,
				tt.path,
				nil,
			)

			recorder := httptest.NewRecorder()

			mux.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatus,
					recorder.Code,
				)
			}

			actualContentType := recorder.Header().Get("Content-Type")
			if actualContentType != tt.expectedContentType {
				t.Errorf(
					"expected Content-Type %q, got %q",
					tt.expectedContentType,
					actualContentType,
				)
			}

			if tt.expectedStatus == http.StatusOK {
				var actualCharacter Character

				err := json.Unmarshal(
					recorder.Body.Bytes(),
					&actualCharacter,
				)

				if err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if actualCharacter != tt.expectedCharacter {
					t.Errorf("expected %+v, got %+v", tt.expectedCharacter, actualCharacter)
				}
			} else {
				actualBody := recorder.Body.String()

				if actualBody != tt.expectedBody {
					t.Errorf("expected %q, got %q", tt.expectedBody, actualBody)
				}
			}
		})
	}
}

func TestCreateCharacterHandler(t *testing.T) {
	// 初期化
	resetCharacterState()

	// ① ルーターの生成
	mux := newRouter()

	// ② JSONリクエストボディの用意
	path := "/characters"
	jsonBody := `{"name":"Knight","level":1}`

	// ③ POSTリクエストの生成
	req := httptest.NewRequest(
		http.MethodPost,
		path,
		strings.NewReader(jsonBody),
	)

	req.Header.Set("Content-Type", "application/json")

	// ④ ResponseRecorderの生成
	recorder := httptest.NewRecorder()

	// ⑤ リクエストの処理
	mux.ServeHTTP(recorder, req)

	// ⑥ ステータスコードの検証
	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	// ⑦ Content-Typeの検証
	actualContentType := recorder.Header().Get("Content-Type")
	expectedContentType := "application/json"

	if actualContentType != expectedContentType {
		t.Errorf(
			"expected Content-Type %q, got %q",
			expectedContentType,
			actualContentType,
		)
	}

	// ⑧ Locationの検証
	actualLocation := recorder.Header().Get("Location")
	expectedLocation := "/characters/2003"

	if actualLocation != expectedLocation {
		t.Errorf(
			"expected Location %q, got %q",
			expectedLocation,
			actualLocation,
		)
	}

	// ⑨ JSONレスポンスのデコード
	var actualCharacter Character
	expectedCharacter := Character{
		ID:    2003,
		Name:  "Knight",
		Level: 1,
	}

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&actualCharacter,
	)

	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// ⑩ Characterの各フィールドの検証
	if actualCharacter != expectedCharacter {
		t.Errorf("expected %+v, got %+v", expectedCharacter, actualCharacter)
	}
}

func resetCharacterState() {
	charactersMu.Lock()
	defer charactersMu.Unlock()

	characters = map[int]Character{
		2001: {
			ID:    2001,
			Name:  "Hero",
			Level: 10,
		},
		2002: {
			ID:    2002,
			Name:  "Mage",
			Level: 8,
		},
	}

	nextCharacterID = 2003
}

func TestCreateCharacterHandlerTable(t *testing.T) {
	mux := newRouter()

	// テストケースの定義
	tests := []struct {
		// テストケースの識別情報
		name string

		// JSONリクエストボディ
		jsonBody string

		// 期待値
		expectedStatus      int
		expectedContentType string
		expectedLocation    string
		expectedBody        string
		expectedCharacter   Character
	}{
		{
			// テストケースごとの実行情報
			name:                "creating Knight and level is 1",
			jsonBody:            `{"name": "Knight", "level": 1}`,
			expectedStatus:      http.StatusCreated,
			expectedContentType: "application/json",
			expectedLocation:    "/characters/2003",
			expectedCharacter: Character{
				ID:    2003,
				Name:  "Knight",
				Level: 1,
			},
		},
		{
			name:                "invalid character name is empty",
			jsonBody:            `{"name": "", "level": 1}`,
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedLocation:    "",
			expectedBody:        "Invalid character name\n",
		},
		{
			name:                "invalid character name is blank",
			jsonBody:            `{"name": " ", "level": 1}`,
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedLocation:    "",
			expectedBody:        "Invalid character name\n",
		},
		{
			name:                "creating character and character name length is 20",
			jsonBody:            `{"name": "abcdefghijklmnopqrst", "level": 1}`,
			expectedStatus:      http.StatusCreated,
			expectedContentType: "application/json",
			expectedLocation:    "/characters/2003",
			expectedCharacter: Character{
				ID:    2003,
				Name:  "abcdefghijklmnopqrst",
				Level: 1,
			},
		},
		{
			name:                "invalid character name exceeds maximum length",
			jsonBody:            `{"name": "abcdefghijklmnopqrstu", "level": 1}`,
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedLocation:    "",
			expectedBody:        "Invalid character name\n",
		},
		{
			name:                "creating character and character name is japanese",
			jsonBody:            `{"name": "あいうえおかきく", "level": 1}`,
			expectedStatus:      http.StatusCreated,
			expectedContentType: "application/json",
			expectedLocation:    "/characters/2003",
			expectedCharacter: Character{
				ID:    2003,
				Name:  "あいうえおかきく",
				Level: 1,
			},
		},
		{
			name:                "invalid character level is below minimum",
			jsonBody:            `{"name": "Knight", "level": 0}`,
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedLocation:    "",
			expectedBody:        "Invalid character level\n",
		},
		{
			name:                "creating knight and character level is 100",
			jsonBody:            `{"name": "Knight", "level": 100}`,
			expectedStatus:      http.StatusCreated,
			expectedContentType: "application/json",
			expectedLocation:    "/characters/2003",
			expectedCharacter: Character{
				ID:    2003,
				Name:  "Knight",
				Level: 100,
			},
		},
		{
			name:                "invalid character level exceeds maximum",
			jsonBody:            `{"name": "Knight", "level": 101}`,
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedLocation:    "",
			expectedBody:        "Invalid character level\n",
		},
		{
			name:                "invalid JSON syntax",
			jsonBody:            `{"name": "Knight", "level": 1`,
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedLocation:    "",
			expectedBody:        "Invalid request body\n",
		},
		{
			name:                "invalid character level is a string",
			jsonBody:            `{"name": "Knight", "level": "abc"}`,
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "text/plain; charset=utf-8",
			expectedLocation:    "",
			expectedBody:        "Invalid request body\n",
		},
	}

	method := http.MethodPost
	path := "/characters"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCharacterState()

			req := httptest.NewRequest(
				method,
				path,
				strings.NewReader(tt.jsonBody),
			)

			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			mux.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatus,
					recorder.Code,
				)
			}

			actualContentType := recorder.Header().Get("Content-Type")
			if actualContentType != tt.expectedContentType {
				t.Errorf(
					"expected Content-Type %q, got %q",
					tt.expectedContentType,
					actualContentType,
				)
			}

			actualLocation := recorder.Header().Get("Location")
			if actualLocation != tt.expectedLocation {
				t.Errorf(
					"expected Location %q, got %q",
					tt.expectedLocation,
					actualLocation,
				)
			}

			if tt.expectedStatus == http.StatusCreated {
				var actualCharacter Character

				err := json.Unmarshal(
					recorder.Body.Bytes(),
					&actualCharacter,
				)

				if err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if actualCharacter != tt.expectedCharacter {
					t.Errorf("expected %+v, got %+v", tt.expectedCharacter, actualCharacter)
				}
			} else {
				actualBody := recorder.Body.String()

				if actualBody != tt.expectedBody {
					t.Errorf("expected %q, got %q", tt.expectedBody, actualBody)
				}
			}
		})
	}
}
