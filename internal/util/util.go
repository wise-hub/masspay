package util

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logDir := "log"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logrus.Fatalf("Failed to create log directory: %s", err)
	}

	logPath := filepath.Join(logDir, time.Now().Format("20060102")+".log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Logger.Fatalf("Failed to log to file, using default stderr: %s", err)
	}
	Logger.SetOutput(file)
}

func CreateDirectory(basePath string, dirNames ...string) error {
	for _, dir := range dirNames {
		path := filepath.Join(basePath, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				Logger.Errorf("Failed to create directory %s: %v", dir, err)
				return fmt.Errorf("failed to create %s directory: %v", dir, err)
			}
			Logger.Infof("Created directory: %s", path)
		}
	}
	return nil
}

func ConvertDate(date string, format string) string {
	year := date[0:4]
	month := date[4:6]
	day := date[6:]
	var result string

	if format == "SHORT" {
		result = day + month + year[2:]
	} else if format == "LONG" {
		result = day + month + year
	} else {
		Logger.Warn("Invalid date format provided")
		return ""
	}

	Logger.Infof("Converted date from %s to %s format: %s", date, format, result)
	return result
}

func GenerateFileHash(data []byte) string {
	sum := 0
	for _, b := range data {
		sum += int(b)
	}
	return fmt.Sprintf("%08X", sum%0xFFFFFFFF)
}

func ValidateDate(date string) bool {
	parsedDate, err := time.Parse("20060102", date)
	if err != nil {
		Logger.WithError(err).Error("Invalid date format")
		return false
	}

	today := time.Now()
	startOfToday := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	return !parsedDate.Before(startOfToday)
}

func ValidateIBAN(iban string) bool {

	if iban == "" {
		return false
	}

	if len(iban) != 22 {
		return false
	}

	match, err := regexp.MatchString(`^BG\d{2}FINV9150.{10}$`, iban)
	if err != nil || !match {
		return false
	}

	iban = iban[4:] + iban[:4]

	transformed := ""
	for _, char := range iban {
		if unicode.IsLetter(char) {
			transformed += strconv.Itoa(int(char - 'A' + 10))
		} else if unicode.IsDigit(char) {
			transformed += string(char)
		} else {
			return false
		}
	}

	return mod97(transformed) == 1
}

func mod97(number string) int {
	remainder := 0
	for i := 0; i < len(number); i++ {
		remainder = remainder*10 + int(number[i]-'0')
		remainder %= 97
	}
	return remainder
}

func ValidateFilename(filename string) bool {
	match, _ := regexp.MatchString(`^O\d{6}p\.dp\d+$`, filename)
	return match
}
