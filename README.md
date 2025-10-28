# Go REST API — Upload to IBM COS and save JSON to Cloudant

This small example shows a minimal Go REST API with two endpoints:

- POST /upload-file — multipart/form-data with field `file` and optional `objectName`. Uploads the file to IBM Cloud Object Storage (S3-compatible) using the minio v6 client.
- POST /save-json — application/json body forwarded to an IBM Cloudant database as a new document.

Environment variables

- IBM_COS_ENDPOINT — e.g. `s3.us-south.cloud-object-storage.appdomain.cloud` or with scheme `https://...`
- IBM_COS_ACCESS_KEY
- IBM_COS_SECRET_KEY
- IBM_COS_BUCKET
- IBM_COS_USE_SSL — optional, default `true`
- CLOUDANT_URL — e.g. `https://your-account.cloudant.com`
- CLOUDANT_DB — database name
- CLOUDANT_USERNAME — optional (for basic auth)
- CLOUDANT_PASSWORD — optional

Run locally

1. Fetch dependencies and build:

```powershell
Set-Location -Path 'D:\GitHub\go-server-rest-api'
go mod tidy
go build ./...
```

2. Start server:

```powershell
.\go-server-rest-api.exe
```

3. Examples

Upload file (curl):

```bash
curl -X POST -F "file=@/path/to/file.txt" -F "objectName=myfile.txt" http://localhost:8080/upload-file
```

Save JSON to Cloudant (curl):

```bash
curl -X POST -H "Content-Type: application/json" -d '{"foo":"bar"}' http://localhost:8080/save-json
```

Notes

- This example uses minio v6 (module `github.com/minio/minio-go`) to stay compatible with Go 1.20. If you run Go >= 1.23 you can update to minio v7 and revert the v6-specific code.
- For production use: add request validation, streaming for large files, better error handling, authentication, and TLS.
