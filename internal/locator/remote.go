package locator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const appNameCacheFile = ".gwxapkg/wxid-cache.json"

type remoteMiniProgramInfo struct {
	Nickname string `json:"nickname"`
}

type remoteMiniProgramQueryRequest struct {
	AppID string `json:"appid"`
}

type remoteMiniProgramQueryResponse struct {
	Code   int                   `json:"code"`
	Errors string                `json:"errors"`
	Data   remoteMiniProgramInfo `json:"data"`
}

func EnrichMiniProgramNames(programs []MiniProgramInfo) {
	cache := loadRemoteMiniProgramCache()
	changed := false

	for i := range programs {
		if cached, ok := cache[programs[i].AppID]; ok && cached.Nickname != "" {
			programs[i].AppName = cached.Nickname
			continue
		}

		info, err := queryRemoteMiniProgramInfo(programs[i].AppID)
		if err != nil || info.Nickname == "" {
			continue
		}

		programs[i].AppName = info.Nickname
		cache[programs[i].AppID] = info
		changed = true
	}

	if changed {
		saveRemoteMiniProgramCache(cache)
	}
}

func ResolveMiniProgramName(appID string) string {
	return resolveRemoteMiniProgramName(appID)
}

func queryRemoteMiniProgramInfo(appID string) (remoteMiniProgramInfo, error) {
	body, err := json.Marshal(remoteMiniProgramQueryRequest{AppID: appID})
	if err != nil {
		return remoteMiniProgramInfo{}, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://kainy.cn/api/weapp/info/", bytes.NewReader(body))
	if err != nil {
		return remoteMiniProgramInfo{}, err
	}

	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("User-Agent", "Gwxapkg/2.6.0")

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return remoteMiniProgramInfo{}, err
	}
	defer resp.Body.Close()

	var queryResp remoteMiniProgramQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return remoteMiniProgramInfo{}, err
	}

	if queryResp.Code != 0 {
		return remoteMiniProgramInfo{}, fmt.Errorf(queryResp.Errors)
	}

	return queryResp.Data, nil
}

func resolveRemoteMiniProgramName(appID string) string {
	if appID == "" {
		return ""
	}

	cache := loadRemoteMiniProgramCache()
	if cached, ok := cache[appID]; ok && cached.Nickname != "" {
		return cached.Nickname
	}

	info, err := queryRemoteMiniProgramInfo(appID)
	if err != nil || info.Nickname == "" {
		return ""
	}

	cache[appID] = info
	saveRemoteMiniProgramCache(cache)
	return info.Nickname
}

func loadRemoteMiniProgramCache() map[string]remoteMiniProgramInfo {
	path, err := remoteMiniProgramCachePath()
	if err != nil {
		return map[string]remoteMiniProgramInfo{}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return map[string]remoteMiniProgramInfo{}
	}

	cache := make(map[string]remoteMiniProgramInfo)
	if err := json.Unmarshal(content, &cache); err != nil {
		return map[string]remoteMiniProgramInfo{}
	}

	return cache
}

func saveRemoteMiniProgramCache(cache map[string]remoteMiniProgramInfo) {
	path, err := remoteMiniProgramCachePath()
	if err != nil {
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}

	content, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile(path, content, 0600)
}

func remoteMiniProgramCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, appNameCacheFile), nil
}
