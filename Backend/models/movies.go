package models

type Movie struct {
    ID                 uint   `json:"id" gorm:"primaryKey"`
    Title              string `json:"title"`
    VideoFilePath      string `json:"video_file_path"`
    CoverImageFilePath string `json:"cover_image_file_path"`
    HLSPath            string `json:"hls_path"`
}

type MovieInput struct {
    Title string `form:"title" binding:"required"`
}
