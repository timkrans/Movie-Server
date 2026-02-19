package utils

import (
    "archive/zip"
    "fmt"
    "io"
    "mime/multipart"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type LocalStorage struct {
    BasePath string
}

func (ls LocalStorage) UploadFile(file multipart.File, header *multipart.FileHeader) (string, error) {
    os.MkdirAll(ls.BasePath, os.ModePerm)
    filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
    fullPath := filepath.Join(ls.BasePath, filename)
    out, err := os.Create(fullPath)
    if err != nil {
        return "", err
    }
    defer out.Close()
    _, err = io.Copy(out, file)
    if err != nil {
        return "", err
    }
    return fullPath, nil
}

func ExtractZip(zipPath string, destDir string) (string, string, error) {
    r, err := zip.OpenReader(zipPath)
    if err != nil {
        return "", "", err
    }
    defer r.Close()

    os.MkdirAll(destDir, os.ModePerm)

    var videoPath string
    var imagePath string

    for _, f := range r.File {
        rc, err := f.Open()
        if err != nil {
            continue
        }

        outPath := filepath.Join(destDir, f.Name)
        if f.FileInfo().IsDir() {
            os.MkdirAll(outPath, os.ModePerm)
            rc.Close()
            continue
        }

        os.MkdirAll(filepath.Dir(outPath), os.ModePerm)
        outFile, err := os.Create(outPath)
        if err != nil {
            rc.Close()
            continue
        }

        io.Copy(outFile, rc)
        outFile.Close()
        rc.Close()

        ext := strings.ToLower(filepath.Ext(outPath))

        if videoPath == "" && (ext == ".mp4" || ext == ".mov") {
            videoPath = outPath
        }

        if imagePath == "" && (ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp") {
            imagePath = outPath
        }
    }

    if videoPath == "" {
        return "", "", fmt.Errorf("no video file found in zip")
    }

    return videoPath, imagePath, nil
}
