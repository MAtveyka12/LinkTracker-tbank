package github

import "time"

type Commit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message   string `json:"message"`
		Author    Author `json:"author"`
		Committer Author `json:"committer"`
	} `json:"commit"`
	URL string `json:"html_url"`
}

type Author struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}

type PullRequest struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	Body      string    `json:"body"`
	URL       string    `json:"html_url"`
}

type User struct {
	Login string `json:"login"`
}
