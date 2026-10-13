package models

type Movie struct {
	ID                 uint   `json:"id" gorm:"primaryKey"`
	Title              string `json:"title"`
	VideoFilePath      string `json:"video_file_path"`
	CoverImageFilePath string `json:"cover_image_file_path"`
	HLSPath            string `json:"hls_path"`
	Status             string `json:"status" gorm:"default:processing"` // processing | ready | failed
	CollectionID       *uint  `json:"collection_id"`
	CollectionOrder    *int   `json:"collection_order"`
}

type MovieInput struct {
	Title          string `form:"title" binding:"required"`
	CollectionID   *uint  `form:"collection_id"`
	CollectionOrder *int  `form:"collection_order"`
}

type LiveStream struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Title     string `json:"title"`
	StreamKey string `json:"stream_key" gorm:"uniqueIndex"`
	Status    string `json:"status" gorm:"default:idle"` // idle | live | ended
	HLSDir    string `json:"hls_dir"`
	CreatedAt int64  `json:"created_at"`
}

type LiveStreamInput struct {
	Title string `json:"title" binding:"required"`
}

type Profile struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	AvatarColor string `json:"avatar_color"`
	CreatedAt   int64  `json:"created_at"`
}

type ProfileInput struct {
	Name        string `json:"name" binding:"required"`
	AvatarColor string `json:"avatar_color"`
}

type Collection struct {
	ID                 uint    `json:"id" gorm:"primaryKey"`
	Title              string  `json:"title"`
	CoverImageFilePath string  `json:"cover_image_file_path"`
	Movies             []Movie `json:"movies,omitempty" gorm:"foreignKey:CollectionID"`
}

type CollectionInput struct {
	Title string `form:"title" binding:"required"`
}

type AttachMovieInput struct {
	MovieID uint `json:"movie_id" binding:"required"`
	Order   int  `json:"order"`
}

type Show struct {
	ID                 uint     `json:"id" gorm:"primaryKey"`
	Title              string   `json:"title"`
	CoverImageFilePath string   `json:"cover_image_file_path"`
	Seasons            []Season `json:"seasons,omitempty" gorm:"foreignKey:ShowID"`
}

type ShowInput struct {
	Title string `form:"title" binding:"required"`
}

type Season struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ShowID       uint      `json:"show_id" gorm:"index"`
	SeasonNumber int       `json:"season_number"`
	Title        string    `json:"title"`
	Episodes     []Episode `json:"episodes,omitempty" gorm:"foreignKey:SeasonID"`
}

type SeasonInput struct {
	SeasonNumber int    `json:"season_number" binding:"required"`
	Title        string `json:"title"`
}

type Episode struct {
	ID                 uint   `json:"id" gorm:"primaryKey"`
	SeasonID           uint   `json:"season_id" gorm:"index"`
	EpisodeNumber      int    `json:"episode_number"`
	Title              string `json:"title"`
	VideoFilePath      string `json:"video_file_path"`
	CoverImageFilePath string `json:"cover_image_file_path"`
	HLSPath            string `json:"hls_path"`
	Status             string `json:"status" gorm:"default:processing"`
}

type EpisodeInput struct {
	Title         string `form:"title" binding:"required"`
	EpisodeNumber int    `form:"episode_number" binding:"required"`
}

type WatchProgress struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	ProfileID       uint    `json:"profile_id" gorm:"uniqueIndex:idx_profile_content"`
	ContentType     string  `json:"content_type" gorm:"uniqueIndex:idx_profile_content"` // movie | episode
	ContentID       uint    `json:"content_id" gorm:"uniqueIndex:idx_profile_content"`
	PositionSeconds float64 `json:"position_seconds"`
	DurationSeconds float64 `json:"duration_seconds"`
	Completed       bool    `json:"completed"`
	UpdatedAt       int64   `json:"updated_at"`
}

type ProgressInput struct {
	ContentType     string  `json:"content_type" binding:"required"`
	ContentID       uint    `json:"content_id" binding:"required"`
	PositionSeconds float64 `json:"position_seconds"`
	DurationSeconds float64 `json:"duration_seconds"`
}

// ContinueItem is an enriched progress row for the Continue Watching shelf.
type ContinueItem struct {
	ContentType     string  `json:"content_type"`
	ContentID       uint    `json:"content_id"`
	Title           string  `json:"title"`
	Subtitle        string  `json:"subtitle,omitempty"`
	CoverURL        string  `json:"cover_url"`
	WatchURL        string  `json:"watch_url"`
	PositionSeconds float64 `json:"position_seconds"`
	DurationSeconds float64 `json:"duration_seconds"`
	Progress        float64 `json:"progress"`
	UpdatedAt       int64   `json:"updated_at"`
}
