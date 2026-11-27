package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

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

func TestWaitGroup(t *testing.T) {
	// ① WaitGroupを定義する
	var wg sync.WaitGroup

	// ② forループで3つのgoroutineを起動する
	for i := 1; i <= 3; i++ {
		//    各goroutineで以下の処理を行う
		wg.Go(func() {
			//    - 開始ログの出力
			t.Logf("task %d started", i)
			//    - 100ミリ秒の待機
			time.Sleep(100 * time.Millisecond)
			//    - 終了ログの出力
			t.Logf("task %d finished", i)
		})
	}

	// ③ すべてのgoroutineの終了を待機する
	wg.Wait()

	// ④ 全処理完了のログを出力する
	t.Log("all tasks finished")
}
