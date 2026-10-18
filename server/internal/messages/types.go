package messages

const MaxLength = 500

type DTO struct {
	ID        string `json:"id"`
	RoomID    string `json:"roomId"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}
