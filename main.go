package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Config struct {
	BucketName    string `json:"bucket_name"`
	PathDownload  string `json:"path_download"`
	GoogleProject string `json:"project_id"`
}

func main() {
	ctx := context.Background()

	// Load appsetting.json
	cfgFile, err := os.ReadFile("appsetting.json")
	if err != nil {
		panic(fmt.Errorf("failed to read config: %w", err))
	}

	// Parse JSON with nested "AppSetting"
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(cfgFile, &raw); err != nil {
		panic(fmt.Errorf("failed to unmarshal json: %w", err))
	}

	var cfg Config
	if err := json.Unmarshal(raw["AppSetting"], &cfg); err != nil {
		panic(fmt.Errorf("failed to unmarshal AppSetting: %w", err))
	}

	// Remove "AppSetting" key, keep credentials
	delete(raw, "AppSetting")
	credBytes, _ := json.Marshal(raw)

	// Initialize GCS client
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(credBytes))
	if err != nil {
		panic(fmt.Errorf("failed to create GCS client: %w", err))
	}
	defer client.Close()

	// Set today's date and prepare download dir
	dateStr := time.Now().Format("20060102")
	downloadDir := filepath.Join(cfg.PathDownload, dateStr)
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		panic(fmt.Errorf("failed to create download directory: %w", err))
	}

	// List matching .csv files from bucket
	bkt := client.Bucket(cfg.BucketName)
	it := bkt.Objects(ctx, nil)

	var files []string
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(fmt.Errorf("failed to list bucket objects: %w", err))
		}
		if strings.HasSuffix(attrs.Name, ".csv") && strings.Contains(attrs.Name, dateStr) {
			files = append(files, attrs.Name)
		}
	}

	if len(files) == 0 {
		fmt.Println("No CSV files found for today:", dateStr)
		return
	}

	// Start progress container
	p := mpb.New(mpb.WithWidth(60))

	for _, name := range files {
		obj := bkt.Object(name)
		attrs, err := obj.Attrs(ctx)
		if err != nil {
			fmt.Printf("Failed to get attributes for %s: %v\n", name, err)
			continue
		}

		filePath := filepath.Join(downloadDir, filepath.Base(name))
		outFile, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Failed to create file %s: %v\n", filePath, err)
			continue
		}

		r, err := obj.NewReader(ctx)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", name, err)
			outFile.Close()
			continue
		}

		bar := p.New(
			attrs.Size,
			mpb.BarStyle().Rbound("|").Lbound("|").Tip(">").Filler("="),
			mpb.PrependDecorators(
				decor.Name(filepath.Base(name), decor.WC{W: len(filepath.Base(name)) + 1, C: decor.DSyncWidth}),
			),
			mpb.AppendDecorators(
				decor.CountersKibiByte("% .2f / % .2f"),
				decor.Percentage(decor.WC{W: 5}),
			),
		)

		proxyReader := bar.ProxyReader(r)
		_, err = io.Copy(outFile, proxyReader)
		_ = proxyReader.Close()
		_ = r.Close()
		_ = outFile.Close()

		if err != nil {
			fmt.Printf("Failed to download %s: %v\n", name, err)
		} else {
			fmt.Printf("Downloaded %s successfully.\n", name)
		}
	}

	p.Wait()
	fmt.Println("✅ Download completed.")
}
