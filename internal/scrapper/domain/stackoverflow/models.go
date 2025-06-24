package stackoverflow

import "time"

type Answer struct {
	AnswerID     int    `json:"answer_id"`
	QuestionID   int    `json:"question_id"`
	Body         string `json:"body"`
	Owner        User   `json:"owner"`
	CreationDate int64  `json:"creation_date"`
}

type Comment struct {
	CommentID int       `json:"comment_id"`
	PostID    int       `json:"post_id"`
	Body      string    `json:"body"`
	Owner     User      `json:"owner"`
	Creation  time.Time `json:"creation_date"`
}

type User struct {
	DisplayName string `json:"display_name"`
}

func (a Answer) GetCreationTime() time.Time {
	return time.Unix(a.CreationDate, 0)
}
