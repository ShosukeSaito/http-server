package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
