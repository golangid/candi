package storage

import "io"

// UploadParam model
type UploadParam struct {
	ContentType string
	Folder      string
	Filename    string
	File        io.Reader
	Size        int64
}

// Response model
type Response struct {
	Location string `json:"file"`
	Size     int64  `json:"size"`
}
