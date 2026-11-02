package data

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DBRepo struct {
	db *sql.DB
}

func NewDBRepo(dsn string) (*DBRepo, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при открытии соединения с БД: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("Ошибка при подключении к БД: %v", err)
	}

	return &DBRepo{db: db}, nil
}

func (r *DBRepo) Put(ctx context.Context, fileName string) (string, error) {
	var id string
	query := `INSERT INTO videos (file_name) VALUES ($1) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, fileName).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("Ошибка при вставке в БД: %v", err)
	}

	return id, nil
}

func (r *DBRepo) GetFileNameByID(ctx context.Context, id string) (string, error) {
	var out string
	query := `SELECT file_name FROM videos WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&out)
	return out, err
}

type Video struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
}

func (r *DBRepo) GetAll(ctx context.Context) ([]Video, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, file_name FROM videos ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		var v Video
		if err := rows.Scan(&v.ID, &v.FileName); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

func (r *DBRepo) ExistsByName(ctx context.Context, fileName string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM videos WHERE file_name = $1)`
	err := r.db.QueryRowContext(ctx, query, fileName).Scan(&exists)
	return exists, err
}
