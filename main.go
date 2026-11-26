package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Character struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Level int    `json:"level"`
}

var characters = map[int]Character{
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

func characterHandler(w http.ResponseWriter, r *http.Request) {
	// ① パスパラメータを取得
	idStr := r.PathValue("id")

	// ② 文字列を整数へ変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid character ID", http.StatusBadRequest)
		return
	}

	// ③ キャラクターの存在確認
	character, exists := characters[id]
	if !exists {
		http.Error(w, "Character not found", http.StatusNotFound)
		return
	}

	// ④ 正常系のレスポンス
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(character); err != nil {
		log.Printf("failed to encode character: %v", err)
	}
}

func newRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /characters/{id}",
		characterHandler,
	)

	return mux
}

func main() {
	log.Println("Server starting on http://localhost:8080")
	mux := newRouter()

	err := http.ListenAndServe("127.0.0.1:8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
