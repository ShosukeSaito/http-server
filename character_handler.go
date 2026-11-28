package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

type CharacterHandler struct {
	store *CharacterStore
}

func NewCharacterHandler(store *CharacterStore) *CharacterHandler {
	return &CharacterHandler{
		store: store,
	}
}

func (h *CharacterHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	// パスパラメータの取得
	idStr := r.PathValue("id")
	// IDの整数変換
	id, err := strconv.Atoi(idStr)
	// 変換失敗時の400レスポンス
	if err != nil {
		http.Error(w, "Invalid character ID", http.StatusBadRequest)
		return
	}
	// Mutexのロック Storeに委譲
	h.store.mu.Lock()
	// mapへのアクセス Storeに委譲
	character, exists := h.store.GetByID(id)
	// Mutexのアンロック Storeに委譲
	h.store.mu.Unlock()
	// 存在しない場合の404レスポンス
	if !exists {
		http.Error(w, "Character not found", http.StatusNotFound)
		return
	}
	// JSONレスポンスの返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(character); err != nil {
		log.Printf("failed to encode character: %v", err)
	}
}

func (h *CharacterHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	// 	JSONのデコード
	var req CreateCharacterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// Nameのバリデーション
	if strings.TrimSpace(req.Name) == "" || utf8.RuneCountInString(req.Name) > 20 {
		http.Error(w, "Invalid character name", http.StatusBadRequest)
		return
	}
	// Levelのバリデーション
	if req.Level < 1 || req.Level > 100 {
		http.Error(w, "Invalid character level", http.StatusBadRequest)
		return
	}
	// Mutexの操作	Storeに委譲
	h.store.mu.Lock()
	// IDの採番	Storeに委譲
	id := h.store.nextID
	// キャラクターの生成・登録	Storeに委譲
	character := h.store.Create(req.Name, req.Level)

	h.store.characters[id] = character

	// 次回採番値の更新	Storeに委譲
	h.store.nextID++

	h.store.mu.Unlock()
	// Locationヘッダーの設定
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/characters/%d", character.ID))
	// 201 Createdの返却
	w.WriteHeader(http.StatusCreated)
	// JSONレスポンスの返却
	if err := json.NewEncoder(w).Encode(character); err != nil {
		log.Printf("failed to encode character: %v", err)
	}
}
