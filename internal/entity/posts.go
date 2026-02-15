package entity

import "time"

// Post is a data model for users table
type Post struct {
	Id          int       `json:"Id"`
	Title       string    `json:"Title"`
	Description string    `json:"Description"`
	Slug        string    `json:"Slug"`
	Content     string    `json:"Content"`
	AuthorName  string    `json:"AuthorName"`
	Tags        []string  `json:"Tags"`
	PublishedAt time.Time `json:"PublishedAt"`
}
