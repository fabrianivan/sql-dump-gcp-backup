# GCS CSV Downloader (Go)

A simple Go utility to **download CSV files from Google Cloud Storage (GCS)**, with:
- Progress bars
- Date-based filtering
- Configurable destination folder
- Automatic folder creation
- Config and credential handling in one `appsetting.json`

## 📦 Features

- ✅ Downloads all `.csv` files from GCS bucket matching today’s date (`yyyyMMdd`)
- ✅ Displays per-file progress bars using [`mpb`](https://github.com/vbauerster/mpb)
- ✅ Uses GCP service account credentials
- ✅ Automatic retry, directory creation, and error logging


## 🛠 Requirements

- Go 1.18 or higher
- GCS Bucket with `.csv` files
- A service account JSON key
- Internet connection


## ⚙️ Configuration

Create a file called `appsetting.json`:

```json
{
  "AppSetting": {
    "bucket_name": "your-gcs-bucket-name",
    "path_download": "C:/FileFscm",
    "project_id": "your-gcp-project-id"
  },
  "type": "service_account",
  "project_id": "your-gcp-project-id",
  "private_key_id": "xxxxxx",
  "private_key": "-----BEGIN PRIVATE KEY-----\\n...\\n-----END PRIVATE KEY-----\\n",
  "client_email": "your-service-account@your-project.iam.gserviceaccount.com",
  "client_id": "123456789012345678901",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/your-service-account"
}
```

Creates a local folder named after today's date (e.g. C:/FileFscm/20250923)

Lists all .csv files in the GCS bucket that contain today’s date in the name

Downloads them one by one with a progress bar

🚀 Usage
Run directly
```
go run main.go
```

Or build and run
```
go build -o gcs-downloader
./gcs-downloader
```

⏳ Example Output
```
[===============================>-------------------------]  31.00 MiB / 60.00 MiB  51%
Downloading: sales_20250923.csv → C:/FileFscm/20250923/sales_20250923.csv
Downloaded sales_20250923.csv successfully.
```
### 🧩 Dependencies
Google Cloud Storage Client
mpb – terminal progress bar
Google Auth Libraries

Install with:
```
go get cloud.google.com/go/storage
go get github.com/vbauerster/mpb/v7
go get golang.org/x/oauth2/google
go get google.golang.org/api/option
```

### 📝 License
This project is licensed under the GNU General Public License v3.0.

### 🙋 FAQ
Q: What happens if there are no matching files?
A: The app exits cleanly with a "No CSV files found" message.

Q: Can I customize the file filter?
A: Yes, just modify the ListObjects() filter logic in main.go.

### 📂 Project Structure
```
├── main.go
├── appsetting.json
└── README.md
```

### 📣 Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you'd like to add or improve.