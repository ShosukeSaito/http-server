package main

import (
	"log"
	"net/http"
)

type Character struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Level int    `json:"level"`
}

type CreateCharacterRequest struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

// 引数としてCharacterStoreを受け取る。
// 既存のHTTPメソッドとURLパスは変更しない。
func newRouter(store *CharacterStore) *http.ServeMux {
	mux := http.NewServeMux()
	// NewCharacterHandlerを使用してHandlerを生成する。
	handler := NewCharacterHandler(store)

	// GETのルーティングに新しいGetByIDメソッドを登録する。
	mux.HandleFunc(
		"GET /characters/{id}",
		handler.GetByID,
	)
	// POSTのルーティングに新しいCreateメソッドを登録する。
	mux.HandleFunc(
		"POST /characters",
		handler.Create,
	)

	return mux
}

func main() {
	log.Println("Server starting on http://localhost:8080")
	store := NewCharacterStore()
	mux := newRouter(store)

	err := http.ListenAndServe("127.0.0.1:8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
