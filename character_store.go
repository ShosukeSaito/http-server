package main

import "sync"

type CharacterStore struct {
	Characters map[int]Character
	NextID     int
	Mu         sync.Mutex
}

func NewCharacterStore() *CharacterStore {
	// 1. 呼び出すたびに新しい map を生成する
	newChars := make(map[int]Character, len(characters))

	// 2. 初期データ（characters）の中身を新しい map にコピーする
	for k, v := range characters {
		newChars[k] = v
	}

	// 3. 実体を生成してポインタを返す
	return &CharacterStore{
		Characters: newChars,
		NextID:     nextCharacterID, // 2003
	}
}

func (s *CharacterStore) GetByID(id int) (Character, bool) {
	// 	Mutexによる排他制御を行う。
	s.Mu.Lock()
	defer s.Mu.Unlock()

	// 指定されたIDのキャラクターを取得する。
	// キャラクターと存在確認結果を返す。
	character, exists := s.Characters[id]

	return character, exists
}

func (s *CharacterStore) Create(name string, level int) Character {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	// nextIDを使用してキャラクターを生成する。
	character := Character{
		ID:    s.NextID,
		Name:  name,
		Level: level,
	}

	// mapへ登録する。
	s.Characters[character.ID] = character
	// 次回採番値を更新する。
	s.NextID++

	// 登録したキャラクターを返す。
	return character
}
