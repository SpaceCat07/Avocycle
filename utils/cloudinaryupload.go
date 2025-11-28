package utils

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
    "strings" // Pastikan package strings diimpor
	"time"

    "github.com/cloudinary/cloudinary-go/v2"
    "github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

const uploadTimeout = 30 * time.Second
const maxUploadSize = 10 << 20 // 10 MB

var allowedExts = map[string]bool{
    ".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

func AsyncUploadOptionalImage(file *multipart.FileHeader, folder string) (string, string, error) {
    if file == nil {
        return "", "", nil
    }
    if file.Size > maxUploadSize {
        return "", "", fmt.Errorf("ukuran file maksimal %d MB", maxUploadSize/(1<<20))
    }
    ext := strings.ToLower(filepath.Ext(file.Filename))
    if !allowedExts[ext] {
        return "", "", fmt.Errorf("format file tidak didukung")
    }

    tmpPath := fmt.Sprintf("./tmp_%d_%s", time.Now().UnixNano(), file.Filename)
    if err := saveMultipart(file, tmpPath); err != nil {
        return "", "", err
    }

    resultCh := make(chan *uploader.UploadResult, 1)
    errCh := make(chan error, 1)

    go func() {
        defer func() {
            os.Remove(tmpPath)
            close(resultCh)
            close(errCh)
        }()

        cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
        if err != nil {
            errCh <- err
            return
        }

        ctx, cancel := context.WithTimeout(context.Background(), uploadTimeout)
        defer cancel()

        res, err := cld.Upload.Upload(ctx, tmpPath, uploader.UploadParams{
            Folder:         fmt.Sprintf("avocycle/%s", folder),
            UniqueFilename: &[]bool{true}[0],
            Overwrite:      &[]bool{false}[0],
        })
        if err != nil {
            errCh <- err
            return
        }
        resultCh <- res
    }()

    select {
    case res := <-resultCh:
        if res == nil {
            return "", "", nil
        }
        return res.SecureURL, res.PublicID, nil
    case err := <-errCh:
        return "", "", err
    case <-time.After(uploadTimeout):
        return "", "", fmt.Errorf("upload timeout")
    }
}

func saveMultipart(file *multipart.FileHeader, dst string) error {
    src, err := file.Open()
    if err != nil {
        return err
    }
    defer src.Close()

    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = out.ReadFrom(src)
    return err
}

func DeleteCloudinaryAsset(publicID string) error {
    if publicID == "" {
        return nil
    }

    cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
    if err != nil {
        return fmt.Errorf("cloudinary init failed: %w", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), uploadTimeout)
    defer cancel()

    _, err = cld.Upload.Destroy(ctx, uploader.DestroyParams{
        PublicID:   publicID,
        Invalidate: &[]bool{true}[0],
    })
    return err
}