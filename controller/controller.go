package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"myproject/service"
	"net/http"
)

type Controller struct {
	serv service.Service
}

func NewController(serv service.Service) *Controller {
	return &Controller{
		serv: serv,
	}
}

func (c *Controller) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Такого метода нет", http.StatusMethodNotAllowed)
		return
	}

	urlPath := r.URL.Path
	id := urlPath[5:]

	reader, err := c.serv.GetVideo(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")

	_, err = io.Copy(w, reader)
}

func (c *Controller) SaveByURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Такого меода нет", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
		return
	}

	urlStr := r.FormValue("url")
	if urlStr == "" {
		http.Error(w, "Параметр url не передан", http.StatusBadRequest)
		return
	}

	ind, err := c.serv.SaveByURL(r.Context(), urlStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, `{"id":"%s","get_url":"https://localhost:8080/get/%s"}`, ind, ind)
}

func (c *Controller) SaveByFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Такого меода нет", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Файл не найден", http.StatusBadRequest)
		return
	}
	defer file.Close()

	id, err := c.serv.SaveByFile(r.Context(), header.Filename, file, header.Size)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, `{"id":"%s","get_url":"https://localhost:8080/get/%s"}`, id, id)
}

func (c *Controller) ListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	videos, err := c.serv.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type responseItem struct {
		ID       string `json:"id"`
		FileName string `json:"file_name"`
		GetURL   string `json:"get_url"`
	}
	resp := make([]responseItem, 0, len(videos))
	for _, v := range videos {
		resp = append(resp, responseItem{
			ID:       v.ID,
			FileName: v.FileName,
			GetURL:   fmt.Sprintf("https://localhost:8080/get/%s", v.ID),
		})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}
