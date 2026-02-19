package utils

import "mime/multipart"

type Storage interface {
    UploadFile(file multipart.File, header *multipart.FileHeader) (string, error)
}
