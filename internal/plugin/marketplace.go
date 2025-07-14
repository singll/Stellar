package plugin

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Marketplace 插件市场
type Marketplace struct {
	config MarketplaceConfig
	client *http.Client
	cache  *MarketplaceCache
}

// MarketplaceConfig 市场配置
type MarketplaceConfig struct {
	// 官方仓库URL
	OfficialRepository string `json:"official_repository"`
	// 第三方仓库URL列表
	ThirdPartyRepositories []string `json:"third_party_repositories"`
	// 缓存目录
	CacheDir string `json:"cache_dir"`
	// 下载目录
	DownloadDir string `json:"download_dir"`
	// 安装目录
	InstallDir string `json:"install_dir"`
	// 缓存过期时间（小时）
	CacheExpireHours int `json:"cache_expire_hours"`
	// 启用安全验证
	EnableSignatureVerification bool `json:"enable_signature_verification"`
	// 可信任的签名密钥
	TrustedKeys []string `json:"trusted_keys"`
}

// MarketplaceCache 市场缓存
type MarketplaceCache struct {
	Repositories map[string]*RepositoryCache `json:"repositories"`
	LastUpdate   time.Time                   `json:"last_update"`
}

// RepositoryCache 仓库缓存
type RepositoryCache struct {
	URL         string                    `json:"url"`
	Plugins     map[string]*PluginPackage `json:"plugins"`
	LastSync    time.Time                 `json:"last_sync"`
	Fingerprint string                    `json:"fingerprint"`
}

// PluginPackage 插件包
type PluginPackage struct {
	// 基本信息
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	License     string    `json:"license"`
	Homepage    string    `json:"homepage"`
	Repository  string    `json:"repository"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 包信息
	DownloadURL string `json:"download_url"`
	FileSize    int64  `json:"file_size"`
	Checksum    string `json:"checksum"`
	Signature   string `json:"signature"`

	// 依赖信息
	Dependencies []string `json:"dependencies"`
	Conflicts    []string `json:"conflicts"`

	// 评分信息
	Rating      float64 `json:"rating"`
	Downloads   int64   `json:"downloads"`
	Reviews     int     `json:"reviews"`

	// 兼容性信息
	MinVersion    string   `json:"min_version"`
	MaxVersion    string   `json:"max_version"`
	Platforms     []string `json:"platforms"`
	Languages     []string `json:"languages"`

	// 安全信息
	SecurityRating  string   `json:"security_rating"`
	Vulnerabilities []string `json:"vulnerabilities"`
	LastScanned     time.Time `json:"last_scanned"`
}

// InstallStatus 安装状态
type InstallStatus struct {
	Status     string    `json:"status"` // installing, installed, failed, updating, uninstalling
	Progress   float64   `json:"progress"`
	Message    string    `json:"message"`
	StartTime  time.Time `json:"start_time"`
	FinishTime time.Time `json:"finish_time"`
	Error      string    `json:"error"`
}

// NewMarketplace 创建插件市场
func NewMarketplace(config MarketplaceConfig) *Marketplace {
	// 设置默认值
	if config.OfficialRepository == "" {
		config.OfficialRepository = "https://plugins.stellar.security"
	}
	if config.CacheDir == "" {
		config.CacheDir = "/tmp/stellar/plugin_cache"
	}
	if config.DownloadDir == "" {
		config.DownloadDir = "/tmp/stellar/plugin_downloads"
	}
	if config.InstallDir == "" {
		config.InstallDir = "/opt/stellar/plugins"
	}
	if config.CacheExpireHours <= 0 {
		config.CacheExpireHours = 24
	}

	// 创建必要的目录
	os.MkdirAll(config.CacheDir, 0755)
	os.MkdirAll(config.DownloadDir, 0755)
	os.MkdirAll(config.InstallDir, 0755)

	marketplace := &Marketplace{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: &MarketplaceCache{
			Repositories: make(map[string]*RepositoryCache),
			LastUpdate:   time.Time{},
		},
	}

	// 加载缓存
	marketplace.loadCache()

	return marketplace
}

// UpdateIndex 更新插件索引
func (m *Marketplace) UpdateIndex() error {
	// 更新官方仓库
	if err := m.updateRepository(m.config.OfficialRepository); err != nil {
		return fmt.Errorf("更新官方仓库失败: %v", err)
	}

	// 更新第三方仓库
	for _, repoURL := range m.config.ThirdPartyRepositories {
		if err := m.updateRepository(repoURL); err != nil {
			// 第三方仓库错误不中断流程，仅记录错误
			fmt.Printf("更新第三方仓库失败 [%s]: %v\n", repoURL, err)
		}
	}

	// 更新缓存时间
	m.cache.LastUpdate = time.Now()

	// 保存缓存
	return m.saveCache()
}

// updateRepository 更新单个仓库
func (m *Marketplace) updateRepository(repoURL string) error {
	// 获取仓库索引
	indexURL := repoURL + "/index.json"
	resp, err := m.client.Get(indexURL)
	if err != nil {
		return fmt.Errorf("获取仓库索引失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("仓库索引返回错误状态: %d", resp.StatusCode)
	}

	// 读取索引内容
	indexData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取仓库索引失败: %v", err)
	}

	// 计算指纹
	fingerprint := fmt.Sprintf("%x", md5.Sum(indexData))

	// 检查是否需要更新
	if repoCache, exists := m.cache.Repositories[repoURL]; exists {
		if repoCache.Fingerprint == fingerprint {
			// 内容未变化，更新同步时间即可
			repoCache.LastSync = time.Now()
			return nil
		}
	}

	// 解析索引
	var plugins map[string]*PluginPackage
	if err := json.Unmarshal(indexData, &plugins); err != nil {
		return fmt.Errorf("解析仓库索引失败: %v", err)
	}

	// 更新缓存
	m.cache.Repositories[repoURL] = &RepositoryCache{
		URL:         repoURL,
		Plugins:     plugins,
		LastSync:    time.Now(),
		Fingerprint: fingerprint,
	}

	return nil
}

// Search 搜索插件
func (m *Marketplace) Search(query string, category string, tags []string) ([]*PluginPackage, error) {
	var results []*PluginPackage

	// 遍历所有仓库
	for _, repoCache := range m.cache.Repositories {
		for _, plugin := range repoCache.Plugins {
			if m.matchSearchCriteria(plugin, query, category, tags) {
				results = append(results, plugin)
			}
		}
	}

	// 按评分和下载量排序
	sort.Slice(results, func(i, j int) bool {
		// 先按评分排序
		if results[i].Rating != results[j].Rating {
			return results[i].Rating > results[j].Rating
		}
		// 再按下载量排序
		return results[i].Downloads > results[j].Downloads
	})

	return results, nil
}

// matchSearchCriteria 匹配搜索条件
func (m *Marketplace) matchSearchCriteria(plugin *PluginPackage, query string, category string, tags []string) bool {
	// 查询字符串匹配
	if query != "" {
		queryLower := strings.ToLower(query)
		if !strings.Contains(strings.ToLower(plugin.Name), queryLower) &&
			!strings.Contains(strings.ToLower(plugin.Description), queryLower) &&
			!strings.Contains(strings.ToLower(plugin.Author), queryLower) {
			return false
		}
	}

	// 分类匹配
	if category != "" && plugin.Category != category {
		return false
	}

	// 标签匹配
	if len(tags) > 0 {
		for _, requiredTag := range tags {
			found := false
			for _, pluginTag := range plugin.Tags {
				if strings.EqualFold(pluginTag, requiredTag) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// GetPlugin 获取插件详情
func (m *Marketplace) GetPlugin(id string) (*PluginPackage, error) {
	for _, repoCache := range m.cache.Repositories {
		if plugin, exists := repoCache.Plugins[id]; exists {
			return plugin, nil
		}
	}
	return nil, fmt.Errorf("插件不存在: %s", id)
}

// Download 下载插件
func (m *Marketplace) Download(plugin *PluginPackage) (string, error) {
	// 创建下载路径
	filename := fmt.Sprintf("%s_%s.zip", plugin.ID, plugin.Version)
	downloadPath := filepath.Join(m.config.DownloadDir, filename)

	// 检查文件是否已存在
	if _, err := os.Stat(downloadPath); err == nil {
		// 验证文件完整性
		if m.verifyFile(downloadPath, plugin.Checksum) {
			return downloadPath, nil
		}
		// 文件损坏，删除重新下载
		os.Remove(downloadPath)
	}

	// 下载文件
	resp, err := m.client.Get(plugin.DownloadURL)
	if err != nil {
		return "", fmt.Errorf("下载插件失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载插件返回错误状态: %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(downloadPath)
	if err != nil {
		return "", fmt.Errorf("创建下载文件失败: %v", err)
	}
	defer file.Close()

	// 复制数据
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("写入下载文件失败: %v", err)
	}

	// 验证文件完整性
	if !m.verifyFile(downloadPath, plugin.Checksum) {
		os.Remove(downloadPath)
		return "", fmt.Errorf("下载文件校验失败")
	}

	// 验证签名（如果启用）
	if m.config.EnableSignatureVerification {
		if err := m.verifySignature(downloadPath, plugin.Signature); err != nil {
			os.Remove(downloadPath)
			return "", fmt.Errorf("插件签名验证失败: %v", err)
		}
	}

	return downloadPath, nil
}

// verifyFile 验证文件校验和
func (m *Marketplace) verifyFile(filePath string, expectedChecksum string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false
	}

	actualChecksum := fmt.Sprintf("%x", hash.Sum(nil))
	return actualChecksum == expectedChecksum
}

// verifySignature 验证插件签名
func (m *Marketplace) verifySignature(filePath string, signature string) error {
	// TODO: 实现数字签名验证
	// 这里需要实现实际的数字签名验证逻辑
	// 可以使用 crypto/rsa 或其他加密库
	return nil
}

// Install 安装插件
func (m *Marketplace) Install(pluginID string) (*InstallStatus, error) {
	// 获取插件信息
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	// 创建安装状态
	status := &InstallStatus{
		Status:    "installing",
		Progress:  0.0,
		StartTime: time.Now(),
	}

	// 检查依赖
	status.Progress = 10.0
	status.Message = "检查依赖"
	if err := m.checkDependencies(plugin); err != nil {
		status.Status = "failed"
		status.Error = err.Error()
		return status, err
	}

	// 下载插件
	status.Progress = 30.0
	status.Message = "下载插件"
	downloadPath, err := m.Download(plugin)
	if err != nil {
		status.Status = "failed"
		status.Error = err.Error()
		return status, err
	}

	// 解压插件
	status.Progress = 60.0
	status.Message = "解压插件"
	installPath := filepath.Join(m.config.InstallDir, plugin.ID)
	if err := m.extractPlugin(downloadPath, installPath); err != nil {
		status.Status = "failed"
		status.Error = err.Error()
		return status, err
	}

	// 安装完成
	status.Progress = 100.0
	status.Status = "installed"
	status.Message = "安装完成"
	status.FinishTime = time.Now()

	return status, nil
}

// checkDependencies 检查插件依赖
func (m *Marketplace) checkDependencies(plugin *PluginPackage) error {
	for _, depID := range plugin.Dependencies {
		if _, err := m.GetPlugin(depID); err != nil {
			return fmt.Errorf("缺少依赖插件: %s", depID)
		}
	}
	return nil
}

// extractPlugin 解压插件
func (m *Marketplace) extractPlugin(downloadPath, installPath string) error {
	// TODO: 实现插件解压逻辑
	// 可以使用 archive/zip 包来解压插件文件
	os.MkdirAll(installPath, 0755)
	return nil
}

// GetCategories 获取插件分类
func (m *Marketplace) GetCategories() []string {
	categories := make(map[string]bool)
	
	for _, repoCache := range m.cache.Repositories {
		for _, plugin := range repoCache.Plugins {
			if plugin.Category != "" {
				categories[plugin.Category] = true
			}
		}
	}

	var result []string
	for category := range categories {
		result = append(result, category)
	}
	sort.Strings(result)
	
	return result
}

// GetTags 获取所有标签
func (m *Marketplace) GetTags() []string {
	tags := make(map[string]bool)
	
	for _, repoCache := range m.cache.Repositories {
		for _, plugin := range repoCache.Plugins {
			for _, tag := range plugin.Tags {
				tags[tag] = true
			}
		}
	}

	var result []string
	for tag := range tags {
		result = append(result, tag)
	}
	sort.Strings(result)
	
	return result
}

// loadCache 加载缓存
func (m *Marketplace) loadCache() error {
	cachePath := filepath.Join(m.config.CacheDir, "marketplace.json")
	
	data, err := os.ReadFile(cachePath)
	if err != nil {
		// 缓存文件不存在，使用默认值
		return nil
	}

	return json.Unmarshal(data, m.cache)
}

// saveCache 保存缓存
func (m *Marketplace) saveCache() error {
	cachePath := filepath.Join(m.config.CacheDir, "marketplace.json")
	
	data, err := json.MarshalIndent(m.cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

// IsCacheExpired 检查缓存是否过期
func (m *Marketplace) IsCacheExpired() bool {
	expireTime := time.Duration(m.config.CacheExpireHours) * time.Hour
	return time.Since(m.cache.LastUpdate) > expireTime
}

// GetStats 获取市场统计信息
func (m *Marketplace) GetStats() map[string]interface{} {
	totalPlugins := 0
	totalDownloads := int64(0)
	categoriesCount := make(map[string]int)

	for _, repoCache := range m.cache.Repositories {
		for _, plugin := range repoCache.Plugins {
			totalPlugins++
			totalDownloads += plugin.Downloads
			categoriesCount[plugin.Category]++
		}
	}

	return map[string]interface{}{
		"total_plugins":    totalPlugins,
		"total_downloads":  totalDownloads,
		"categories_count": categoriesCount,
		"repositories":     len(m.cache.Repositories),
		"last_update":      m.cache.LastUpdate,
	}
}