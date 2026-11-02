package service

import (
	"context"
	"fmt"
	"io"
	"myproject/data"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Service struct {
	minioRepo data.MinIORepo
	dbRepo    data.DBRepo
}

func NewService(minioRepo data.MinIORepo, dbRepo data.DBRepo) *Service {
	return &Service{
		minioRepo: minioRepo,
		dbRepo:    dbRepo,
	}
}

func (s *Service) GetVideo(ctx context.Context, id string) (io.ReadCloser, error) {
	fileName, err := s.dbRepo.GetFileNameByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении имени файла: %v", err)
	}

	return s.minioRepo.GetObject(ctx, fileName)
}

func (s *Service) SaveByURL(ctx context.Context, urlStr string) (string, error) {
	fileName, directURL, err := extractMeta(ctx, urlStr)
	if err != nil {
		return "", err
	}

	exists, err := s.dbRepo.ExistsByName(ctx, fileName)
	if err != nil {
		return "", fmt.Errorf("Ошибка проверки существования файла: %v", err)
	}
	if exists {
		return "", fmt.Errorf("Файл с именем '%s' уже загружен", fileName)
	}

	resp, err := http.Get(directURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ошибка при получении содержимого ссылки: %v", err)
	}
	defer resp.Body.Close()

	_, err = s.minioRepo.Put(
		ctx,
		fileName,
		resp.Body,
		resp.ContentLength,
	)
	if err != nil {
		return "", fmt.Errorf("Ошибка при загрузке файла в MinIO: %v", err)
	}

	id, err := s.dbRepo.Put(ctx, fileName)
	if err != nil {
		err = s.minioRepo.Remove(
			ctx,
			fileName,
		)
		if err != nil {
			return "", fmt.Errorf("Ошибка при загрузке файла в БД и удалении его из MinIO: %v", err)
		}
		return "", fmt.Errorf("Ошибка при загрузке файла в БД: %v", err)
	}

	return id, nil
}

func (s *Service) SaveByFile(ctx context.Context, fileName string, reader io.Reader, size int64) (string, error) {
	_, err := s.minioRepo.Put(ctx, fileName, reader, size)
	if err != nil {
		return "", fmt.Errorf("Ошибка при загрузке файла в MinIO: %v", err)
	}

	exists, err := s.dbRepo.ExistsByName(ctx, fileName)
	if err != nil {
		return "", fmt.Errorf("Ошибка проверки существования файла: %v", err)
	}
	if exists {
		return "", fmt.Errorf("Файл с именем '%s' уже загружен", fileName)
	}

	id, err := s.dbRepo.Put(ctx, fileName)
	if err != nil {
		err = s.minioRepo.Remove(
			ctx,
			fileName,
		)
		if err != nil {
			return "", fmt.Errorf("Ошибка при загрузке файла в БД и удалении его из MinIO: %v", err)
		}
		return "", fmt.Errorf("Ошибка при загрузке файла в БД: %v", err)
	}

	return id, nil
}

func (s *Service) GetAll(ctx context.Context) ([]data.Video, error) {
	return s.dbRepo.GetAll(ctx)
}

func extractMeta(ctx context.Context, pageURL string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "yt-dlp", "--get-filename", "--get-url", "-f", "best", pageURL)
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("yt-dlp не смог извлечь данные: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return "", "", fmt.Errorf("yt-dlp вернул неожиданный результат: %q", string(output))
	}

	rawFileName := strings.TrimSpace(lines[0])
	directURL := strings.TrimSpace(lines[1])

	safeName := sanitizeFileName(rawFileName)
	if safeName == "" || safeName == "NA" {
		safeName = "video_" + time.Now().Format("20060102_150405")
	}

	if filepath.Ext(safeName) == "" {
		safeName += ".mp4"
	}

	return safeName, directURL, nil
}

func sanitizeFileName(name string) string {
	re := regexp.MustCompile(`[^\p{L}\p{N}._-]+`)
	safe := re.ReplaceAllString(name, "_")
	safe = strings.Trim(safe, "._")
	return safe
}
