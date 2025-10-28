package main

import (
    "context"
    "encoding/base64"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
)

// SaveJSONToCloudant saves the provided JSON bytes as a new document in the given Cloudant DB.
// cloudantURL should be the base URL (e.g. https://account.cloudant.com). db is the database name.
// If username/password are provided, basic auth will be used.
func SaveJSONToCloudant(ctx context.Context, cloudantURL, db, username, password string, jsonBytes []byte) ([]byte, int, error) {
    // Build request URL
    u, err := url.Parse(strings.TrimRight(cloudantURL, "/") + "/" + db)
    if err != nil {
        return nil, 0, fmt.Errorf("invalid cloudant url: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(string(jsonBytes)))
    if err != nil {
        return nil, 0, fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    if username != "" {
        auth := username + ":" + password
        req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, 0, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
    }

    if resp.StatusCode >= 300 {
        return body, resp.StatusCode, fmt.Errorf("cloudant returned status %d", resp.StatusCode)
    }

    return body, resp.StatusCode, nil
}
