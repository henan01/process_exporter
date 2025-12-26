package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

// ProcessInfo 存储进程信息
type ProcessInfo struct {
	PID            int
	Name           string
	Cmdline        string
	MemoryBytes    uint64
	MemoryPercent  float64
	CPUPercent     float64
	CPUTime        uint64
	RuntimeSeconds int64
}

// HostInfo 存储主机标识信息
type HostInfo struct {
	Hostname string
	IP       string
	MAC      string
}

// SystemMemoryInfo 存储系统内存信息
type SystemMemoryInfo struct {
	TotalBytes     uint64  // 总内存（字节）
	UsedBytes      uint64  // 已用内存（字节）
	AvailableBytes uint64  // 可用内存（字节）
	UsedPercent    float64 // 使用率（百分比）
}

// customLabels 用于存储自定义标签
type customLabels map[string]string

func (c *customLabels) String() string {
	if c == nil || len(*c) == 0 {
		return ""
	}
	var parts []string
	for k, v := range *c {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
}

func (c *customLabels) Set(value string) error {
	if *c == nil {
		*c = make(map[string]string)
	}
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("标签格式错误，应为 key=value: %s", value)
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	if key == "" {
		return fmt.Errorf("标签 key 不能为空")
	}
	(*c)[key] = val
	return nil
}

var (
	remoteURL          = flag.String("remote.url", "", "Remote write endpoint URL (required)")
	remoteUsername     = flag.String("remote.username", "", "Username for basic auth (optional)")
	remotePassword     = flag.String("remote.password", "", "Password for basic auth (optional)")
	interval           = flag.Duration("interval", 60*time.Second, "Scrape and push interval")
	topN               = flag.Int("top", 10, "Number of top processes to monitor")
	retryTimes         = flag.Int("retry", 1, "Number of retries on push failure")
	retryDelay         = flag.Duration("retry.delay", 5*time.Second, "Delay between retries")
	insecureSkipVerify = flag.Bool("insecure-skip-verify", false, "Skip TLS certificate verification (insecure, use only for testing)")
	labels             customLabels
)

func main() {
	// 注册自定义标签参数
	flag.Var(&labels, "label", "Custom label in key=value format (can be specified multiple times)")
	flag.Parse()

	if *remoteURL == "" {
		log.Fatal("错误：必须指定 --remote.url 参数")
	}

	log.Printf("启动 VmAgent Process Exporter")
	log.Printf("远程端点: %s", *remoteURL)
	log.Printf("采集间隔: %v", *interval)
	log.Printf("监控进程数: %d", *topN)
	log.Printf("失败重试次数: %d", *retryTimes)
	log.Printf("重试间隔: %v", *retryDelay)
	if *insecureSkipVerify {
		log.Printf("警告: 已禁用 TLS 证书验证 (不安全)")
	}

	// 获取主机信息
	hostInfo := getHostInfo()
	log.Printf("主机标识 - 主机名: %s, IP: %s, MAC: %s", hostInfo.Hostname, hostInfo.IP, hostInfo.MAC)

	// 显示自定义标签
	if len(labels) > 0 {
		log.Printf("自定义标签: %s", labels.String())
	}

	// 立即执行一次
	collectAndPush(hostInfo, labels)

	// 定时执行
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for range ticker.C {
		collectAndPush(hostInfo, labels)
	}
}

// collectAndPush 收集指标并推送到远程端点(带重试)
func collectAndPush(hostInfo HostInfo, customLabels customLabels) {
	log.Printf("开始收集进程指标...")

	topMemProcs, topCPUProcs, err := getTopProcesses(*topN)
	if err != nil {
		log.Printf("错误：获取进程信息失败: %v", err)
		return
	}

	// 构建 Prometheus 远程写入请求
	timeseries := buildTimeSeries(topMemProcs, topCPUProcs, hostInfo, customLabels)

	// 发送到远程端点,失败时重试
	var lastErr error
	for attempt := 0; attempt <= *retryTimes; attempt++ {
		if attempt > 0 {
			log.Printf("第 %d 次重试推送指标...", attempt)
			time.Sleep(*retryDelay)
		}

		if err := sendRemoteWrite(timeseries); err != nil {
			lastErr = err
			log.Printf("错误：推送指标失败: %v", err)
			continue
		}

		// 成功推送
		log.Printf("成功推送 %d 个时间序列", len(timeseries))
		return
	}

	// 所有重试都失败
	log.Printf("错误：推送指标失败,已重试 %d 次: %v", *retryTimes, lastErr)
}

// buildTimeSeries 构建 Prometheus 时间序列
func buildTimeSeries(topMemProcs, topCPUProcs []ProcessInfo, hostInfo HostInfo, customLabels customLabels) []prompb.TimeSeries {
	var timeseries []prompb.TimeSeries
	now := time.Now().UnixNano() / int64(time.Millisecond)

	// 合并所有进程用于运行时间指标
	allProcs := mergeProcesses(topMemProcs, topCPUProcs)

	// 公共标签（主机标识）
	commonLabels := []prompb.Label{
		{Name: "hostname", Value: hostInfo.Hostname},
		{Name: "ip", Value: hostInfo.IP},
		{Name: "mac", Value: hostInfo.MAC},
	}

	// 添加自定义标签
	for key, value := range customLabels {
		commonLabels = append(commonLabels, prompb.Label{Name: key, Value: value})
	}

	// 内存使用量指标
	for i, proc := range topMemProcs {
		labels := append(commonLabels,
			prompb.Label{Name: "__name__", Value: "process_memory_bytes"},
			prompb.Label{Name: "pid", Value: strconv.Itoa(proc.PID)},
			prompb.Label{Name: "name", Value: proc.Name},
			prompb.Label{Name: "cmdline", Value: escapeLabelValue(proc.Cmdline)},
			prompb.Label{Name: "rank", Value: strconv.Itoa(i + 1)},
		)
		timeseries = append(timeseries, prompb.TimeSeries{
			Labels:  labels,
			Samples: []prompb.Sample{{Value: float64(proc.MemoryBytes), Timestamp: now}},
		})
	}

	// 内存百分比指标
	for i, proc := range topMemProcs {
		labels := append(commonLabels,
			prompb.Label{Name: "__name__", Value: "process_memory_percent"},
			prompb.Label{Name: "pid", Value: strconv.Itoa(proc.PID)},
			prompb.Label{Name: "name", Value: proc.Name},
			prompb.Label{Name: "cmdline", Value: escapeLabelValue(proc.Cmdline)},
			prompb.Label{Name: "rank", Value: strconv.Itoa(i + 1)},
		)
		timeseries = append(timeseries, prompb.TimeSeries{
			Labels:  labels,
			Samples: []prompb.Sample{{Value: proc.MemoryPercent, Timestamp: now}},
		})
	}

	// CPU 使用率指标
	for i, proc := range topCPUProcs {
		labels := append(commonLabels,
			prompb.Label{Name: "__name__", Value: "process_cpu_percent"},
			prompb.Label{Name: "pid", Value: strconv.Itoa(proc.PID)},
			prompb.Label{Name: "name", Value: proc.Name},
			prompb.Label{Name: "cmdline", Value: escapeLabelValue(proc.Cmdline)},
			prompb.Label{Name: "rank", Value: strconv.Itoa(i + 1)},
		)
		timeseries = append(timeseries, prompb.TimeSeries{
			Labels:  labels,
			Samples: []prompb.Sample{{Value: proc.CPUPercent, Timestamp: now}},
		})
	}

	// 运行时间指标
	for _, proc := range allProcs {
		labels := append(commonLabels,
			prompb.Label{Name: "__name__", Value: "process_runtime_seconds"},
			prompb.Label{Name: "pid", Value: strconv.Itoa(proc.PID)},
			prompb.Label{Name: "name", Value: proc.Name},
			prompb.Label{Name: "cmdline", Value: escapeLabelValue(proc.Cmdline)},
		)
		timeseries = append(timeseries, prompb.TimeSeries{
			Labels:  labels,
			Samples: []prompb.Sample{{Value: float64(proc.RuntimeSeconds), Timestamp: now}},
		})
	}

	// 系统内存指标
	sysMemInfo := getSystemMemoryInfo()

	// 系统总内存
	sysMemTotalLabels := append([]prompb.Label{},
		prompb.Label{Name: "__name__", Value: "system_memory_total_bytes"},
	)
	sysMemTotalLabels = append(sysMemTotalLabels, commonLabels...)
	timeseries = append(timeseries, prompb.TimeSeries{
		Labels:  sysMemTotalLabels,
		Samples: []prompb.Sample{{Value: float64(sysMemInfo.TotalBytes), Timestamp: now}},
	})

	// 系统已用内存
	sysMemUsedLabels := append([]prompb.Label{},
		prompb.Label{Name: "__name__", Value: "system_memory_used_bytes"},
	)
	sysMemUsedLabels = append(sysMemUsedLabels, commonLabels...)
	timeseries = append(timeseries, prompb.TimeSeries{
		Labels:  sysMemUsedLabels,
		Samples: []prompb.Sample{{Value: float64(sysMemInfo.UsedBytes), Timestamp: now}},
	})

	// 系统可用内存
	sysMemAvailLabels := append([]prompb.Label{},
		prompb.Label{Name: "__name__", Value: "system_memory_available_bytes"},
	)
	sysMemAvailLabels = append(sysMemAvailLabels, commonLabels...)
	timeseries = append(timeseries, prompb.TimeSeries{
		Labels:  sysMemAvailLabels,
		Samples: []prompb.Sample{{Value: float64(sysMemInfo.AvailableBytes), Timestamp: now}},
	})

	// 系统内存使用率
	sysMemPercentLabels := append([]prompb.Label{},
		prompb.Label{Name: "__name__", Value: "system_memory_used_percent"},
	)
	sysMemPercentLabels = append(sysMemPercentLabels, commonLabels...)
	timeseries = append(timeseries, prompb.TimeSeries{
		Labels:  sysMemPercentLabels,
		Samples: []prompb.Sample{{Value: sysMemInfo.UsedPercent, Timestamp: now}},
	})

	return timeseries
}

// sendRemoteWrite 发送远程写入请求
func sendRemoteWrite(timeseries []prompb.TimeSeries) error {
	req := &prompb.WriteRequest{
		Timeseries: timeseries,
	}

	data, err := req.Marshal()
	if err != nil {
		return fmt.Errorf("marshal 失败: %v", err)
	}

	// Snappy 压缩
	compressed := snappy.Encode(nil, data)

	httpReq, err := http.NewRequest("POST", *remoteURL, bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	httpReq.Header.Set("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	// 添加基本认证
	if *remoteUsername != "" && *remotePassword != "" {
		httpReq.SetBasicAuth(*remoteUsername, *remotePassword)
	}

	// 配置 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: *insecureSkipVerify,
			},
		},
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("远程端点返回错误状态 %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// getHostInfo 获取主机标识信息
func getHostInfo() HostInfo {
	info := HostInfo{
		Hostname: "unknown",
		IP:       "unknown",
		MAC:      "unknown",
	}

	// 获取主机名
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	// 获取主要网卡的 IP 和 MAC
	interfaces, err := net.Interfaces()
	if err != nil {
		return info
	}

	for _, iface := range interfaces {
		// 跳过回环接口和未启用的接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 获取 IP 地址
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 只取 IPv4 地址，且非回环地址
			if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
				info.IP = ip.String()
				info.MAC = iface.HardwareAddr.String()
				return info
			}
		}
	}

	return info
}

// getTopProcesses 获取内存和CPU占用最高的前N个进程
func getTopProcesses(n int) ([]ProcessInfo, []ProcessInfo, error) {
	procDir := "/proc"

	entries, err := ioutil.ReadDir(procDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read /proc: %v", err)
	}

	var processes []ProcessInfo
	totalMemory := getTotalMemory()
	systemUptime := getSystemUptime()
	totalCPUTime := getTotalCPUTime()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 检查目录名是否为数字（PID）
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		proc, err := getProcessInfo(pid, totalMemory, systemUptime, totalCPUTime)
		if err != nil {
			// 跳过无法读取的进程（可能已退出或无权限）
			continue
		}

		processes = append(processes, proc)
	}

	// 复制一份用于CPU排序
	cpuProcesses := make([]ProcessInfo, len(processes))
	copy(cpuProcesses, processes)

	// 按内存使用量排序
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].MemoryBytes > processes[j].MemoryBytes
	})

	// 按CPU使用率排序
	sort.Slice(cpuProcesses, func(i, j int) bool {
		return cpuProcesses[i].CPUPercent > cpuProcesses[j].CPUPercent
	})

	// 返回前N个内存进程
	topMemProcs := processes
	if len(topMemProcs) > n {
		topMemProcs = topMemProcs[:n]
	}

	// 返回前N个CPU进程
	topCPUProcs := cpuProcesses
	if len(topCPUProcs) > n {
		topCPUProcs = topCPUProcs[:n]
	}

	return topMemProcs, topCPUProcs, nil
}

// mergeProcesses 合并两个进程列表并去重（按PID）
func mergeProcesses(list1, list2 []ProcessInfo) []ProcessInfo {
	seen := make(map[int]bool)
	var result []ProcessInfo

	for _, proc := range list1 {
		if !seen[proc.PID] {
			seen[proc.PID] = true
			result = append(result, proc)
		}
	}

	for _, proc := range list2 {
		if !seen[proc.PID] {
			seen[proc.PID] = true
			result = append(result, proc)
		}
	}

	return result
}

// getProcessInfo 获取单个进程的信息
func getProcessInfo(pid int, totalMemory uint64, systemUptime int64, totalCPUTime uint64) (ProcessInfo, error) {
	proc := ProcessInfo{PID: pid}

	// 读取进程名称
	commPath := filepath.Join("/proc", strconv.Itoa(pid), "comm")
	commData, err := ioutil.ReadFile(commPath)
	if err != nil {
		return proc, err
	}
	proc.Name = strings.TrimSpace(string(commData))

	// 读取完整命令行
	cmdlinePath := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")
	cmdlineData, err := ioutil.ReadFile(cmdlinePath)
	if err != nil {
		return proc, err
	}
	// cmdline 中参数由 null 字符分隔，替换为空格
	cmdline := strings.ReplaceAll(string(cmdlineData), "\x00", " ")
	proc.Cmdline = strings.TrimSpace(cmdline)
	if proc.Cmdline == "" {
		proc.Cmdline = "[" + proc.Name + "]"
	}

	// 读取内存信息
	statusPath := filepath.Join("/proc", strconv.Itoa(pid), "status")
	statusData, err := ioutil.ReadFile(statusPath)
	if err != nil {
		return proc, err
	}

	// 解析 VmRSS (实际物理内存使用)
	for _, line := range strings.Split(string(statusData), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				rss, _ := strconv.ParseUint(fields[1], 10, 64)
				proc.MemoryBytes = rss * 1024 // 转换为字节
				if totalMemory > 0 {
					proc.MemoryPercent = float64(proc.MemoryBytes) / float64(totalMemory) * 100
				}
			}
			break
		}
	}

	// 读取进程启动时间和 CPU 时间
	statPath := filepath.Join("/proc", strconv.Itoa(pid), "stat")
	statData, err := ioutil.ReadFile(statPath)
	if err == nil {
		fields := strings.Fields(string(statData))
		if len(fields) >= 22 {
			// 字段 13: utime (用户态 CPU 时间)
			// 字段 14: stime (内核态 CPU 时间)
			utime, _ := strconv.ParseUint(fields[13], 10, 64)
			stime, _ := strconv.ParseUint(fields[14], 10, 64)
			proc.CPUTime = utime + stime

			// 字段 21: starttime (进程启动时间)
			starttime, _ := strconv.ParseInt(fields[21], 10, 64)
			// starttime 是以时钟滴答为单位，需要转换为秒
			clockTick := int64(100) // 通常是 100 Hz
			processStartTime := starttime / clockTick
			proc.RuntimeSeconds = systemUptime - processStartTime

			// 计算 CPU 使用百分比
			// CPU% = (进程CPU时间/时钟频率) / 进程运行时间 × 100
			// 这表示进程从启动到现在的平均 CPU 使用率
			if proc.RuntimeSeconds > 0 {
				processCPUSeconds := float64(proc.CPUTime) / float64(clockTick)
				proc.CPUPercent = (processCPUSeconds / float64(proc.RuntimeSeconds)) * 100.0
			}
		}
	}

	return proc, nil
}

// getTotalMemory 获取系统总内存（字节）
func getTotalMemory() uint64 {
	data, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				total, _ := strconv.ParseUint(fields[1], 10, 64)
				return total * 1024 // 转换为字节
			}
		}
	}
	return 0
}

// getSystemMemoryInfo 获取系统内存详细信息
func getSystemMemoryInfo() SystemMemoryInfo {
	info := SystemMemoryInfo{}

	data, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return info
	}

	var memTotal, memFree, memAvailable, buffers, cached uint64

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, _ := strconv.ParseUint(fields[1], 10, 64)
		value = value * 1024 // 转换为字节

		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			memTotal = value
		case strings.HasPrefix(line, "MemFree:"):
			memFree = value
		case strings.HasPrefix(line, "MemAvailable:"):
			memAvailable = value
		case strings.HasPrefix(line, "Buffers:"):
			buffers = value
		case strings.HasPrefix(line, "Cached:"):
			cached = value
		}
	}

	info.TotalBytes = memTotal

	// 优先使用 MemAvailable（Linux 3.14+）
	if memAvailable > 0 {
		info.AvailableBytes = memAvailable
		info.UsedBytes = memTotal - memAvailable
	} else {
		// 旧版本 Linux 的估算方法
		info.AvailableBytes = memFree + buffers + cached
		info.UsedBytes = memTotal - info.AvailableBytes
	}

	// 计算使用率
	if memTotal > 0 {
		info.UsedPercent = float64(info.UsedBytes) / float64(memTotal) * 100.0
	}

	return info
}

// getSystemUptime 获取系统运行时间（秒）
func getSystemUptime() int64 {
	data, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}

	fields := strings.Fields(string(data))
	if len(fields) >= 1 {
		uptime, _ := strconv.ParseFloat(fields[0], 64)
		return int64(uptime)
	}
	return 0
}

// getTotalCPUTime 获取系统总 CPU 时间（jiffies）
func getTotalCPUTime() uint64 {
	data, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return 0
	}

	// 读取第一行 cpu 行
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				return 0
			}

			// CPU 时间字段：user, nice, system, idle, iowait, irq, softirq, ...
			// 累加所有时间得到总 CPU 时间
			var total uint64
			for i := 1; i < len(fields); i++ {
				val, _ := strconv.ParseUint(fields[i], 10, 64)
				total += val
			}
			return total
		}
	}
	return 0
}

// escapeLabelValue 转义 Prometheus 标签值中的特殊字符
func escapeLabelValue(s string) string {
	// 限制长度，避免命令行过长
	if len(s) > 200 {
		s = s[:200] + "..."
	}

	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
