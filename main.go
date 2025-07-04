package main

import (
	_ "embed"
	"flag"
	"fmt"
	"mac-guest-agent/internal/agent"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

var (
	version   = "1.1.0"
	daemon    = flag.Bool("daemon", false, "运行为守护进程")
	verbose   = flag.Bool("verbose", false, "启用详细日志")
	device    = flag.String("device", "", "指定virtio设备路径")
	testMode  = flag.Bool("test", false, "测试模式（使用标准输入输出模拟设备）")
	install   = flag.Bool("install", false, "安装为系统服务")
	uninstall = flag.Bool("uninstall", false, "卸载系统服务")
)

//go:embed configs/com.macos.guest-agent.plist
var plistContent []byte

const (
	serviceName = "com.macos.guest-agent"
	binaryPath  = "/usr/local/bin/mac-guest-agent"
	plistPath   = "/Library/LaunchDaemons/com.macos.guest-agent.plist"
	logPath     = "/var/log/mac-guest-agent.log"
	sharePath   = "/usr/local/share/mac-guest-agent"
)

func main() {
	flag.Parse()

	// 处理系统服务安装/卸载
	if *install {
		installService()
		return
	}

	if *uninstall {
		uninstallService()
		return
	}

	// 配置日志
	setupLogging()

	logrus.WithField("version", version).Info("macOS Guest Agent 启动中...")

	// 检测QEMU环境（测试模式下跳过检测）
	if !*testMode {
		if !isRunningInQEMU() {
			logrus.Error("检测到当前系统不是运行在QEMU虚拟化环境中")
			logrus.Error("macOS Guest Agent 仅支持在QEMU虚拟机中运行")
			logrus.Error("如果您确定要在非QEMU环境中测试，请使用 --test 参数")
			os.Exit(1)
		}
		logrus.Info("检测到QEMU虚拟化环境，继续启动...")
	}

	// 测试模式下不需要root权限
	if !*testMode && os.Geteuid() != 0 {
		logrus.Fatal("Guest Agent需要root权限运行，请使用sudo")
	}

	// 创建Agent实例
	var guestAgent *agent.Agent
	var err error

	if *testMode {
		logrus.Info("运行在测试模式下")
		guestAgent, err = agent.NewTestMode()
	} else {
		guestAgent, err = agent.New(*device)
	}

	if err != nil {
		logrus.WithError(err).Fatal("创建Guest Agent失败")
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动Agent
	go func() {
		if err := guestAgent.Start(); err != nil {
			logrus.WithError(err).Fatal("启动Guest Agent失败")
		}
	}()

	logrus.Info("Guest Agent已启动，等待命令...")

	// 等待退出信号
	<-sigChan
	logrus.Info("收到退出信号，正在关闭...")

	// 优雅关闭
	guestAgent.Stop()
	logrus.Info("Guest Agent已停止")
}

// setupLogging 配置日志
func setupLogging() {
	// 设置日志级别
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// 如果运行为守护进程，将日志写入系统日志文件
	if *daemon {
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logrus.WithError(err).Fatal("无法打开日志文件")
		}
		logrus.SetOutput(file)

		// 设置日志格式为文本格式，提高可读性
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableColors:   true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			// 使用简洁的字段格式
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "time",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "msg",
			},
		})
	} else {
		// 非守护进程模式使用彩色输出
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
}

// installService 安装系统服务
func installService() {
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "错误: 安装系统服务需要root权限\n")
		fmt.Fprintf(os.Stderr, "请使用: sudo %s --install\n", os.Args[0])
		os.Exit(1)
	}

	fmt.Println("开始安装 macOS Guest Agent 系统服务...")

	// 停止现有服务
	stopExistingService()

	// 检查二进制文件是否存在
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 二进制文件不存在: %s\n", binaryPath)
		fmt.Fprintf(os.Stderr, "请先将编译好的二进制文件复制到该路径\n")
		os.Exit(1)
	}

	// 创建必要目录
	if err := createDirectories(); err != nil {
		fmt.Fprintf(os.Stderr, "创建目录失败: %v\n", err)
		os.Exit(1)
	}

	// 安装LaunchDaemon配置
	if err := installPlist(); err != nil {
		fmt.Fprintf(os.Stderr, "安装LaunchDaemon配置失败: %v\n", err)
		os.Exit(1)
	}

	// 创建日志文件
	if err := createLogFile(); err != nil {
		fmt.Fprintf(os.Stderr, "创建日志文件失败: %v\n", err)
		os.Exit(1)
	}

	// 加载并启动服务
	if err := loadService(); err != nil {
		fmt.Fprintf(os.Stderr, "启动服务失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ macOS Guest Agent 系统服务安装成功!")
	fmt.Printf("  - 可执行文件: %s\n", binaryPath)
	fmt.Printf("  - 配置文件: %s\n", plistPath)
	fmt.Printf("  - 日志文件: %s\n", logPath)
	fmt.Println("")
	fmt.Println("服务管理命令:")
	fmt.Printf("  查看状态: sudo launchctl list %s\n", serviceName)
	fmt.Printf("  查看日志: tail -f %s\n", logPath)
	fmt.Printf("  停止服务: sudo launchctl stop %s\n", serviceName)
	fmt.Printf("  启动服务: sudo launchctl start %s\n", serviceName)
	fmt.Printf("  卸载服务: %s --uninstall\n", os.Args[0])
}

// uninstallService 卸载系统服务
func uninstallService() {
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "错误: 卸载系统服务需要root权限\n")
		fmt.Fprintf(os.Stderr, "请使用: sudo %s --uninstall\n", os.Args[0])
		os.Exit(1)
	}

	fmt.Println("开始卸载 macOS Guest Agent 系统服务...")

	// 停止并卸载服务
	stopExistingService()
	unloadService()

	// 删除文件
	removeFiles()

	fmt.Println("✓ macOS Guest Agent 系统服务卸载完成!")
}

// stopExistingService 停止现有服务
func stopExistingService() {
	// 检查服务是否存在
	if !isServiceLoaded() {
		return
	}

	fmt.Println("停止现有服务...")
	exec.Command("launchctl", "stop", serviceName).Run()
	exec.Command("launchctl", "unload", plistPath).Run()
}

// createDirectories 创建必要目录
func createDirectories() error {
	fmt.Println("创建系统目录...")

	dirs := []string{
		"/usr/local/bin",
		"/usr/local/share",
		sharePath,
		filepath.Dir(logPath),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %v", dir, err)
		}
	}

	return nil
}

// installPlist 安装LaunchDaemon配置
func installPlist() error {
	fmt.Println("安装LaunchDaemon配置...")

	// 将嵌入的plist内容写入目标文件
	err := os.WriteFile(plistPath, plistContent, 0644)
	if err != nil {
		return fmt.Errorf("写入LaunchDaemon配置文件失败: %v", err)
	}

	fmt.Printf("已将配置写入: %s\n", plistPath)

	return nil
}

// createLogFile 创建日志文件
func createLogFile() error {
	fmt.Println("创建日志文件...")

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	file.Close()

	return nil
}

// loadService 加载并启动服务
func loadService() error {
	fmt.Println("加载并启动服务...")

	// 加载服务
	cmd := exec.Command("launchctl", "load", plistPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("加载服务失败: %v", err)
	}

	// 启动服务
	cmd = exec.Command("launchctl", "start", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启动服务失败: %v", err)
	}

	// 验证服务状态
	if !isServiceLoaded() {
		return fmt.Errorf("服务启动验证失败")
	}

	return nil
}

// unloadService 卸载服务
func unloadService() {
	if isServiceLoaded() {
		fmt.Println("卸载LaunchDaemon服务...")
		exec.Command("launchctl", "unload", plistPath).Run()
	}
}

// removeFiles 删除安装的文件
func removeFiles() {
	files := []string{binaryPath, plistPath}
	dirs := []string{sharePath}

	fmt.Println("删除安装文件...")
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			os.Remove(file)
			fmt.Printf("已删除: %s\n", file)
		}
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); err == nil {
			os.RemoveAll(dir)
			fmt.Printf("已删除: %s\n", dir)
		}
	}

	// 询问是否删除日志文件
	if _, err := os.Stat(logPath); err == nil {
		fmt.Printf("是否删除日志文件 %s? [y/N]: ", logPath)
		var response string
		fmt.Scanln(&response)
		if response == "y" || response == "Y" {
			os.Remove(logPath)
			fmt.Println("已删除日志文件")
		} else {
			fmt.Println("保留日志文件")
		}
	}
}

// isServiceLoaded 检查服务是否已加载
func isServiceLoaded() bool {
	cmd := exec.Command("launchctl", "list", serviceName)
	return cmd.Run() == nil
}

// isRunningInQEMU 检查是否运行在QEMU虚拟化环境中
func isRunningInQEMU() bool {
	// 方法1: 检查硬件模型信息
	if checkHardwareModel() {
		return true
	}

	// 方法2: 检查virtio设备是否存在
	if checkVirtioDevices() {
		return true
	}

	// 方法3: 检查系统信息中的虚拟化标识
	if checkSystemProfiler() {
		return true
	}

	return false
}

// checkHardwareModel 检查硬件模型是否为QEMU
func checkHardwareModel() bool {
	cmd := exec.Command("sysctl", "-n", "hw.model")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	model := strings.TrimSpace(string(output))
	logrus.WithField("hw.model", model).Debug("检查硬件模型")

	// QEMU通常会显示包含"QEMU"的硬件模型信息
	return strings.Contains(strings.ToUpper(model), "QEMU")
}

// checkVirtioDevices 检查virtio设备是否存在
func checkVirtioDevices() bool {
	// 检查QEMU guest agent设备文件
	qemuDevices := []string{
		"/dev/cu.org.qemu.guest_agent.0",
		"/dev/tty.org.qemu.guest_agent.0",
	}

	for _, device := range qemuDevices {
		if _, err := os.Stat(device); err == nil {
			logrus.WithField("device", device).Debug("找到QEMU设备文件")
			return true
		}
	}

	// 检查virtio相关设备
	virtioDevs := []string{
		"/dev/virtio",
		"/dev/vda",
		"/dev/vdb",
	}

	for _, device := range virtioDevs {
		if _, err := os.Stat(device); err == nil {
			logrus.WithField("device", device).Debug("找到virtio设备")
			return true
		}
	}

	return false
}

// checkSystemProfiler 通过系统信息检查虚拟化环境
func checkSystemProfiler() bool {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	content := strings.ToUpper(string(output))
	logrus.Debug("检查系统硬件信息")

	// 检查是否包含QEMU或虚拟化相关关键词
	keywords := []string{"QEMU", "VIRTUAL", "VIRTUALIZATION"}
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			logrus.WithField("keyword", keyword).Debug("在系统信息中找到虚拟化标识")
			return true
		}
	}

	return false
}
