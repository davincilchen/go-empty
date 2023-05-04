package models

type EventData struct {
	ID        string          `json:"id"`
	Pubkey    string          `json:"pubkey"`
	CreatedAt int             `json:"created_at"`
	Kind      int             `json:"kind"`
	Content   string          `json:"content"`
	Tags      [][]interface{} `json:"tags"`
	Sig       string          `json:"sig"`
}
