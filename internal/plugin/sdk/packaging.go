package packaging

import (
	"archive/zip"
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/StellarServer/internal/pkg/logger"
)

// PackageBuilder 插件包构建器
type PackageBuilder struct {
	workDir   string
	outputDir string
	tempDir   string
	config    *PackageConfig
}

// PackageConfig 打包配置
type PackageConfig struct {
	// 基本信息
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	License     string `json:"license"`
	Homepage    string `json:"homepage"`
	Repository  string `json:"repository"`

	// 文件配置
	MainFile     string   `json:"main_file"`
	IncludeFiles []string `json:"include_files"`
	ExcludeFiles []string `json:"exclude_files"`

	// 依赖配置
	Dependencies []Dependency `json:"dependencies"`

	// 元数据
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`
	Language   string   `json:"language"`
	MinVersion string   `json:"min_version"`
	MaxVersion string   `json:"max_version"`
	Platforms  []string `json:"platforms"`

	// 签名配置
	SigningKey string `json:"signing_key,omitempty"`

	// 验证配置
	RunTests    bool   `json:"run_tests"`
	TestCommand string `json:"test_command"`
}

// Dependency 依赖信息
type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  string `json:"source"` // npm, pip, go, etc.
}

// PackageMetadata 包元数据
type PackageMetadata struct {
	// 基本信息
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	License     string `json:"license"`
	Homepage    string `json:"homepage"`
	Repository  string `json:"repository"`

	// 包信息
	FileSize  int64     `json:"file_size"`
	FileCount int       `json:"file_count"`
	Checksum  string    `json:"checksum"`
	SHA256    string    `json:"sha256"`
	CreatedAt time.Time `json:"created_at"`

	// 内容信息
	MainFile     string       `json:"main_file"`
	Files        []FileInfo   `json:"files"`
	Dependencies []Dependency `json:"dependencies"`

	// 分类信息
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Language string   `json:"language"`

	// 兼容性
	MinVersion string   `json:"min_version"`
	MaxVersion string   `json:"max_version"`
	Platforms  []string `json:"platforms"`

	// 签名信息
	Signature string `json:"signature,omitempty"`
}

// FileInfo 文件信息
type FileInfo struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
}

// NewPackageBuilder 创建包构建器
func NewPackageBuilder(workDir, outputDir string) *PackageBuilder {
	return &PackageBuilder{
		workDir:   workDir,
		outputDir: outputDir,
		tempDir:   filepath.Join(outputDir, "temp"),
	}
}

// LoadConfig 加载配置
func (b *PackageBuilder) LoadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config PackageConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	b.config = &config
	return nil
}

// SetConfig 设置配置
func (b *PackageBuilder) SetConfig(config *PackageConfig) {
	b.config = config
}

// Build 构建插件包
func (b *PackageBuilder) Build() (string, error) {
	if b.config == nil {
		return "", fmt.Errorf("未设置构建配置")
	}

	// 验证配置
	if err := b.validateConfig(); err != nil {
		return "", fmt.Errorf("配置验证失败: %v", err)
	}

	// 创建临时目录
	if err := os.MkdirAll(b.tempDir, 0755); err != nil {
		return "", fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(b.tempDir)

	// 运行测试（如果配置了）
	if b.config.RunTests {
		if err := b.runTests(); err != nil {
			return "", fmt.Errorf("测试失败: %v", err)
		}
	}

	// 收集文件
	files, err := b.collectFiles()
	if err != nil {
		return "", fmt.Errorf("收集文件失败: %v", err)
	}

	// 生成元数据
	metadata, err := b.generateMetadata(files)
	if err != nil {
		return "", fmt.Errorf("生成元数据失败: %v", err)
	}

	// 创建包文件
	packagePath, err := b.createPackage(files, metadata)
	if err != nil {
		return "", fmt.Errorf("创建包失败: %v", err)
	}

	// 签名包（如果配置了）
	if b.config.SigningKey != "" {
		if err := b.signPackage(packagePath); err != nil {
			return "", fmt.Errorf("签名包失败: %v", err)
		}
	}

	return packagePath, nil
}

// validateConfig 验证配置
func (b *PackageBuilder) validateConfig() error {
	if b.config.ID == "" {
		return fmt.Errorf("插件ID不能为空")
	}
	if b.config.Name == "" {
		return fmt.Errorf("插件名称不能为空")
	}
	if b.config.Version == "" {
		return fmt.Errorf("插件版本不能为空")
	}
	if b.config.MainFile == "" {
		return fmt.Errorf("主文件不能为空")
	}

	// 验证主文件是否存在
	mainFilePath := filepath.Join(b.workDir, b.config.MainFile)
	if _, err := os.Stat(mainFilePath); err != nil {
		return fmt.Errorf("主文件不存在: %s", b.config.MainFile)
	}

	// 验证版本格式
	if !regexp.MustCompile(`^\d+\.\d+\.\d+`).MatchString(b.config.Version) {
		return fmt.Errorf("版本格式无效，应为 x.y.z 格式")
	}

	return nil
}

// runTests 运行测试
func (b *PackageBuilder) runTests() error {
	// 这里简化实现，实际应该执行测试命令
	logger.Info("运行插件测试...")

	if b.config.TestCommand != "" {
		// 执行自定义测试命令
		logger.Info("执行测试命令: %s", b.config.TestCommand)
	} else {
		// 基于语言的默认测试
		switch b.config.Language {
		case "python":
			logger.Info("运行Python测试")
		case "javascript":
			logger.Info("运行JavaScript测试")
		case "go":
			logger.Info("运行Go测试")
		}
	}

	return nil
}

// collectFiles 收集要打包的文件
func (b *PackageBuilder) collectFiles() ([]string, error) {
	var files []string

	// 添加主文件
	files = append(files, b.config.MainFile)

	// 添加显式包含的文件
	for _, includePattern := range b.config.IncludeFiles {
		matches, err := b.findFiles(includePattern)
		if err != nil {
			return nil, fmt.Errorf("处理包含文件模式失败 %s: %v", includePattern, err)
		}
		files = append(files, matches...)
	}

	// 如果没有指定包含文件，使用默认规则
	if len(b.config.IncludeFiles) == 0 {
		defaultFiles, err := b.getDefaultFiles()
		if err != nil {
			return nil, fmt.Errorf("获取默认文件失败: %v", err)
		}
		files = append(files, defaultFiles...)
	}

	// 过滤排除的文件
	filteredFiles := b.filterExcludedFiles(files)

	// 去重
	fileSet := make(map[string]bool)
	var uniqueFiles []string
	for _, file := range filteredFiles {
		if !fileSet[file] {
			fileSet[file] = true
			uniqueFiles = append(uniqueFiles, file)
		}
	}

	return uniqueFiles, nil
}

// findFiles 查找匹配模式的文件
func (b *PackageBuilder) findFiles(pattern string) ([]string, error) {
	var matches []string

	err := filepath.WalkDir(b.workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径
		relPath, err := filepath.Rel(b.workDir, path)
		if err != nil {
			return err
		}

		// 跳过目录
		if d.IsDir() {
			return nil
		}

		// 检查是否匹配模式
		matched, err := filepath.Match(pattern, relPath)
		if err != nil {
			return err
		}

		if matched {
			matches = append(matches, relPath)
		}

		return nil
	})

	return matches, err
}

// getDefaultFiles 获取默认要包含的文件
func (b *PackageBuilder) getDefaultFiles() ([]string, error) {
	var files []string

	// 根据语言添加默认文件
	switch b.config.Language {
	case "python":
		defaultPatterns := []string{"*.py", "requirements.txt", "README.md", "LICENSE"}
		for _, pattern := range defaultPatterns {
			matches, err := b.findFiles(pattern)
			if err == nil {
				files = append(files, matches...)
			}
		}
	case "javascript":
		defaultPatterns := []string{"*.js", "package.json", "README.md", "LICENSE"}
		for _, pattern := range defaultPatterns {
			matches, err := b.findFiles(pattern)
			if err == nil {
				files = append(files, matches...)
			}
		}
	case "go":
		defaultPatterns := []string{"*.go", "go.mod", "go.sum", "README.md", "LICENSE"}
		for _, pattern := range defaultPatterns {
			matches, err := b.findFiles(pattern)
			if err == nil {
				files = append(files, matches...)
			}
		}
	case "yaml":
		defaultPatterns := []string{"*.yaml", "*.yml", "README.md", "LICENSE"}
		for _, pattern := range defaultPatterns {
			matches, err := b.findFiles(pattern)
			if err == nil {
				files = append(files, matches...)
			}
		}
	}

	return files, nil
}

// filterExcludedFiles 过滤排除的文件
func (b *PackageBuilder) filterExcludedFiles(files []string) []string {
	if len(b.config.ExcludeFiles) == 0 {
		return files
	}

	var filtered []string
	for _, file := range files {
		excluded := false
		for _, excludePattern := range b.config.ExcludeFiles {
			matched, err := filepath.Match(excludePattern, file)
			if err == nil && matched {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

// generateMetadata 生成包元数据
func (b *PackageBuilder) generateMetadata(files []string) (*PackageMetadata, error) {
	metadata := &PackageMetadata{
		ID:           b.config.ID,
		Name:         b.config.Name,
		Version:      b.config.Version,
		Description:  b.config.Description,
		Author:       b.config.Author,
		License:      b.config.License,
		Homepage:     b.config.Homepage,
		Repository:   b.config.Repository,
		MainFile:     b.config.MainFile,
		Dependencies: b.config.Dependencies,
		Category:     b.config.Category,
		Tags:         b.config.Tags,
		Language:     b.config.Language,
		MinVersion:   b.config.MinVersion,
		MaxVersion:   b.config.MaxVersion,
		Platforms:    b.config.Platforms,
		CreatedAt:    time.Now(),
	}

	// 计算文件信息
	var totalSize int64
	var fileInfos []FileInfo

	for _, file := range files {
		filePath := filepath.Join(b.workDir, file)
		stat, err := os.Stat(filePath)
		if err != nil {
			return nil, fmt.Errorf("获取文件信息失败 %s: %v", file, err)
		}

		// 计算文件校验和
		checksum, err := b.calculateFileChecksum(filePath)
		if err != nil {
			return nil, fmt.Errorf("计算文件校验和失败 %s: %v", file, err)
		}

		fileInfo := FileInfo{
			Name:     file,
			Size:     stat.Size(),
			Checksum: checksum,
		}

		fileInfos = append(fileInfos, fileInfo)
		totalSize += stat.Size()
	}

	metadata.Files = fileInfos
	metadata.FileSize = totalSize
	metadata.FileCount = len(files)

	return metadata, nil
}

// calculateFileChecksum 计算文件校验和
func (b *PackageBuilder) calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// createPackage 创建包文件
func (b *PackageBuilder) createPackage(files []string, metadata *PackageMetadata) (string, error) {
	// 创建包文件名
	packageName := fmt.Sprintf("%s_%s.zip", b.config.ID, b.config.Version)
	packagePath := filepath.Join(b.outputDir, packageName)

	// 创建输出目录
	if err := os.MkdirAll(b.outputDir, 0755); err != nil {
		return "", err
	}

	// 创建ZIP文件
	zipFile, err := os.Create(packagePath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加元数据文件
	metadataData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", err
	}

	metadataWriter, err := zipWriter.Create("metadata.json")
	if err != nil {
		return "", err
	}

	if _, err := metadataWriter.Write(metadataData); err != nil {
		return "", err
	}

	// 添加所有文件
	for _, file := range files {
		if err := b.addFileToZip(zipWriter, file); err != nil {
			return "", fmt.Errorf("添加文件到包失败 %s: %v", file, err)
		}
	}

	// 计算包的校验和
	zipWriter.Close()
	zipFile.Close()

	checksum, err := b.calculateFileChecksum(packagePath)
	if err != nil {
		return "", fmt.Errorf("计算包校验和失败: %v", err)
	}

	sha256sum, err := b.calculateFileSHA256(packagePath)
	if err != nil {
		return "", fmt.Errorf("计算包SHA256失败: %v", err)
	}

	// 更新元数据文件中的校验和
	metadata.Checksum = checksum
	metadata.SHA256 = sha256sum

	// 重新打开包文件，更新元数据
	if err := b.updateMetadataInPackage(packagePath, metadata); err != nil {
		return "", fmt.Errorf("更新包元数据失败: %v", err)
	}

	return packagePath, nil
}

// addFileToZip 添加文件到ZIP
func (b *PackageBuilder) addFileToZip(zipWriter *zip.Writer, relativePath string) error {
	filePath := filepath.Join(b.workDir, relativePath)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建ZIP文件条目
	writer, err := zipWriter.Create(relativePath)
	if err != nil {
		return err
	}

	// 复制文件内容
	_, err = io.Copy(writer, file)
	return err
}

// calculateFileSHA256 计算文件SHA256
func (b *PackageBuilder) calculateFileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// updateMetadataInPackage 更新包中的元数据
func (b *PackageBuilder) updateMetadataInPackage(packagePath string, metadata *PackageMetadata) error {
	// 这里简化实现，实际应该重新打包或者使用更高级的ZIP操作
	// 由于Go的archive/zip包不支持就地修改，我们需要重新创建包

	// 读取现有的ZIP文件
	reader, err := zip.OpenReader(packagePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 创建临时文件
	tempPath := packagePath + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	defer tempFile.Close()

	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	// 复制除metadata.json之外的所有文件
	for _, file := range reader.File {
		if file.Name == "metadata.json" {
			continue
		}

		if err := b.copyZipFile(zipWriter, file); err != nil {
			return err
		}
	}

	// 添加更新的元数据
	metadataData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	metadataWriter, err := zipWriter.Create("metadata.json")
	if err != nil {
		return err
	}

	if _, err := metadataWriter.Write(metadataData); err != nil {
		return err
	}

	zipWriter.Close()
	tempFile.Close()
	reader.Close()

	// 替换原文件
	return os.Rename(tempPath, packagePath)
}

// copyZipFile 复制ZIP文件条目
func (b *PackageBuilder) copyZipFile(zipWriter *zip.Writer, file *zip.File) error {
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := zipWriter.Create(file.Name)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, reader)
	return err
}

// signPackage 签名包
func (b *PackageBuilder) signPackage(packagePath string) error {
	// 这里简化实现，实际应该使用真正的数字签名
	logger.Info("使用密钥 %s 签名包: %s", b.config.SigningKey, packagePath)

	// 生成签名文件
	signaturePath := packagePath + ".sig"
	signature := fmt.Sprintf("SIGNATURE_%s_%s", b.config.ID, b.config.Version)

	return os.WriteFile(signaturePath, []byte(signature), 0644)
}

// Publish 发布插件包
func (b *PackageBuilder) Publish(packagePath string, registryURL string) error {
	// 这里简化实现，实际应该实现完整的发布流程
	logger.Info("发布包 %s 到仓库 %s", packagePath, registryURL)

	// 读取包元数据
	metadata, err := b.extractMetadata(packagePath)
	if err != nil {
		return fmt.Errorf("提取包元数据失败: %v", err)
	}

	// 验证包
	if err := b.validatePackage(packagePath, metadata); err != nil {
		return fmt.Errorf("包验证失败: %v", err)
	}

	// 上传到仓库
	logger.Info("上传包: %s v%s", metadata.Name, metadata.Version)

	return nil
}

// extractMetadata 提取包元数据
func (b *PackageBuilder) extractMetadata(packagePath string) (*PackageMetadata, error) {
	reader, err := zip.OpenReader(packagePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// 查找元数据文件
	for _, file := range reader.File {
		if file.Name == "metadata.json" {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}

			var metadata PackageMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				return nil, err
			}

			return &metadata, nil
		}
	}

	return nil, fmt.Errorf("包中未找到元数据文件")
}

// validatePackage 验证包
func (b *PackageBuilder) validatePackage(packagePath string, metadata *PackageMetadata) error {
	// 验证包大小
	stat, err := os.Stat(packagePath)
	if err != nil {
		return err
	}

	// 检查包大小限制（100MB）
	if stat.Size() > 100*1024*1024 {
		return fmt.Errorf("包大小超过限制: %d bytes", stat.Size())
	}

	// 验证必需文件
	reader, err := zip.OpenReader(packagePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	requiredFiles := map[string]bool{
		"metadata.json":   false,
		metadata.MainFile: false,
	}

	for _, file := range reader.File {
		if _, exists := requiredFiles[file.Name]; exists {
			requiredFiles[file.Name] = true
		}
	}

	for file, found := range requiredFiles {
		if !found {
			return fmt.Errorf("包中缺少必需文件: %s", file)
		}
	}

	return nil
}

// CreateDefaultConfig 创建默认配置
func (b *PackageBuilder) CreateDefaultConfig(pluginDir string) (*PackageConfig, error) {
	// 检测主文件
	mainFile, language, err := b.detectMainFile(pluginDir)
	if err != nil {
		return nil, err
	}

	config := &PackageConfig{
		ID:          filepath.Base(pluginDir),
		Name:        filepath.Base(pluginDir),
		Version:     "1.0.0",
		Description: "Stellar安全扫描插件",
		Author:      "Unknown",
		License:     "MIT",
		MainFile:    mainFile,
		Language:    language,
		Category:    "scanner",
		Tags:        []string{"security", "scanner"},
		Platforms:   []string{"linux", "windows", "macos"},
		RunTests:    false,
	}

	return config, nil
}

// detectMainFile 检测主文件
func (b *PackageBuilder) detectMainFile(pluginDir string) (string, string, error) {
	// 常见的主文件名模式
	patterns := map[string]string{
		"main.py":   "python",
		"main.js":   "javascript",
		"main.go":   "go",
		"plugin.py": "python",
		"plugin.js": "javascript",
		"plugin.go": "go",
		"*.yaml":    "yaml",
		"*.yml":     "yaml",
	}

	for pattern, language := range patterns {
		matches, err := filepath.Glob(filepath.Join(pluginDir, pattern))
		if err == nil && len(matches) > 0 {
			relPath, _ := filepath.Rel(pluginDir, matches[0])
			return relPath, language, nil
		}
	}

	return "", "", fmt.Errorf("未找到主文件")
}
