package dto

type AddLinkRequestDTO struct {
	Link    string   `json:"link"`
	Tags    []string `json:"tags"`
	Filters []string `json:"filters"`
}

type DeleteLinkRequestDTO struct {
	Link string `json:"link"`
}

type LinkResponseDTO struct {
	ID      int64    `json:"id"`
	URL     string   `json:"url"`
	Tags    []string `json:"tags"`
	Filters []string `json:"filters"`
}

type ListLinksResponseDTO struct {
	Links []LinkDTO `json:"links"`
	Size  int       `json:"size"`
}
