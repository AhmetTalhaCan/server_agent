package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// ajan tarafından toplanan veriler
type AgentData struct {
	InstalledSoftware string `json:"installedSoftware"`
	Uptime            string `json:"uptime"`
	CPUInfo           string `json:"cpuInfo"`
	MemoryInfo        string `json:"memoryInfo"`
	DiskInfo          string `json:"diskInfo"`
	OSInfo            string `json:"osInfo"`
	SystemInfo        string `json:"systemInfo"`
	BootTime          string `json:"bootTime"`
}

// sistem bilgilerini alır
func getSystemInfo() string {
	hostInfo, err := host.Info()
	if err != nil {
		return fmt.Sprintf("Sistem bilgisi alınırken hata: %v", err)
	}

	cpuCores, err := cpu.Counts(false)
	if err != nil {
		return fmt.Sprintf("CPU bilgisi alınırken hata: %v", err)
	}

	return fmt.Sprintf("İşletim Sistemi: %s %s %s, Kernel: %s, CPU: %d çekirdek", hostInfo.Platform, hostInfo.PlatformVersion, hostInfo.PlatformFamily, hostInfo.KernelVersion, cpuCores)
}

// bellek bilgilerini alır
func getMemoryInfo() string {
	virtualMemory, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Sprintf("Bellek bilgisi alınırken hata: %v", err)
	}
	return fmt.Sprintf("Toplam Bellek: %.2f GB, Kullanılan Bellek: %.2f GB (%.2f%%)", float64(virtualMemory.Total)/1e9, float64(virtualMemory.Used)/1e9, virtualMemory.UsedPercent)
}

// CPU kullanım bilgilerini alır
func getCPUInfo() string {
	cpus, err := cpu.Percent(0, true)
	if err != nil {
		return fmt.Sprintf("CPU bilgisi alınırken hata: %v", err)
	}
	var cpuInfo []string
	for i, percent := range cpus {
		cpuInfo = append(cpuInfo, fmt.Sprintf("CPU %d: %.2f%%", i, percent))
	}
	return strings.Join(cpuInfo, "\n")
}

// disk bilgilerini alır
func getDiskInfo() string {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return fmt.Sprintf("Hata: %v", err)
	}

	var info []string
	for _, p := range partitions {
		d, err := disk.Usage(p.Mountpoint)
		if err != nil {
			return fmt.Sprintf("Hata: %v", err)
		}
		info = append(info, fmt.Sprintf("%s: Total: %v, Free: %v, UsedPercent: %f%%", p.Mountpoint, d.Total, d.Free, d.UsedPercent))
	}
	return strings.Join(info, "\n")
}

// işletim sistemi bilgilerini alır
func getOSInfo() string {
	info, err := host.Info()
	if err != nil {
		return fmt.Sprintf("Hata: %v", err)
	}
	return fmt.Sprintf("OS: %v, Platform: %v, PlatformFamily: %v, PlatformVersion: %v, KernelVersion: %v", info.OS, info.Platform, info.PlatformFamily, info.PlatformVersion, info.KernelVersion)
}

// yüklü yazılımları ve sürümlerini listeler
func listInstalledSoftwareWithVersion() string {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "Get-ItemProperty", "HKLM:\\Software\\Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\*", "|", "Select-Object", "DisplayName, DisplayVersion")
	} else if runtime.GOOS == "linux" {
		cmd = exec.Command("dpkg-query", "-W", "-f='${binary:Package} ${Version}\n'")
	} else {
		return "Unsupported OS"
	}

	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	return string(output)
}

// sistemin çalışma süresini alır
func getUptime() string {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "(get-date) - (gcim Win32_OperatingSystem).LastBootUpTime")
	} else if runtime.GOOS == "linux" {
		cmd = exec.Command("uptime", "-p")
	} else {
		return "Unsupported OS"
	}

	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	return strings.TrimSpace(string(output))
}

// sistemin açılış zamanını alır
func getBootTime() string {
	bootTime, err := host.BootTime()
	if err != nil {
		return fmt.Sprintf("Boot time bilgisi alınırken hata: %v", err)
	}
	t := time.Unix(int64(bootTime), 0)
	return t.Format("2006-01-02 15:04:05")
}

// ajan verilerini sunucuya gönderir
func sendDataToServer(agentData AgentData, serverURL string) error {
	jsonData, err := json.Marshal(agentData)
	if err != nil {
		return fmt.Errorf("JSON verisi oluşturulurken hata: %v", err)
	}

	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("sunucuya veri gönderilirken hata: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sunucudan beklenmeyen yanıt: %v", resp.Status)
	}

	return nil
}

// ajanı çalıştırır ve verileri toplar
func runAgent() {
	fmt.Println("Ajan çalışıyor...")

	installedSoftware := listInstalledSoftwareWithVersion()
	fmt.Println("Yüklü Yazılımlar ve Sürümleri:\n", installedSoftware)

	uptime := getUptime()
	fmt.Println("Sistem Açılış Süresi:\n", uptime)

	cpuInfo := getCPUInfo()
	fmt.Println("CPU Bilgisi:\n", cpuInfo)

	memoryInfo := getMemoryInfo()
	fmt.Println("Bellek Bilgisi:\n", memoryInfo)

	diskInfo := getDiskInfo()
	fmt.Println("Disk Bilgisi:\n", diskInfo)

	osInfo := getOSInfo()
	fmt.Println("İşletim Sistemi Bilgisi:\n", osInfo)

	systemInfo := getSystemInfo()
	fmt.Println("Sistem Bilgisi:\n", systemInfo)

	bootTime := getBootTime()
	fmt.Println("Açılış Süresi:\n", bootTime)

	agentData := AgentData{
		InstalledSoftware: installedSoftware,
		Uptime:            uptime,
		CPUInfo:           cpuInfo,
		MemoryInfo:        memoryInfo,
		DiskInfo:          diskInfo,
		OSInfo:            osInfo,
		SystemInfo:        systemInfo,
		BootTime:          bootTime,
	}

	serverURL := "http://localhost:8080/getAgentData"
	err := sendDataToServer(agentData, serverURL)
	if err != nil {
		fmt.Printf("Sunucuya veri gönderilemedi: %v\n", err)
		return
	}

	fmt.Println("Veriler başarıyla sunucuya gönderildi.")
}

// ajanı belirli aralıklarla çalıştırır
func main() {
	for {
		fmt.Println("Agent working... Time:", time.Now())
		runAgent()
		time.Sleep(1 * time.Minute) // 1 dakika bekle
	}
}
