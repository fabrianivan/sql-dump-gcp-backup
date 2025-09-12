package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// read entire appsetting.json
	fullJson, err := ioutil.ReadFile("appsetting.json")
	if err != nil {
		fmt.Println("failed to read appsetting.json:", err)
		writeLog("failed to read appsetting.json: " + err.Error())
		return
	}

	// parse top-level JSON into a map so we can remove AppSetting
	var top map[string]interface{}
	if err := json.Unmarshal(fullJson, &top); err != nil {
		fmt.Println("failed to parse appsetting.json:", err)
		writeLog("failed to parse appsetting.json: " + err.Error())
		return
	}

	// extract AppSetting
	appSetting := map[string]interface{}{}
	if as, ok := top["AppSetting"]; ok {
		if am, ok2 := as.(map[string]interface{}); ok2 {
			appSetting = am
		}
	}

	bucketName := getString(appSetting, "bucket_name")
	pathDownload := getString(appSetting, "path_download")

	if bucketName == "" || pathDownload == "" {
		msg := "bucket_name or path_download missing in AppSetting"
		fmt.Println(msg)
		writeLog(msg)
		return
	}

	dateString := time.Now().Format("20060102") // yyyyMMdd
	datedPath := filepath.Join(pathDownload, dateString)
	if err := os.MkdirAll(datedPath, 0o755); err != nil {
		fmt.Println("failed to create download directory:", err)
		writeLog("failed to create download directory: " + err.Error())
		return
	}

	// remove AppSetting from top to produce service account JSON
	delete(top, "AppSetting")
	credsJSON, err := json.Marshal(top)
	if err != nil {
		fmt.Println("failed to marshal credential JSON:", err)
		writeLog("failed to marshal credential JSON: " + err.Error())
		return
	}

	// create storage client with credentials from JSON bytes
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(credsJSON))
	if err != nil {
		fmt.Println("failed to create storage client:", err)
		writeLog("failed to create storage client: " + err.Error())
		return
	}
	defer client.Close()

	// list objects in bucket (empty prefix)
	it := client.Bucket(bucketName).Objects(ctx, &storage.Query{Prefix: ""})
	countFile := 0

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println("listing objects failed:", err)
			writeLog("listing objects failed: " + err.Error())
			return
		}

		name := attrs.Name
		if strings.HasSuffix(strings.ToLower(name), ".csv") && strings.Contains(name, dateString) {
			fmt.Println("Found:", name)

			// build local file path similar to C# Replace("sql-dump/", "").Replace(":", "")
			cleanName := strings.ReplaceAll(name, "sql-dump/", "")
			cleanName = strings.ReplaceAll(cleanName, ":", "")
			filePath := filepath.Join(datedPath, cleanName)

			if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
				writeLog("failed to create directory for file "+filePath+": "+err.Error())
				continue
			}

			// read object
			rc, err := client.Bucket(bucketName).Object(name).NewReader(ctx)
			if err != nil {
				writeLog("failed to create reader for object "+name+": "+err.Error())
				continue
			}

			outFile, err := os.Create(filePath)
			if err != nil {
				writeLog("failed to create local file "+filePath+": "+err.Error())
				rc.Close()
				continue
			}

			_, err = io.Copy(outFile, rc)
			// close both
			outFile.Close()
			rc.Close()

			if err != nil {
				writeLog("failed to download object "+name+": "+err.Error())
				// optionally remove partial file
				_ = os.Remove(filePath)
				continue
			}

			writeLog("Success Get File " + name)
			countFile++
		}
	}

	if countFile == 0 {
		writeLog("Not found file")
	}
}

// writeLog writes/appends a log entry to LogFile/log_yyyyMMdd.log next to the executable
func writeLog(message string) {
	exePath, err := os.Executable()
	curpath := "."
	if err == nil {
		curpath = filepath.Dir(exePath)
	}

	logDir := filepath.Join(curpath, "LogFile")
	_ = os.MkdirAll(logDir, 0o755)

	logPath := filepath.Join(logDir, "log_"+time.Now().Format("20060102")+".log")
	entry := time.Now().Format("02-01-2006 15:04:05") + " : " + message + "\n"

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		// if writing to log fails, fallback to stdout
		fmt.Println("LOG ERROR:", err, "original message:", message)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		fmt.Println("LOG WRITE ERROR:", err)
	}
}
