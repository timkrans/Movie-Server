package models

type Movie struct {
	ID       uint   `gorm:"primaryKey"`
	Title    string
	FilePath string
}

type MovieInput struct {
	Title    string `json:"title" binding:"required"`
	FilePath string `json:"file_path" binding:"required"`
}
