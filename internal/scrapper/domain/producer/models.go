package producer

type LinkUpdate struct {
	ID          int64  `json:"id"`
	URL         string `json:"url"`
	Description string `json:"description"`
	UserID      int64  `json:"user_id"`
	Type        string `json:"type"`
}
