package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Config struct {
	ServerPort int `json:"serverPort"`
}

type AgentData struct {
	InstalledSoftware string `json:"installedSoftware"`
	Uptime            string `json:"uptime"`
	CPUInfo           string `json:"cpuInfo"`
	MemoryInfo        string `json:"memoryInfo"`
	DiskInfo          string `json:"diskInfo"`
	OSInfo            string `json:"osInfo"`
	SystemInfo        string `json:"systemInfo"`
}

// Config dosyasını yükler
func loadConfig(filename string) (Config, error) {
	var config Config
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return config, fmt.Errorf("yapılandırma dosyası bulunamadı: %s", filename)
	}

	configFile, err := os.Open(filename)
	if err != nil {
		return config, fmt.Errorf("yapılandırma dosyası açılamadı: %v", err)
	}
	defer configFile.Close()

	bytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		return config, fmt.Errorf("yapılandırma dosyası okunamadı: %v", err)
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, fmt.Errorf("yapılandırma dosyası çözülemedi: %v", err)
	}

	return config, nil
}

// Ana sayfa endpoint'i
func handleHome(w http.ResponseWriter, r *http.Request) {
	html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Sunucu Mesajı</title>
        </head>
        <body>
            <h1>Sunucuya Mesaj Gönder</h1>
            <form action="/write" method="post">
                <label for="message">Mesaj:</label>
                <input type="text" id="message" name="message">
                <input type="submit" value="Gönder">
            </form>
        </body>
        </html>
    `
	fmt.Fprint(w, html)
}

// Mesaj yazma endpoint'i
func handleWrite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Yalnızca POST istekleri kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		http.Error(w, "Mesaj boş olamaz", http.StatusBadRequest)
		return
	}

	fmt.Printf("Sunucuya yazılan mesaj: %s\n", message)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Mesaj alındı: %s", message)
}

// Ajan verisi alma endpoint'i
func handleGetAgentData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Yalnızca POST istekleri kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	var agentData AgentData
	err := json.NewDecoder(r.Body).Decode(&agentData)
	if err != nil {
		http.Error(w, "Geçersiz veri", http.StatusBadRequest)
		return
	}

	if agentData.InstalledSoftware == "" || agentData.Uptime == "" || agentData.CPUInfo == "" || agentData.MemoryInfo == "" || agentData.DiskInfo == "" || agentData.OSInfo == "" || agentData.SystemInfo == "" {
		http.Error(w, "Eksik ajan verisi", http.StatusBadRequest)
		return
	}

	fmt.Printf("Ajan verisi alındı: %+v\n", agentData)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ajan verisi alındı")
}

// Middleware: Recovery
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panik durumu: %v", err)
				http.Error(w, "Sunucu hatası", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func startServer() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Yapılandırma yüklenemedi: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Sunucu panik durumu: %v", r)
		}
	}()

	http.Handle("/", recoveryMiddleware(http.HandlerFunc(handleHome)))
	http.Handle("/write", recoveryMiddleware(http.HandlerFunc(handleWrite)))
	http.Handle("/getAgentData", recoveryMiddleware(http.HandlerFunc(handleGetAgentData)))

	log.Printf("Sunucu %d portunda çalışıyor...", config.ServerPort)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), nil)
	if err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}

func main() {
	startServer()
}
