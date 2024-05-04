package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"masspay/internal/util"

	"golang.org/x/text/unicode/norm"
)

type jsonResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	FileURL string `json:"file_url,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, response jsonResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func FileUploadHandler(w http.ResponseWriter, r *http.Request) {
	util.Logger.Info("Starting file upload")

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		util.Logger.WithError(err).Error("Failed to parse multipart form")
		respondJSON(w, http.StatusBadRequest, jsonResponse{Success: false, Msg: "Error parsing multipart form"})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		util.Logger.WithError(err).Error("Failed to retrieve file")
		respondJSON(w, http.StatusBadRequest, jsonResponse{Success: false, Msg: "Invalid file"})
		return
	}
	defer file.Close()

	executionDate := r.FormValue("executionDate")
	if !util.ValidateDate(executionDate) {
		respondJSON(w, http.StatusBadRequest, jsonResponse{Success: false, Msg: "The provided date is invalid or in the past"})
		return
	}

	iban := r.FormValue("iban")
	if !util.ValidateIBAN(iban) {
		respondJSON(w, http.StatusBadRequest, jsonResponse{Success: false, Msg: "Invalid IBAN format"})
		return
	}

	if !util.ValidateFilename(fileHeader.Filename) {
		respondJSON(w, http.StatusBadRequest, jsonResponse{Success: false, Msg: "Invalid file name format"})
		return
	}

	companyName := r.FormValue("companyName")
	timeStamp := time.Now().Format("20060102-150405")
	baseDir := filepath.Join("temp", timeStamp)
	if err := util.CreateDirectory(baseDir, "input", "output"); err != nil {
		util.Logger.WithError(err).Error("Failed to create directories")
		respondJSON(w, http.StatusInternalServerError, jsonResponse{Success: false, Msg: "Failed to create directories"})
		return
	}

	inputFilePath := filepath.Join(baseDir, "input", fileHeader.Filename)
	inputFile, err := os.Create(inputFilePath)
	if err != nil {
		util.Logger.WithError(err).Error("Failed to save input file")
		respondJSON(w, http.StatusInternalServerError, jsonResponse{Success: false, Msg: "Failed to save input file"})
		return
	}
	defer inputFile.Close()

	if _, err := io.Copy(inputFile, file); err != nil {
		util.Logger.WithError(err).Error("Failed to write input file")
		respondJSON(w, http.StatusInternalServerError, jsonResponse{Success: false, Msg: "Failed to write input file"})
		return
	}

	outputFileForReading, err := os.Open(inputFilePath)
	if err != nil {
		util.Logger.WithError(err).Error("Failed to open input file for reading")
		http.Error(w, "Failed to open input file for reading", http.StatusInternalServerError)
		return
	}
	defer outputFileForReading.Close()

	outputFilePath := filepath.Join(baseDir, "output", fileHeader.Filename)
	if err := processAndRespond(w, r, outputFileForReading, outputFilePath, util.ConvertDate(executionDate, "LONG"), iban, companyName); err != nil {
		util.Logger.WithError(err).Error("Failed to process file")
		respondJSON(w, http.StatusInternalServerError, jsonResponse{Success: false, Msg: "Failed to process file"})
		return
	}

	respondJSON(w, http.StatusOK, jsonResponse{Success: true, Msg: "File processed successfully", FileURL: outputFilePath})
}
func processAndRespond(w http.ResponseWriter, r *http.Request, file io.Reader, fullPath, date, iban, companyName string) error {
	lineCount, totalAmount, hashValue, err := processFile(file)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)

	header := fmt.Sprintf("OMPDP;%s;%s;;%s;%.2f;%d;%s;\n",
		date, iban, strings.ToUpper(companyName), totalAmount, lineCount, hashValue)

	if _, err := writer.WriteString(header); err != nil {
		return err
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
		if _, err = io.Copy(writer, file); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(fullPath)))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, fullPath)
	return nil
}

func processFile(file io.Reader) (int, float64, string, error) {
	scanner := bufio.NewScanner(file)
	var lineCount int
	var totalAmount float64
	var fileBuffer bytes.Buffer

	hasher := fnv.New32a()

	for scanner.Scan() {
		line := norm.NFC.String(scanner.Text())
		actualLength := utf8.RuneCountInString(line)
		util.Logger.Infof("Processed line length: %d", actualLength)

		if actualLength > 108 {
			util.Logger.Warnf("Skipped line due to incorrect length: %d", actualLength)
			continue
		}

		amountStr := strings.TrimSpace(line[23:36])
		amount, err := strconv.ParseFloat(strings.ReplaceAll(amountStr, " ", ""), 64)
		if err != nil {
			util.Logger.Errorf("Failed to parse amount from line: %v", err)
			return 0, 0, "", fmt.Errorf("failed to parse amount: %v", err)
		}
		totalAmount += amount
		lineCount++

		fileBuffer.WriteString(line)
	}

	if err := scanner.Err(); err != nil {
		util.Logger.Errorf("Failed while reading the file: %v", err)
		return 0, 0, "", fmt.Errorf("failed while reading the file: %v", err)
	}

	if _, err := hasher.Write(fileBuffer.Bytes()); err != nil {
		util.Logger.Errorf("Failed to compute hash: %v", err)
		return 0, 0, "", fmt.Errorf("failed to compute hash: %v", err)
	}
	hashValue := fmt.Sprintf("%08X", hasher.Sum32())

	util.Logger.Infof("Total lines processed: %d, Total amount: %.2f, Hash: %s", lineCount, totalAmount, hashValue)
	return lineCount, totalAmount, hashValue, nil
}
