package domain

type Link struct {
	URL    string
	LinkID int64
	UserID int64
}

var UserTrackData = make(map[int64]Link)
