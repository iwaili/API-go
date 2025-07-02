package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var allowedKeys []string

func main() {
	modelPath := "whisper.cpp/models/ggml-base.en.bin"

	// Load allowed API keys from JSON
	err := loadAllowedKeys("allowed_keys.json")
	if err != nil {
		log.Fatalf("Failed to load API keys: %v", err)
	}

	var wg sync.WaitGroup

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/transcribe", func(w http.ResponseWriter, r *http.Request) {
		handleTranscription(w, r, modelPath, &wg)
	})

	addr := ":8081"
	fmt.Printf("Server running at http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func loadAllowedKeys(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &allowedKeys)
}

func handleTranscription(w http.ResponseWriter, r *http.Request, modelPath string, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	if r.Method != http.MethodPost {
		http.Error(w, "Use POST with a file field named 'audio'", http.StatusMethodNotAllowed)
		return
	}

	apiKey := r.FormValue("apikey")

	// Binary Search to check if the key is in the sorted allowedKeys
	if !binarySearch(apiKey, allowedKeys) {
		http.Error(w, "Unauthorized: Invalid or missing API key", http.StatusUnauthorized)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "Audio file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save uploaded file
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	dir := "uploads"
	os.MkdirAll(dir, os.ModePerm)
	inPath := filepath.Join(dir, id+filepath.Ext(header.Filename))
	outFile, err := os.Create(inPath)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	io.Copy(outFile, file)
	outFile.Close()
	defer os.Remove(inPath)

	// Split audio into 5-second chunks using ffmpeg
	chunkDir := filepath.Join(dir, id+"_chunks")
	os.MkdirAll(chunkDir, os.ModePerm)
	defer os.RemoveAll(chunkDir)

	chunkPattern := filepath.Join(chunkDir, "chunk_%03d.wav")
	splitCmd := exec.Command("ffmpeg", "-i", inPath, "-f", "segment", "-segment_time", "5", "-c", "copy", chunkPattern)
	err = splitCmd.Run()
	if err != nil {
		http.Error(w, "Audio splitting failed", http.StatusInternalServerError)
		return
	}

	// Streaming response
	w.Header().Set("Content-Type", "application/x-ndjson")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Combine transcriptions from all chunks
	var fullTranscript []string

	files, _ := filepath.Glob(filepath.Join(chunkDir, "*.wav"))
	for _, chunk := range files {
		cmd := exec.Command("whisper.cpp/build/bin/whisper-cli", "-f", chunk, "--model", modelPath)
		raw, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(w, `{"error": "Transcription failed"}`+"\n")
		} else {
			transcript := extractText(string(raw))
			fullTranscript = append(fullTranscript, transcript)
			fmt.Fprintf(w, `{"transcription": %q}`+"\n", transcript)
		}
		flusher.Flush()
	}

	// Combine all parts into a single transcription
	finalTranscript := strings.Join(fullTranscript, " ")
	logTranscription(r.RemoteAddr, apiKey, header.Filename, finalTranscript)
}


func extractText(raw string) string {
	var lines []string
	re := regexp.MustCompile(`^\[\d{2}:\d{2}:\d{2}(?:\.\d+)?\s*-->\s*\d{2}:\d{2}:\d{2}(?:\.\d+)?\]\s*(.*)`)
	for _, line := range strings.Split(raw, "\n") {
		if m := re.FindStringSubmatch(line); m != nil {
			text := strings.TrimPrefix(m[1], "♪")
			lines = append(lines, strings.TrimSpace(text))
		}
	}
	return strings.Join(lines, " ")
}

func logTranscription(clientIP, apiKey, filename, transcript string) {
	logEntry := fmt.Sprintf("[%s] IP: %s | Key: %s | File: %s\nTranscription: %s\n\n",
		time.Now().Format(time.RFC3339),
		clientIP,
		apiKey,
		filename,
		transcript,
	)

	f, err := os.OpenFile("tra.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to write log:", err)
		return
	}
	defer f.Close()

	f.WriteString(logEntry)
}

func saveAllowedKeys(path string, allowedKeys []string) error {
	data, err := json.MarshalIndent(allowedKeys, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func binarySearch(key string, sortedKeys []string) bool {
	i := sort.SearchStrings(sortedKeys, key)
	return i < len(sortedKeys) && sortedKeys[i] == key
}
