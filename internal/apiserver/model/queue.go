package model

import (
	"encoding/json"
	"time"
)

// TranslationTask translation task
type TranslationTask struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	SourceLang    string    `json:"source_lang"`
	TargetLang    string    `json:"target_lang"`
	SourceContent string    `json:"source_content"`
	CreatedAt     time.Time `json:"created_at"`
}

func (t *TranslationTask) GetID() string {
	return t.ID
}

func (t *TranslationTask) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t *TranslationTask) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, t)
}
