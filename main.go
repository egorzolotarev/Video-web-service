package main

import (
	"log"
	"myproject/controller"
	"myproject/data"
	"myproject/service"
	"net/http"
	"os"
)

func main() {
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioUser := os.Getenv("MINIO_USER")
	minioPassword := os.Getenv("MINIO_PASSWORD")
	minioBucket := os.Getenv("MINIO_BUCKET")
	minIORepo, err := data.NewMinIORepo(minioEndpoint, minioUser, minioPassword, minioBucket)
	if err != nil {
		log.Fatalf("Ошибка инициализации MinIO: %v", err)
	}

	dsn := os.Getenv("DB_DSN")
	dbRepo, err := data.NewDBRepo(dsn)
	if err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	serv := service.NewService(*minIORepo, *dbRepo)
	cont := controller.NewController(*serv)

	// Используем свой мультиплексор (хотя можно использовать и встроенный)
	mult := http.NewServeMux()
	mult.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html")
	})
	mult.HandleFunc("/upload-by-url", cont.SaveByURLHandler)
	mult.HandleFunc("/upload-file", cont.SaveByFileHandler)
	mult.HandleFunc("/get/", cont.GetHandler)
	mult.HandleFunc("/list", cont.ListHandler)
	log.Fatal(http.ListenAndServeTLS(":8080", "/app/certs/public.crt", "/app/certs/private.key", mult))
}
