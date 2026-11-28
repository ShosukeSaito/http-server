package main

import "sync"

type CharacterStore struct {
	characters map[int]Character
	nextID     int
	mu         sync.Mutex
}

func NewCharacterStore() *CharacterStore {
	// 1. 呼び出すたびに新しい map を生成する
	newChars := map[int]Character{
		// 初期データを定義する
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

	// 3. 実体を生成してポインタを返す
	return &CharacterStore{
		characters: newChars,
		nextID:     2003, // 2003
	}
}

func (s *CharacterStore) GetByID(id int) (Character, bool) {
	// 	Mutexによる排他制御を行う。
	s.mu.Lock()
	defer s.mu.Unlock()

	// 指定されたIDのキャラクターを取得する。
	// キャラクターと存在確認結果を返す。
	character, exists := s.characters[id]

	return character, exists
}

func (s *CharacterStore) Create(name string, level int) Character {
	// 	Mutexによる排他制御を行う。
	s.mu.Lock()
	defer s.mu.Unlock()

	// nextIDを使用してキャラクターを生成する。
	character := Character{
		ID:    s.nextID,
		Name:  name,
		Level: level,
	}

	// mapへ登録する。
	s.characters[character.ID] = character
	// 次回採番値を更新する。
	s.nextID++

	// 登録したキャラクターを返す。
	return character
}
