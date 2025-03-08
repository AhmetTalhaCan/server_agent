package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

type Config struct {
	ServerPort int `json:"serverPort"`
}

// Config dosyasını yükler
func loadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, fmt.Errorf("Yapılandırma dosyası açılamadı: %v", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return config, fmt.Errorf("Dosya okunamadı: %v", err)
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, fmt.Errorf("JSON hatası: %v", err)
	}

	return config, nil
}

// **Agent'ı çalıştıran fonksiyon**
func startAgent() {
	cmd := exec.Command("go", "run", "agent.go")
	cmd.Dir = "../agent" // Agent klasörüne gir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Agent başlatılamadı: %v", err)
	} else {
		log.Println("Agent başarıyla başlatıldı.")
	}
}

// **Sunucuyu başlatan fonksiyon**
func startServer() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Yapılandırma yüklenemedi: %v", err)
	}

	// **Agent'ı başlat**
	go startAgent()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Sunucu çalışıyor!")
	})

	log.Printf("Sunucu %d portunda çalışıyor...", config.ServerPort)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), nil)
	if err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}

func main() {
	startServer()
}
