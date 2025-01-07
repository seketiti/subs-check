package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bestruirui/mihomo-check/check"
	"github.com/bestruirui/mihomo-check/config"
	"github.com/bestruirui/mihomo-check/proxy/ipinfo"
	"github.com/bestruirui/mihomo-check/save"
	"github.com/bestruirui/mihomo-check/utils"
	"github.com/metacubex/mihomo/log"
	"gopkg.in/yaml.v3"
)

// App 结构体用于管理应用程序状态
type App struct {
	configPath string
	interval   int
}

// NewApp 创建新的应用实例
func NewApp() *App {
	configPath := flag.String("f", "", "配置文件路径")
	flag.Parse()

	return &App{
		configPath: *configPath,
	}
}

// Initialize 初始化应用程序
func (app *App) Initialize() error {
	// 初始化配置文件路径
	if err := app.initConfigPath(); err != nil {
		return fmt.Errorf("初始化配置文件路径失败: %w", err)
	}

	// 加载配置文件
	if err := app.loadConfig(); err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}

	// 初始化IP数据库
	ipinfo.GetIPdb()

	app.interval = config.GlobalConfig.CheckInterval
	return nil
}

// initConfigPath 初始化配置文件路径
func (app *App) initConfigPath() error {
	if app.configPath == "" {
		execPath := utils.GetExecutablePath()
		configDir := filepath.Join(execPath, "config")

		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("创建配置目录失败: %w", err)
		}

		app.configPath = filepath.Join(configDir, "config.yaml")
	}
	return nil
}

// loadConfig 加载配置文件
func (app *App) loadConfig() error {
	yamlFile, err := os.ReadFile(app.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return app.createDefaultConfig()
		}
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(yamlFile, &config.GlobalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	log.Infoln("配置文件读取成功")
	return nil
}

// createDefaultConfig 创建默认配置文件
func (app *App) createDefaultConfig() error {
	log.Infoln("配置文件不存在，创建默认配置文件")

	if err := os.WriteFile(app.configPath, []byte(config.DefaultConfigTemplate), 0644); err != nil {
		return fmt.Errorf("写入默认配置文件失败: %w", err)
	}

	log.Infoln("默认配置文件创建成功")
	log.Infoln("请编辑配置文件: %v", app.configPath)
	os.Exit(0)
	return nil
}

// Run 运行应用程序主循环
func (app *App) Run() {
	log.Infoln("进度展示: %v", config.GlobalConfig.PrintProgress)

	for {
		if err := app.checkProxies(); err != nil {
			log.Errorln("检测代理失败: %v", err)
			os.Exit(1)
		}

		nextCheck := time.Now().Add(time.Duration(app.interval) * time.Minute)
		log.Infoln("下次检查时间: %v", nextCheck.Format("2006-01-02 15:04:05"))
		time.Sleep(time.Duration(app.interval) * time.Minute)
	}
}

// checkProxies 执行代理检测
func (app *App) checkProxies() error {
	log.Infoln("开始检测代理")

	results, err := check.Check()
	if err != nil {
		return fmt.Errorf("检测代理失败: %w", err)
	}

	log.Infoln("检测完成")
	save.SaveConfig(results)
	return nil
}

func main() {
	app := NewApp()

	if err := app.Initialize(); err != nil {
		log.Errorln("初始化失败: %v", err)
		os.Exit(1)
	}

	app.Run()
}
