package main

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type Config struct {
	BucketName    string `json:"bucket_name"`
	PathDownload  string `json:"path_download"`
	GoogleProject string `json:"project_id"`
}

func main() {
	ctx := context.Background()

	// Load config
	cfgFile, err := os.ReadFile("appsetting.json")
	if err != nil {
		panic(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(cfgFile, &raw); err != nil {
		panic(err)
	}

	var cfg Config
	if err := json.Unmarshal(raw["AppSetting"], &cfg); err != nil {
		panic(err)
	}

	delete(raw, "AppSetting")
	credBytes, _ := json.Marshal(raw)

	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(credBytes))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	dateStr := time.Now().Format("20060102")
	downloadDir := filepath.Join(cfg.PathDownload, dateStr)
	os.MkdirAll(downloadDir, 0755)

	bkt := client.Bucket(cfg.BucketName)
	it := bkt.Objects(ctx, nil)

	var files []string
	for {
		attrs, err := it.Next()
		if err == storage.Done {
			break
		}
		if err != nil {
			panic(err)
		}

		if strings.HasSuffix(attrs.Name, ".csv") && strings.Contains(attrs.Name, dateStr) {
			files = append(files, attrs.Name)
		}
	}

	if len(files) == 0 {
		fmt.Println("No CSV files found.")
		return
	}

	p := mpb.New(mpb.WithWidth(60))

	for _, name := range files {
		obj := bkt.Object(name)
		attrs, _ := obj.Attrs(ctx)

		filePath := filepath.Join(downloadDir, filepath.Base(name))
		outFile, _ := os.Create(filePath)
		defer outFile.Close()

		r, err := obj.NewReader(ctx)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", name, err)
			continue
		}

		bar := p.New(attrs.Size,
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

		if err != nil {
			fmt.Printf("Failed to download %s: %v\n", name, err)
		} else {
			fmt.Printf("Downloaded %s successfully.\n", name)
		}
	}

	p.Wait()
	fmt.Println("Download completed.")
}
