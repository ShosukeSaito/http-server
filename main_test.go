package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestCharacterHandlerTable(t *testing.T) {
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
			store := NewCharacterStore()
			mux := newRouter(store)

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

func TestCreateCharacterHandlerTable(t *testing.T) {
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
			store := NewCharacterStore()
			mux := newRouter(store)

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

func TestCreateCharacterHandlerConcurrent(t *testing.T) {
	store := NewCharacterStore()
	// ルーターを生成する
	mux := newRouter(store)

	// リクエスト数を定義する
	const (
		requestCount       = 10
		initialCount       = 2
		expectedMinID      = 2003
		expectedMaxID      = expectedMinID + requestCount - 1
		expectedTotalCount = initialCount + requestCount
		expectedNextID     = expectedMinID + requestCount
	)

	// 結果格納用のスライスを生成する
	results := make([]Character, requestCount)

	// WaitGroupを定義する
	var wg sync.WaitGroup

	method := http.MethodPost
	path := "/characters"
	jsonBody := `{"name": "Knight", "level": 1}`

	// forループで10個のgoroutineを起動する
	for i := 0; i < requestCount; i++ {
		wg.Go(func() {
			// ===POST APIを実行===
			//    各goroutine内で以下を実行する

			//    - POSTリクエストの生成
			req := httptest.NewRequest(
				method,
				path,
				strings.NewReader(jsonBody),
			)
			//    - Content-Typeの設定
			req.Header.Set("Content-Type", "application/json")
			//    - ResponseRecorderの生成
			recorder := httptest.NewRecorder()
			//    - APIの実行
			mux.ServeHTTP(recorder, req)
			//    - ステータスコードの検証
			if recorder.Code != http.StatusCreated {
				t.Errorf(
					"expected status %d, got %d",
					http.StatusCreated,
					recorder.Code,
				)
				return
			}
			//    - JSONレスポンスのデコード
			var character Character
			err := json.Unmarshal(
				recorder.Body.Bytes(),
				&character,
			)

			if err != nil {
				t.Errorf("failed to decode response: %v", err)
				return
			}
			//    - 結果をスライスに格納
			results[i] = character
		})
	}

	// 全goroutineの終了を待機する
	wg.Wait()

	seenIDs := make(map[int]bool)

	for _, character := range results {
		// IDの重複がないことを検証する
		if seenIDs[character.ID] {
			t.Errorf("duplicate character ID found: %d", character.ID)
		}

		seenIDs[character.ID] = true

		// IDが2003〜2012の範囲内であることを検証する
		if character.ID < expectedMinID ||
			character.ID > expectedMaxID {
			t.Errorf("character ID %d is out of expected range", character.ID)
		}
	}

	// charactersの要素数を検証する
	if len(store.characters) != expectedTotalCount {
		t.Errorf("expected characters length %d, got %d", expectedTotalCount, len(store.characters))
	}

	// nextIDを検証する
	if store.nextID != expectedNextID {
		t.Errorf("expected nextCharacterID %d, got %d", expectedNextID, store.nextID)
	}
}

func TestCreateAndGetCharacter(t *testing.T) {
	// テスト環境の準備
	store := NewCharacterStore()
	mux := newRouter(store)

	postPath := "/characters"
	getPath := "/characters/2003"
	jsonBody := `{"name": "Knight", "level": 15}`

	// 期待値の定義
	expectedContentType := "application/json"
	expectedLocation := "/characters/2003"
	expectedCharacter := Character{
		ID:    2003,
		Name:  "Knight",
		Level: 15,
	}

	postReq := httptest.NewRequest(
		http.MethodPost,
		postPath,
		strings.NewReader(jsonBody),
	)

	postReq.Header.Set("Content-Type", "application/json")

	postRecorder := httptest.NewRecorder()

	// POSTリクエストの実行
	mux.ServeHTTP(postRecorder, postReq)

	// POSTレスポンスの検証
	if postRecorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			postRecorder.Code,
		)
	}

	postActualContentType := postRecorder.Header().Get("Content-Type")
	if postActualContentType != expectedContentType {
		t.Errorf(
			"expected Content-Type %q, got %q",
			expectedContentType,
			postActualContentType,
		)
	}

	postActualLocation := postRecorder.Header().Get("Location")
	if postActualLocation != expectedLocation {
		t.Errorf(
			"expected Location %q, got %q",
			expectedLocation,
			postActualLocation,
		)
	}

	var postActualCharacter Character
	err := json.Unmarshal(
		postRecorder.Body.Bytes(),
		&postActualCharacter,
	)

	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if postActualCharacter != expectedCharacter {
		t.Errorf("expected %+v, got %+v", expectedCharacter, postActualCharacter)
	}

	getReq := httptest.NewRequest(
		http.MethodGet,
		getPath,
		nil,
	)
	getRecorder := httptest.NewRecorder()

	// GETリクエストの実行
	mux.ServeHTTP(getRecorder, getReq)

	// GETレスポンスの検証
	if getRecorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			getRecorder.Code,
		)
	}

	getActualContentType := getRecorder.Header().Get("Content-Type")
	if getActualContentType != expectedContentType {
		t.Errorf(
			"expected Content-Type %q, got %q",
			expectedContentType,
			getActualContentType,
		)
	}

	var getActualCharacter Character

	err = json.Unmarshal(
		getRecorder.Body.Bytes(),
		&getActualCharacter,
	)

	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if getActualCharacter != expectedCharacter {
		t.Errorf("expected %+v, got %+v", expectedCharacter, getActualCharacter)
	}
}

func TestCharacterStoreIndependence(t *testing.T) {
	storeA := NewCharacterStore()
	muxA := newRouter(storeA)

	storeB := NewCharacterStore()
	muxB := newRouter(storeB)

	method := http.MethodPost
	path := "/characters"
	jsonBody := `{"name": "Knight", "level": 1}`

	reqA := httptest.NewRequest(
		method,
		path,
		strings.NewReader(jsonBody),
	)

	reqA.Header.Set("Content-Type", "application/json")

	recA := httptest.NewRecorder()

	muxA.ServeHTTP(recA, reqA)

	if recA.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recA.Code,
		)
	}

	var charaA Character
	if err := json.Unmarshal(
		recA.Body.Bytes(),
		&charaA,
	); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	reqB := httptest.NewRequest(
		method,
		path,
		strings.NewReader(jsonBody),
	)

	reqB.Header.Set("Content-Type", "application/json")

	recB := httptest.NewRecorder()

	muxB.ServeHTTP(recB, reqB)

	if recB.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recB.Code,
		)
	}

	var charaB Character
	if err := json.Unmarshal(
		recB.Body.Bytes(),
		&charaB,
	); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if charaB.ID != 2003 {
		t.Errorf("expected storeB first character ID to be 2003, got %d", charaB.ID)

	}
}

// func TestCharacterStoreIndependence(t *testing.T) {
// 	const requestCount = 2

// 	// results := make(map[int]Character)
// 	results := make([]Character, requestCount*requestCount)

// 	var wg sync.WaitGroup

// 	method := http.MethodPost
// 	path := "/characters"
// 	jsonBody := `{"name": "Knight", "level": 1}`

// 	for i := 0; i < requestCount; i++ {
// 		store := NewCharacterStore()
// 		mux := newRouter(store)

// 		wg.Go(func() {
// 			for j := 0; j < requestCount; j++ {

// 				req := httptest.NewRequest(
// 					method,
// 					path,
// 					strings.NewReader(jsonBody),
// 				)

// 				req.Header.Set("Content-Type", "application/json")

// 				recorder := httptest.NewRecorder()

// 				mux.ServeHTTP(recorder, req)

// 				if recorder.Code != http.StatusCreated {
// 					t.Errorf(
// 						"expected status %d, got %d",
// 						http.StatusCreated,
// 						recorder.Code,
// 					)
// 					return
// 				}

// 				var character Character
// 				err := json.Unmarshal(
// 					recorder.Body.Bytes(),
// 					&character,
// 				)

// 				if err != nil {
// 					t.Errorf("failed to decode response: %v", err)
// 					return
// 				}

// 				n := i*requestCount + j

// 				results[n] = character
// 			}
// 		})
// 	}

// 	wg.Wait()

// 	for idx, character := range results {
// 		t.Logf("[%d] got %+v", idx, character)
// 	}
// }
