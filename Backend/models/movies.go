package models

type Movie struct {
	ID                uint   `gorm:"primaryKey"`
	Title             string
	VideoFilePath     string
	CoverImageFilePath string
}

type MovieInput struct {
	Title             string `json:"title" binding:"required"`
	VideoFilePath     string `json:"video_file_path" binding:"required"`
	CoverImageFilePath string `json:"cover_image_file_path"`
}
