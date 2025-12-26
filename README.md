# Process Exporter

用于监控服务器内存和 CPU 占用最高的进程，并直接推送指标到远程 **VictoriaMetrics** 或 **Prometheus** 的导出器。

## 特性

- ✅ 监控内存和 CPU 占用最高的进程（默认 Top 10）
- ✅ **系统内存监控** - 总内存、已用内存、可用内存和使用率
- ✅ **自动采集本机标识**（主机名、IP 地址、MAC 地址）作为指标标签
- ✅ **自定义标签** - 支持添加任意自定义标签，方便查询和分组
- ✅ **直接推送**指标到远程 VictoriaMetrics 或 Prometheus 端点
- ✅ **兼容性** - 同时支持 VictoriaMetrics 和 Prometheus（使用 Prometheus 远程写入 API）
- ✅ 支持可配置的采集间隔
- ✅ 支持基本认证（用户名/密码）
- ✅ 单一可执行文件，无需额外依赖
- ✅ 无需单独部署 vmagent

## 快速开始

### 编译

支持编译 **Linux AMD64**、**ARM64** 和 **ARM32** 三种架构：

```bash
# 使用编译脚本（推荐）
cd /path/to/process_exporter
chmod +x build.sh
./build.sh
```

编译完成后，会在 `build/` 目录下生成三个可执行文件：

- `process_exporter_linux_amd64` - AMD64 架构（x86_64）
- `process_exporter_linux_arm64` - ARM64 架构（aarch64）
- `process_exporter_linux_arm32` - ARM32 架构（ARMv7）

或手动编译特定架构：

```bash
# AMD64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o process_exporter_amd64 main.go

# ARM64
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o process_exporter_arm64 main.go

# ARM32 (ARMv7)
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o process_exporter_arm32 main.go
```

### 运行

```bash
# 基本用法（必须指定远程端点）
./process_exporter --remote.url=http://victoriametrics:8428/api/v1/write

# 完整配置示例
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --remote.username=admin \
  --remote.password=secret \
  --interval=30s \
  --top=20
```

### 配置参数

| 参数                | 说明                                     | 默认值 | 必需  |
| ------------------- | ---------------------------------------- | ------ | ----- |
| `--remote.url`      | 远程写入端点 URL（支持 VictoriaMetrics 和 Prometheus） | 无     | ✅ 是 |
| `--remote.username` | 基本认证用户名                           | 空     | ❌ 否 |
| `--remote.password` | 基本认证密码                             | 空     | ❌ 否 |
| `--interval`        | 采集和推送间隔                           | 60s    | ❌ 否 |
| `--top`             | 监控的 Top 进程数量                      | 10     | ❌ 否 |
| `--label`           | 自定义标签（key=value 格式，可多次指定） | 无     | ❌ 否 |

### 示例

#### 1. 推送到本地 VictoriaMetrics 或 Prometheus

**推送到 VictoriaMetrics：**

```bash
# 最简单的用法，推送到本地 VictoriaMetrics
./process_exporter --remote.url=http://localhost:8428/api/v1/write

# 查看运行日志
./process_exporter --remote.url=http://localhost:8428/api/v1/write 2>&1 | tee exporter.log
```

**推送到 Prometheus（使用 Prometheus Remote Write API）：**

```bash
# 推送到本地 Prometheus（需要配置 remote_write）
./process_exporter --remote.url=http://localhost:9090/api/v1/write

# 推送到远程 Prometheus
./process_exporter --remote.url=http://prometheus:9090/api/v1/write

# 推送到 Prometheus 并添加标签
./process_exporter \
  --remote.url=http://prometheus:9090/api/v1/write \
  --label env=production \
  --label region=us-east-1
```

> **注意**：Prometheus 需要配置 `remote_write` 接收器（如使用 Prometheus Agent 模式或配置远程写入端点）。VictoriaMetrics 原生支持 `/api/v1/write` 端点。

#### 2. 推送到远程 VictoriaMetrics 或 Prometheus（带认证）

**VictoriaMetrics 示例：**

```bash
# HTTPS + 基本认证
./process_exporter \
  --remote.url=https://vm.example.com/api/v1/write \
  --remote.username=monitor \
  --remote.password=P@ssw0rd \
  --interval=30s

# HTTP + 基本认证（内网环境）
./process_exporter \
  --remote.url=http://192.168.1.100:8428/api/v1/write \
  --remote.username=admin \
  --remote.password=secret123 \
  --interval=60s \
  --top=10
```

**Prometheus 示例：**

```bash
# 推送到 Prometheus（带基本认证）
./process_exporter \
  --remote.url=https://prometheus.example.com/api/v1/write \
  --remote.username=monitor \
  --remote.password=P@ssw0rd \
  --interval=30s

# 推送到 Prometheus（内网环境）
./process_exporter \
  --remote.url=http://192.168.1.100:9090/api/v1/write \
  --remote.username=admin \
  --remote.password=secret123 \
  --interval=60s \
  --top=10
```

#### 3. 监控 Top 20 进程，每 15 秒采集一次

```bash
# 高频采集场景（适合关键服务器）
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --interval=15s \
  --top=20

# 低频采集场景（适合资源受限的服务器）
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --interval=120s \
  --top=5
```

#### 4. 添加自定义标签（推荐）

自定义标签可以帮助您在 VictoriaMetrics 或 Prometheus 中更好地组织和查询数据：

**生产环境示例：**

```bash
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --label env=production \
  --label region=us-east-1 \
  --label cluster=web-cluster \
  --label app=myapp \
  --label team=backend \
  --label datacenter=dc1
```

**开发环境示例：**

```bash
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --label env=development \
  --label region=local \
  --label cluster=dev-cluster \
  --label app=myapp-dev \
  --label team=backend
```

**测试环境示例：**

```bash
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --label env=staging \
  --label region=us-west-2 \
  --label cluster=test-cluster \
  --label app=myapp-test \
  --interval=30s \
  --top=15
```

**边缘节点示例：**

```bash
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --label env=edge \
  --label region=asia \
  --label node_type=edge \
  --label location=beijing \
  --interval=120s \
  --top=5
```

#### 5. 完整配置示例

**生产环境完整配置：**

```bash
./process_exporter \
  --remote.url=https://vm.example.com/api/v1/write \
  --remote.username=monitor \
  --remote.password=secret \
  --interval=30s \
  --top=15 \
  --label env=production \
  --label datacenter=dc1 \
  --label team=platform \
  --label cluster=prod-cluster \
  --label app=monitoring
```

**多租户环境配置：**

```bash
# 租户 A 的服务器
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --label tenant=tenant-a \
  --label env=production \
  --label service=web \
  --interval=60s

# 租户 B 的服务器
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --label tenant=tenant-b \
  --label env=production \
  --label service=api \
  --interval=60s
```

**Kubernetes 节点配置：**

```bash
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --label env=production \
  --label cluster=k8s-prod \
  --label node_role=worker \
  --label zone=us-east-1a \
  --label instance_type=m5.large \
  --interval=30s \
  --top=20
```

#### 6. 后台运行示例

```bash
# 使用 nohup 后台运行
nohup ./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --label env=production \
  --interval=60s \
  > exporter.log 2>&1 &

# 查看进程
ps aux | grep process_exporter

# 查看日志
tail -f exporter.log
```

#### 7. 使用环境变量（通过脚本封装）

创建启动脚本 `start_exporter.sh`：

```bash
#!/bin/bash
export VM_REMOTE_URL="http://victoriametrics:8428/api/v1/write"
export VM_REMOTE_USERNAME="monitor"
export VM_REMOTE_PASSWORD="secret"
export VM_INTERVAL="60s"
export VM_TOP="10"

./process_exporter \
  --remote.url="${VM_REMOTE_URL}" \
  --remote.username="${VM_REMOTE_USERNAME}" \
  --remote.password="${VM_REMOTE_PASSWORD}" \
  --interval="${VM_INTERVAL}" \
  --top="${VM_TOP}" \
  --label env=production \
  --label region=us-east-1
```

#### 8. Docker 容器内运行示例

```bash
# 在 Docker 容器中运行（需要挂载 /proc）
docker run -d \
  --name process-exporter \
  --restart=unless-stopped \
  -v /proc:/host/proc:ro \
  -v $(pwd):/app \
  -w /app \
  alpine:latest \
  ./process_exporter \
    --remote.url=http://victoriametrics:8428/api/v1/write \
    --label env=production \
    --label container=docker
```

## 指标说明

所有指标都会自动包含以下**主机标识标签**：

- `hostname` - 主机名
- `ip` - 本机主要 IP 地址
- `mac` - 主要网卡 MAC 地址

**自定义标签**：

通过 `--label` 参数添加的所有自定义标签也会自动包含在所有指标中。

### 指标列表

#### 进程指标

| 指标名称                  | 类型  | 说明                   | 额外标签                         |
| ------------------------- | ----- | ---------------------- | -------------------------------- |
| `process_memory_bytes`    | Gauge | 进程内存使用量（字节） | `pid`, `name`, `cmdline`, `rank` |
| `process_memory_percent`  | Gauge | 进程内存占用百分比     | `pid`, `name`, `cmdline`, `rank` |
| `process_cpu_percent`     | Gauge | 进程 CPU 使用百分比    | `pid`, `name`, `cmdline`, `rank` |
| `process_runtime_seconds` | Gauge | 进程运行时间（秒）     | `pid`, `name`, `cmdline`         |

#### 系统内存指标

| 指标名称                        | 类型  | 说明                     |
| ------------------------------- | ----- | ------------------------ |
| `system_memory_total_bytes`     | Gauge | 系统总内存（字节）       |
| `system_memory_used_bytes`      | Gauge | 系统已用内存（字节）     |
| `system_memory_available_bytes` | Gauge | 系统可用内存（字节）     |
| `system_memory_used_percent`    | Gauge | 系统内存使用率（百分比） |

### 标签说明

**自动标签**（所有指标自动包含）：

- `hostname` - 主机名（自动采集）
- `ip` - IP 地址（自动采集）
- `mac` - MAC 地址（自动采集）

**进程标签**（特定于每个进程）：

- `pid` - 进程 ID
- `name` - 进程名称
- `cmdline` - 完整命令行
- `rank` - 排名（1-N，仅用于内存和 CPU 指标）

**自定义标签**（通过 `--label` 参数指定）：

- 您可以添加任意数量的自定义标签
- 格式：`--label key=value`
- 常用标签示例：`env`、`region`、`cluster`、`datacenter`、`team`、`app` 等
- `./process_exporter \
--remote.url=http://victoriametrics:8428/api/v1/write \
-interval=120s \
-top=5 \
-label env=edge \
-label region=asia `

### 指标示例

#### 进程指标示例

**不带自定义标签的进程内存指标**：

```
process_memory_bytes{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",pid="1234",name="nginx",cmdline="/usr/sbin/nginx -g daemon off;",rank="1"} 524288000
process_memory_bytes{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",pid="5678",name="mysqld",cmdline="/usr/sbin/mysqld --defaults-file=/etc/mysql/my.cnf",rank="2"} 2147483648
process_memory_percent{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",pid="1234",name="nginx",cmdline="/usr/sbin/nginx -g daemon off;",rank="1"} 12.5
```

**带自定义标签的进程内存指标**：

```
process_memory_bytes{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web",team="backend",pid="1234",name="nginx",cmdline="/usr/sbin/nginx -g daemon off;",rank="1"} 524288000
process_memory_percent{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web",team="backend",pid="1234",name="nginx",cmdline="/usr/sbin/nginx -g daemon off;",rank="1"} 12.5
```

**进程 CPU 指标示例**：

```
process_cpu_percent{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web",pid="5678",name="mysqld",cmdline="/usr/sbin/mysqld --defaults-file=/etc/mysql/my.cnf",rank="1"} 45.2
process_cpu_percent{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web",pid="1234",name="nginx",cmdline="/usr/sbin/nginx -g daemon off;",rank="2"} 8.5
```

**进程运行时间指标示例**：

```
process_runtime_seconds{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",pid="1234",name="nginx",cmdline="/usr/sbin/nginx -g daemon off;"} 86400
process_runtime_seconds{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",pid="5678",name="mysqld",cmdline="/usr/sbin/mysqld --defaults-file=/etc/mysql/my.cnf"} 172800
```

#### 系统内存指标示例

**不带自定义标签的系统内存指标**：

```
system_memory_total_bytes{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e"} 17179869184
system_memory_used_bytes{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e"} 8589934592
system_memory_available_bytes{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e"} 8589934592
system_memory_used_percent{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e"} 50.0
```

**带自定义标签的系统内存指标**：

```
system_memory_total_bytes{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web",team="backend"} 17179869184
system_memory_used_bytes{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web",team="backend"} 8589934592
system_memory_available_bytes{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web",team="backend"} 8589934592
system_memory_used_percent{hostname="server01",ip="192.168.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web",team="backend"} 50.0
```

#### 完整指标输出示例

以下是一次完整的指标推送示例（Top 3 进程）：

```
# 系统内存指标
system_memory_total_bytes{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster"} 34359738368
system_memory_used_bytes{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster"} 25769803776
system_memory_available_bytes{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster"} 8589934592
system_memory_used_percent{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster"} 75.0

# 进程内存指标（Top 3）
process_memory_bytes{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="1234",name="java",cmdline="/usr/bin/java -Xmx8g -jar app.jar",rank="1"} 8589934592
process_memory_bytes{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="5678",name="mysqld",cmdline="/usr/sbin/mysqld --defaults-file=/etc/mysql/my.cnf",rank="2"} 4294967296
process_memory_bytes{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="9012",name="nginx",cmdline="/usr/sbin/nginx -g daemon off;",rank="3"} 1073741824

process_memory_percent{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="1234",name="java",cmdline="/usr/bin/java -Xmx8g -jar app.jar",rank="1"} 25.0
process_memory_percent{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="5678",name="mysqld",cmdline="/usr/sbin/mysqld --defaults-file=/etc/mysql/my.cnf",rank="2"} 12.5
process_memory_percent{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="9012",name="nginx",cmdline="/usr/sbin/nginx -g daemon off;",rank="3"} 3.125

# 进程 CPU 指标（Top 3）
process_cpu_percent{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="1234",name="java",cmdline="/usr/bin/java -Xmx8g -jar app.jar",rank="1"} 65.5
process_cpu_percent{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="5678",name="mysqld",cmdline="/usr/sbin/mysqld --defaults-file=/etc/mysql/my.cnf",rank="2"} 32.8
process_cpu_percent{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="9012",name="nginx",cmdline="/usr/sbin/nginx -g daemon off;",rank="3"} 5.2

# 进程运行时间指标
process_runtime_seconds{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="1234",name="java",cmdline="/usr/bin/java -Xmx8g -jar app.jar"} 259200
process_runtime_seconds{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="5678",name="mysqld",cmdline="/usr/sbin/mysqld --defaults-file=/etc/mysql/my.cnf"} 604800
process_runtime_seconds{hostname="web-server-01",ip="10.0.1.100",mac="00:1a:2b:3c:4d:5e",env="production",region="us-east-1",cluster="web-cluster",pid="9012",name="nginx",cmdline="/usr/sbin/nginx -g daemon off;"} 86400
```

#### 指标值说明

- **内存值**：以字节为单位（1 GB = 1073741824 字节）
- **百分比值**：0-100 之间的浮点数
- **时间值**：以秒为单位
- **rank 标签**：表示进程在内存或 CPU 使用中的排名（1 表示最高）

## 部署

### 使用 systemd

#### 1. 准备可执行文件

```bash
# 创建部署目录
sudo mkdir -p /opt/process-exporter
sudo cp process_exporter_linux_amd64 /opt/process-exporter/process_exporter
sudo chmod +x /opt/process-exporter/process_exporter
```

#### 2. 创建服务文件

复制服务文件：

```bash
sudo cp process-exporter.service /etc/systemd/system/
```

#### 3. 编辑服务文件

编辑服务文件，根据实际环境修改参数：

```bash
sudo vim /etc/systemd/system/process-exporter.service
```

**生产环境配置示例：**

```ini
[Unit]
Description=VmAgent Process Exporter
Documentation=https://github.com/your-org/monitor
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/process-exporter
ExecStart=/opt/process-exporter/process_exporter \
  --remote.url=https://vm.example.com/api/v1/write \
  --remote.username=monitor \
  --remote.password=secret \
  --interval=60s \
  --top=10 \
  --label env=production \
  --label region=us-east-1 \
  --label cluster=prod-cluster \
  --label team=platform

# 自动重启配置
Restart=on-failure
RestartSec=10s
StartLimitInterval=300
StartLimitBurst=5

# 日志配置
StandardOutput=journal
StandardError=journal
SyslogIdentifier=process-exporter

# 安全配置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/process-exporter

[Install]
WantedBy=multi-user.target
```

**开发环境配置示例：**

```ini
[Unit]
Description=VmAgent Process Exporter (Development)
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/process-exporter
ExecStart=/opt/process-exporter/process_exporter \
  --remote.url=http://localhost:8428/api/v1/write \
  --interval=30s \
  --top=15 \
  --label env=development \
  --label region=local

Restart=on-failure
RestartSec=5s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

#### 4. 启动和管理服务

```bash
# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 启用开机自启
sudo systemctl enable process-exporter

# 启动服务
sudo systemctl start process-exporter

# 查看服务状态
sudo systemctl status process-exporter

# 停止服务
sudo systemctl stop process-exporter

# 重启服务
sudo systemctl restart process-exporter

# 重新加载配置（不中断服务）
sudo systemctl reload process-exporter
```

#### 5. 查看和管理日志

```bash
# 实时查看日志
sudo journalctl -u process-exporter -f

# 查看最近 100 行日志
sudo journalctl -u process-exporter -n 100

# 查看今天的日志
sudo journalctl -u process-exporter --since today

# 查看最近 1 小时的日志
sudo journalctl -u process-exporter --since "1 hour ago"

# 查看错误日志
sudo journalctl -u process-exporter -p err

# 导出日志到文件
sudo journalctl -u process-exporter > exporter.log
```

### 使用 Supervisor

#### 1. 准备可执行文件

```bash
# 创建部署目录
sudo mkdir -p /opt/process-exporter
sudo cp process_exporter_linux_amd64 /opt/process-exporter/process_exporter
sudo chmod +x /opt/process-exporter/process_exporter
```

#### 2. 创建配置文件

复制配置文件：

```bash
sudo cp process-exporter-supervisor.ini /etc/supervisor/conf.d/
```

#### 3. 编辑配置文件

编辑配置文件，根据实际环境修改参数：

```bash
sudo vim /etc/supervisor/conf.d/process-exporter-supervisor.ini
```

**生产环境配置示例：**

```ini
[program:process-exporter]
command=/opt/process-exporter/process_exporter --remote.url=https://vm.example.com/api/v1/write --remote.username=monitor --remote.password=secret --interval=60s --top=10 --label env=production --label region=us-east-1 --label cluster=prod-cluster
directory=/opt/process-exporter
user=root
autostart=true
autorestart=true
startretries=3
startsecs=5
stopwaitsecs=10
redirect_stderr=true
stdout_logfile=/var/log/supervisor/process-exporter.log
stdout_logfile_maxbytes=50MB
stdout_logfile_backups=10
stderr_logfile=/var/log/supervisor/process-exporter-error.log
stderr_logfile_maxbytes=50MB
stderr_logfile_backups=10
environment=PATH="/usr/local/bin:/usr/bin:/bin"
```

**开发环境配置示例：**

```ini
[program:process-exporter]
command=/opt/process-exporter/process_exporter --remote.url=http://localhost:8428/api/v1/write --interval=30s --top=15 --label env=development --label region=local
directory=/opt/process-exporter
user=root
autostart=true
autorestart=true
startretries=5
startsecs=3
redirect_stderr=true
stdout_logfile=/var/log/supervisor/process-exporter.log
stdout_logfile_maxbytes=20MB
stdout_logfile_backups=5
environment=PATH="/usr/local/bin:/usr/bin:/bin"
```

#### 4. 启动和管理服务

```bash
# 重新读取配置
sudo supervisorctl reread

# 更新配置并启动
sudo supervisorctl update

# 启动服务
sudo supervisorctl start process-exporter

# 停止服务
sudo supervisorctl stop process-exporter

# 重启服务
sudo supervisorctl restart process-exporter

# 查看服务状态
sudo supervisorctl status process-exporter

# 查看所有服务状态
sudo supervisorctl status

# 重新加载配置（不中断服务）
sudo supervisorctl reread
sudo supervisorctl update process-exporter
```

#### 5. 查看日志

```bash
# 查看标准输出日志
tail -f /var/log/supervisor/process-exporter.log

# 查看错误日志
tail -f /var/log/supervisor/process-exporter-error.log

# 查看最近 100 行日志
tail -n 100 /var/log/supervisor/process-exporter.log

# 搜索错误信息
grep -i error /var/log/supervisor/process-exporter-error.log
```

### 使用 Docker Compose

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  process-exporter:
    image: alpine:latest
    container_name: process-exporter
    restart: unless-stopped
    volumes:
      - /proc:/host/proc:ro
      - ./process_exporter:/app/process_exporter:ro
    working_dir: /app
    command: >
      ./process_exporter
      --remote.url=http://victoriametrics:8428/api/v1/write
      --interval=60s
      --top=10
      --label env=production
      --label region=us-east-1
      --label container=docker
    network_mode: host
    privileged: true
```

启动服务：

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重启服务
docker-compose restart
```

### 批量部署脚本示例

创建批量部署脚本 `deploy.sh`：

```bash
#!/bin/bash

# 配置变量
REMOTE_URL="http://victoriametrics:8428/api/v1/write"
ENV="production"
REGION="us-east-1"
CLUSTER="prod-cluster"
INTERVAL="60s"
TOP="10"

# 部署目录
DEPLOY_DIR="/opt/process-exporter"
SERVICE_FILE="/etc/systemd/system/process-exporter.service"

# 创建目录
sudo mkdir -p $DEPLOY_DIR

# 复制文件
sudo cp process_exporter_linux_amd64 $DEPLOY_DIR/process_exporter
sudo chmod +x $DEPLOY_DIR/process_exporter

# 创建服务文件
sudo tee $SERVICE_FILE > /dev/null <<EOF
[Unit]
Description=VmAgent Process Exporter
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DEPLOY_DIR/process_exporter \\
  --remote.url=$REMOTE_URL \\
  --interval=$INTERVAL \\
  --top=$TOP \\
  --label env=$ENV \\
  --label region=$REGION \\
  --label cluster=$CLUSTER

Restart=on-failure
RestartSec=10s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable process-exporter
sudo systemctl start process-exporter

# 检查状态
sudo systemctl status process-exporter
```

使用脚本：

```bash
chmod +x deploy.sh
./deploy.sh
```

## 查询示例

在 **VictoriaMetrics**、**Prometheus** 或 **Grafana** 中查询指标：

> **提示**：所有查询示例使用 PromQL 语法，同时适用于 VictoriaMetrics 和 Prometheus。

### 基础查询

#### 进程内存查询

```promql
# 查询特定主机的内存使用 Top 5 进程
topk(5, process_memory_bytes{hostname="server01"})

# 查询所有主机的内存使用 Top 10 进程
topk(10, process_memory_bytes)

# 查询特定 IP 主机的所有进程内存使用
process_memory_bytes{ip="192.168.1.100"}

# 查询特定主机的内存使用百分比
process_memory_percent{hostname="server01"}

# 查询特定进程的内存使用（按进程名）
process_memory_bytes{name="nginx"}

# 查询特定进程的内存使用（按命令行匹配）
process_memory_bytes{cmdline=~".*nginx.*"}

# 按主机分组统计进程总内存
sum(process_memory_bytes) by (hostname, ip)

# 按主机分组统计进程平均内存使用率
avg(process_memory_percent) by (hostname)

# 查询内存使用超过 1GB 的进程
process_memory_bytes > 1073741824

# 查询内存使用率超过 10% 的进程
process_memory_percent > 10
```

#### 进程 CPU 查询

```promql
# 查询所有主机的 CPU 使用率最高的进程
topk(10, process_cpu_percent) by (hostname, name)

# 查询特定主机的 CPU 使用 Top 5 进程
topk(5, process_cpu_percent{hostname="server01"})

# 查询特定进程的 CPU 使用率
process_cpu_percent{name="mysqld"}

# 查询 CPU 使用率超过 50% 的进程
process_cpu_percent > 50

# 按主机分组统计进程总 CPU 使用率
sum(process_cpu_percent) by (hostname)

# 按主机分组统计进程平均 CPU 使用率
avg(process_cpu_percent) by (hostname)

# 查询特定主机的所有进程 CPU 使用情况
process_cpu_percent{hostname="server01"}
```

#### 进程运行时间查询

```promql
# 查询所有进程的运行时间
process_runtime_seconds

# 查询特定主机的进程运行时间
process_runtime_seconds{hostname="server01"}

# 查询运行时间最长的 Top 10 进程
topk(10, process_runtime_seconds)

# 查询运行时间超过 24 小时的进程（86400 秒）
process_runtime_seconds > 86400

# 按主机分组统计平均运行时间
avg(process_runtime_seconds) by (hostname)
```

### 使用自定义标签查询

#### 环境维度查询

```promql
# 查询生产环境的所有进程内存使用
process_memory_bytes{env="production"}

# 查询开发环境的所有进程 CPU 使用
process_cpu_percent{env="development"}

# 查询测试环境的系统内存使用率
system_memory_used_percent{env="staging"}

# 按环境统计平均 CPU 使用率
avg(process_cpu_percent) by (env)

# 按环境统计总内存使用
sum(process_memory_bytes) by (env)
```

#### 区域和集群查询

```promql
# 查询特定区域和集群的 CPU 使用情况
process_cpu_percent{region="us-east-1", cluster="web-cluster"}

# 查询特定区域的所有主机系统内存
system_memory_used_percent{region="us-east-1"}

# 按环境和区域分组统计
sum(process_memory_bytes) by (env, region, hostname)

# 查询特定集群的 Top 10 内存使用进程
topk(10, process_memory_bytes{cluster="web-cluster"})

# 按区域统计平均内存使用率
avg(system_memory_used_percent) by (region)

# 查询多个区域的进程内存使用
process_memory_bytes{region=~"us-east-1|us-west-2"}
```

#### 多维度过滤查询

```promql
# 多维度过滤：生产环境 + 特定区域 + 特定应用
topk(10, process_cpu_percent{env="production", region="us-east-1", app="myapp"})

# 查询特定团队管理的服务器
process_memory_bytes{team="platform"}

# 查询特定数据中心的所有指标
process_memory_bytes{datacenter="dc1"}

# 组合多个自定义标签查询
process_cpu_percent{env="production", region="us-east-1", cluster="web", team="backend"}

# 排除特定环境的查询
process_memory_bytes{env!="development"}

# 使用正则表达式匹配标签值
process_memory_bytes{cluster=~"web-.*"}
```

#### 聚合统计查询

```promql
# 按自定义标签聚合 CPU 使用率
avg(process_cpu_percent) by (env, cluster)

# 按自定义标签聚合内存使用
sum(process_memory_bytes) by (env, region, team)

# 按环境统计进程数量
count(process_memory_bytes) by (env)

# 按集群统计最大内存使用
max(process_memory_bytes) by (cluster)

# 按区域统计最小可用内存
min(system_memory_available_bytes) by (region)
```

### 系统内存查询

#### 基础系统内存查询

```promql
# 查询系统总内存
system_memory_total_bytes

# 查询系统总内存（GB）
system_memory_total_bytes / 1024 / 1024 / 1024

# 查询系统已用内存
system_memory_used_bytes

# 查询系统已用内存（GB）
system_memory_used_bytes / 1024 / 1024 / 1024

# 查询系统可用内存
system_memory_available_bytes

# 查询系统可用内存（GB）
system_memory_available_bytes / 1024 / 1024 / 1024

# 查询系统内存使用率
system_memory_used_percent

# 查询特定主机的系统内存
system_memory_total_bytes{hostname="server01"}
```

#### 系统内存告警查询

```promql
# 查询内存使用率超过 80% 的主机
system_memory_used_percent > 80

# 查询内存使用率超过 90% 的主机
system_memory_used_percent > 90

# 查询可用内存少于 1GB 的主机
system_memory_available_bytes < 1073741824

# 查询可用内存少于 500MB 的主机
system_memory_available_bytes < 524288000

# 查询内存使用率超过阈值的生产环境主机
system_memory_used_percent{env="production"} > 85
```

#### 系统内存统计查询

```promql
# 按环境统计平均内存使用率
avg(system_memory_used_percent) by (env)

# 按区域统计平均内存使用率
avg(system_memory_used_percent) by (region)

# 按集群统计平均内存使用率
avg(system_memory_used_percent) by (cluster)

# 查询可用内存最少的 5 台主机
bottomk(5, system_memory_available_bytes)

# 查询可用内存最少的 5 台主机（按主机名）
bottomk(5, system_memory_available_bytes) by (hostname)

# 查询总内存最大的 10 台主机
topk(10, system_memory_total_bytes)

# 按环境统计总内存
sum(system_memory_total_bytes) by (env)

# 按环境统计已用内存
sum(system_memory_used_bytes) by (env)
```

#### 系统内存趋势查询

```promql
# 计算内存使用趋势（5 分钟）
rate(system_memory_used_bytes[5m])

# 计算内存使用变化率（1 小时）
rate(system_memory_used_bytes[1h])

# 计算内存使用增量（5 分钟）
increase(system_memory_used_bytes[5m])

# 计算内存使用率变化（1 小时）
delta(system_memory_used_percent[1h])
```

#### 组合查询

```promql
# 组合查询：系统内存使用率和进程内存总和
system_memory_used_percent and on(hostname) sum(process_memory_bytes) by (hostname)

# 计算进程内存占系统内存的比例
sum(process_memory_bytes) by (hostname) / system_memory_total_bytes * 100

# 查询系统内存使用率和 Top 进程内存使用
system_memory_used_percent{hostname="server01"} and on(hostname) topk(5, process_memory_bytes{hostname="server01"})
```

### Grafana 面板查询示例

#### 面板 1：主机概览

```promql
# 系统内存使用率（百分比）
system_memory_used_percent{hostname="$hostname"}

# 系统可用内存（GB）
system_memory_available_bytes{hostname="$hostname"} / 1024 / 1024 / 1024

# Top 5 进程内存使用（GB）
topk(5, process_memory_bytes{hostname="$hostname"}) / 1024 / 1024 / 1024

# Top 5 进程 CPU 使用率
topk(5, process_cpu_percent{hostname="$hostname"})
```

#### 面板 2：环境统计

```promql
# 按环境统计平均内存使用率
avg(system_memory_used_percent) by (env)

# 按环境统计平均 CPU 使用率
avg(process_cpu_percent) by (env)

# 按环境统计主机数量
count(system_memory_total_bytes) by (env)

# 按环境统计总内存（GB）
sum(system_memory_total_bytes) by (env) / 1024 / 1024 / 1024
```

#### 面板 3：集群监控

```promql
# 按集群统计平均内存使用率
avg(system_memory_used_percent) by (cluster)

# 按集群统计 Top 10 进程内存使用
topk(10, process_memory_bytes{cluster="$cluster"}) / 1024 / 1024 / 1024

# 按集群统计进程数量
count(process_memory_bytes) by (cluster)

# 按集群统计总内存使用
sum(process_memory_bytes) by (cluster) / 1024 / 1024 / 1024
```

### 告警规则示例

#### 告警规则 1：系统内存使用率过高

```yaml
groups:
  - name: process_exporter_alerts
    interval: 30s
    rules:
      - alert: HighMemoryUsage
        expr: system_memory_used_percent > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "主机 {{ $labels.hostname }} 内存使用率过高"
          description: "主机 {{ $labels.hostname }} ({{ $labels.ip }}) 内存使用率为 {{ $value }}%，超过 85% 阈值"
```

#### 告警规则 2：进程内存使用过高

```yaml
      - alert: HighProcessMemoryUsage
        expr: process_memory_percent > 20
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "进程 {{ $labels.name }} 内存使用率过高"
          description: "主机 {{ $labels.hostname }} 上的进程 {{ $labels.name }} (PID: {{ $labels.pid }}) 内存使用率为 {{ $value }}%"
```

#### 告警规则 3：进程 CPU 使用过高

```yaml
      - alert: HighProcessCPUUsage
        expr: process_cpu_percent > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "进程 {{ $labels.name }} CPU 使用率过高"
          description: "主机 {{ $labels.hostname }} 上的进程 {{ $labels.name }} (PID: {{ $labels.pid }}) CPU 使用率为 {{ $value }}%"
```

#### 告警规则 4：系统可用内存不足

```yaml
      - alert: LowAvailableMemory
        expr: system_memory_available_bytes < 1073741824
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "主机 {{ $labels.hostname }} 可用内存不足"
          description: "主机 {{ $labels.hostname }} ({{ $labels.ip }}) 可用内存为 {{ $value | humanize1024 }}，低于 1GB 阈值"
```

#### 告警规则 5：生产环境内存告警

```yaml
      - alert: ProductionHighMemoryUsage
        expr: system_memory_used_percent{env="production"} > 90
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "生产环境主机 {{ $labels.hostname }} 内存使用率严重过高"
          description: "生产环境主机 {{ $labels.hostname }} ({{ $labels.ip }}) 内存使用率为 {{ $value }}%，超过 90% 严重阈值"
```

## 注意事项

1. **必须在 Linux 系统上运行**，因为依赖 `/proc` 文件系统
2. **必须指定 `--remote.url` 参数**，否则程序无法启动
3. **远程端点 URL 格式**：`http(s)://host:port/api/v1/write`
   - **VictoriaMetrics**：默认端口 8428，原生支持 `/api/v1/write` 端点
   - **Prometheus**：默认端口 9090，需要配置 remote_write 接收器或使用 Prometheus Agent 模式
4. **认证支持**：
   - 如果 VictoriaMetrics 或 Prometheus 启用了基本认证，必须提供 `--remote.username` 和 `--remote.password`
   - 支持 HTTP Basic Authentication
5. **采集间隔**：不建议设置过短（建议 ≥ 10s），避免对系统造成负担
6. **主机标识**：主机标识（hostname、IP、MAC）在程序启动时自动采集，运行期间不会改变
7. **兼容性**：
   - 使用 Prometheus 远程写入 API 格式，兼容 VictoriaMetrics 和 Prometheus
   - 指标格式符合 Prometheus 规范，可在任何支持 PromQL 的系统中查询

## 故障排查

### 问题：无法连接到远程端点

#### 检查步骤

1. **验证 URL 格式**

```bash
# 检查 URL 格式是否正确
# 正确格式：http(s)://host:port/api/v1/write
# VictoriaMetrics 示例：http://victoriametrics:8428/api/v1/write
# Prometheus 示例：http://prometheus:9090/api/v1/write
# 错误示例：http://victoriametrics:8428  （缺少 /api/v1/write）
# 错误示例：http://victoriametrics:8428/metrics  （错误的路径）
```

2. **测试网络连接**

```bash
# 测试 HTTP 连接
curl -v http://victoriametrics:8428/api/v1/write

# 测试 HTTPS 连接
curl -v https://vm.example.com/api/v1/write

# 测试带认证的连接
curl -v -u username:password http://victoriametrics:8428/api/v1/write

# 使用 telnet 测试端口
telnet victoriametrics 8428

# 使用 nc 测试端口
nc -zv victoriametrics 8428
```

3. **检查 DNS 解析**

```bash
# 检查主机名解析
nslookup victoriametrics
dig victoriametrics
host victoriametrics

# 检查 /etc/hosts 文件
cat /etc/hosts | grep victoriametrics
```

4. **检查防火墙规则**

```bash
# 检查 iptables 规则
sudo iptables -L -n | grep 8428

# 检查 firewalld 规则
sudo firewall-cmd --list-all

# 临时关闭防火墙测试（仅用于测试）
sudo systemctl stop firewalld
```

5. **检查服务状态**

**检查 VictoriaMetrics：**

```bash
# 检查 VictoriaMetrics 是否运行
curl http://victoriametrics:8428/health

# 检查 VictoriaMetrics 版本
curl http://victoriametrics:8428/version

# 检查 VictoriaMetrics 指标
curl http://victoriametrics:8428/metrics
```

**检查 Prometheus：**

```bash
# 检查 Prometheus 是否运行
curl http://prometheus:9090/-/healthy

# 检查 Prometheus 版本
curl http://prometheus:9090/api/v1/status/buildinfo

# 检查 Prometheus 配置
curl http://prometheus:9090/api/v1/status/config

# 检查 Prometheus 指标
curl http://prometheus:9090/metrics
```

#### 常见错误和解决方案

```bash
# 错误：dial tcp: lookup victoriametrics: no such host
# 解决：检查 DNS 配置或使用 IP 地址

# 错误：connection refused
# 解决：检查 VictoriaMetrics 是否运行，端口是否正确

# 错误：timeout
# 解决：检查网络连接和防火墙规则
```

### 问题：认证失败

#### 检查步骤

1. **验证用户名和密码**

```bash
# 使用 curl 测试认证
curl -u username:password http://victoriametrics:8428/api/v1/write

# 如果失败，检查返回的错误信息
curl -v -u username:password http://victoriametrics:8428/api/v1/write
```

2. **检查认证配置**

**检查 VictoriaMetrics 认证：**

```bash
# 检查 VictoriaMetrics 配置文件
cat /etc/victoriametrics/vmagent.yml | grep -i auth

# 检查 VictoriaMetrics 是否启用了认证
curl http://victoriametrics:8428/api/v1/write
# 如果返回 401 Unauthorized，说明启用了认证
```

**检查 Prometheus 认证：**

```bash
# 检查 Prometheus 配置文件
cat /etc/prometheus/prometheus.yml | grep -i auth

# 检查 Prometheus 是否启用了认证
curl http://prometheus:9090/api/v1/write
# 如果返回 401 Unauthorized，说明启用了认证
```

3. **检查密码中的特殊字符**

```bash
# 如果密码包含特殊字符，可能需要转义或使用引号
# 例如：密码是 P@ssw0rd，在命令行中使用引号
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --remote.password='P@ssw0rd'
```

#### 常见错误和解决方案

```bash
# 错误：401 Unauthorized
# 解决：检查用户名和密码是否正确

# 错误：403 Forbidden
# 解决：检查用户权限配置

# 错误：密码包含特殊字符导致解析错误
# 解决：使用引号包裹密码，或使用环境变量
```

### 问题：没有数据推送

#### 检查步骤

1. **检查程序是否运行**

```bash
# 检查进程
ps aux | grep process_exporter

# 检查 systemd 服务状态
sudo systemctl status process-exporter

# 检查 supervisor 服务状态
sudo supervisorctl status process-exporter
```

2. **查看程序日志**

```bash
# systemd 日志
sudo journalctl -u process-exporter -n 100
sudo journalctl -u process-exporter -f

# supervisor 日志
tail -f /var/log/supervisor/process-exporter.log
tail -f /var/log/supervisor/process-exporter-error.log

# 如果直接运行，查看标准输出
./process_exporter --remote.url=... 2>&1 | tee exporter.log
```

3. **检查指标是否推送成功**

**检查 VictoriaMetrics：**

```bash
# 查询进程指标
curl 'http://victoriametrics:8428/api/v1/query?query=process_memory_bytes'

# 查询系统内存指标
curl 'http://victoriametrics:8428/api/v1/query?query=system_memory_total_bytes'

# 查询特定主机的指标
curl 'http://victoriametrics:8428/api/v1/query?query=process_memory_bytes{hostname="server01"}'

# 使用 VictoriaMetrics UI 查询
# 访问 http://victoriametrics:8428/vmui
```

**检查 Prometheus：**

```bash
# 查询进程指标
curl 'http://prometheus:9090/api/v1/query?query=process_memory_bytes'

# 查询系统内存指标
curl 'http://prometheus:9090/api/v1/query?query=system_memory_total_bytes'

# 查询特定主机的指标
curl 'http://prometheus:9090/api/v1/query?query=process_memory_bytes{hostname="server01"}'

# 使用 Prometheus UI 查询
# 访问 http://prometheus:9090/graph
```

4. **检查程序配置**

```bash
# 检查 systemd 服务配置
sudo systemctl cat process-exporter

# 检查 supervisor 配置
sudo cat /etc/supervisor/conf.d/process-exporter-supervisor.ini

# 验证参数是否正确
./process_exporter --help
```

5. **手动测试推送**

```bash
# 手动推送测试数据
echo 'test_metric{hostname="test"} 1' | \
  curl -X POST \
  --data-binary @- \
  http://victoriametrics:8428/api/v1/write

# 如果成功，应该返回 204 No Content
```

#### 常见错误和解决方案

```bash
# 错误：程序启动后立即退出
# 解决：检查日志中的错误信息，通常是配置错误

# 错误：程序运行但没有数据
# 解决：检查 /proc 文件系统是否可访问，检查权限

# 错误：数据推送失败但程序继续运行
# 解决：检查网络连接和 VictoriaMetrics/Prometheus 服务状态
```

### 问题：程序无法读取 /proc 文件系统

#### 检查步骤

```bash
# 检查 /proc 是否存在
ls -la /proc

# 检查 /proc 权限
ls -ld /proc

# 检查是否可以读取进程信息
cat /proc/1/stat
cat /proc/meminfo

# 检查程序运行用户权限
ps aux | grep process_exporter
```

#### 解决方案

```bash
# 如果使用 Docker，确保挂载 /proc
docker run -v /proc:/host/proc:ro ...

# 如果权限不足，使用 root 用户运行
sudo ./process_exporter --remote.url=...

# 或者检查 systemd/supervisor 配置中的 User 设置
```

### 问题：指标数据不准确

#### 检查步骤

1. **验证采集间隔**

```bash
# 检查配置的采集间隔
sudo systemctl cat process-exporter | grep interval

# 验证间隔设置是否合理（建议 ≥ 10s）
```

2. **对比系统命令**

```bash
# 对比 top 命令输出
top -b -n 1 | head -20

# 对比 ps 命令输出
ps aux --sort=-%mem | head -10
ps aux --sort=-%cpu | head -10

# 对比 free 命令输出
free -h
cat /proc/meminfo
```

3. **检查 VictoriaMetrics 数据**

```bash
# 查询最近的指标值
curl 'http://victoriametrics:8428/api/v1/query?query=process_memory_bytes[5m]'

# 检查指标时间戳
curl 'http://victoriametrics:8428/api/v1/query?query=process_memory_bytes' | jq
```

### 问题：程序占用资源过高

#### 检查步骤

```bash
# 检查程序 CPU 和内存使用
top -p $(pgrep process_exporter)

# 检查系统资源
htop
iostat -x 1
```

#### 解决方案

```bash
# 增加采集间隔
--interval=120s

# 减少监控的进程数量
--top=5

# 检查是否有其他程序干扰
ps aux | sort -k3 -rn | head -10
```

### 问题：systemd 服务无法启动

#### 检查步骤

```bash
# 查看详细错误信息
sudo systemctl status process-exporter -l

# 查看启动日志
sudo journalctl -u process-exporter -n 50

# 检查服务文件语法
sudo systemd-analyze verify /etc/systemd/system/process-exporter.service

# 检查可执行文件路径
ls -la /opt/process-exporter/process_exporter

# 手动测试命令
sudo -u root /opt/process-exporter/process_exporter --remote.url=...
```

#### 常见错误

```bash
# 错误：ExecStart 路径不存在
# 解决：检查可执行文件路径是否正确

# 错误：权限不足
# 解决：检查文件权限和 User 配置

# 错误：WorkingDirectory 不存在
# 解决：创建目录或修改配置
```

### 问题：supervisor 服务无法启动

#### 检查步骤

```bash
# 查看详细状态
sudo supervisorctl status process-exporter

# 查看日志
sudo tail -f /var/log/supervisor/process-exporter.log
sudo tail -f /var/log/supervisor/process-exporter-error.log

# 检查配置文件语法
sudo supervisorctl reread

# 手动测试命令
sudo -u root /opt/process-exporter/process_exporter --remote.url=...
```

### 调试技巧

#### 启用详细日志

```bash
# 直接运行并查看输出
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --interval=10s \
  2>&1 | tee debug.log

# 使用 strace 跟踪系统调用
strace -f -e trace=network,file ./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write
```

#### 测试网络连接

```bash
# 使用 tcpdump 抓包（VictoriaMetrics 默认端口 8428）
sudo tcpdump -i any -n port 8428

# 使用 tcpdump 抓包（Prometheus 默认端口 9090）
sudo tcpdump -i any -n port 9090

# 使用 wireshark 分析
sudo tshark -i any -f "port 8428 or port 9090"
```

#### 验证配置

```bash
# 创建测试配置文件
cat > test_config.sh <<EOF
#!/bin/bash
./process_exporter \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --interval=10s \
  --top=5 \
  --label test=true
EOF

chmod +x test_config.sh
./test_config.sh
```

## VictoriaMetrics 和 Prometheus 兼容性

### 支持的监控系统

本工具同时支持 **VictoriaMetrics** 和 **Prometheus**，使用标准的 Prometheus 远程写入 API（`/api/v1/write`）推送指标。

### VictoriaMetrics 配置

VictoriaMetrics **原生支持** `/api/v1/write` 端点，无需额外配置：

```bash
# VictoriaMetrics 默认端口：8428
./process_exporter --remote.url=http://victoriametrics:8428/api/v1/write
```

**VictoriaMetrics 优势：**
- ✅ 原生支持远程写入 API，开箱即用
- ✅ 高性能，适合大规模数据采集
- ✅ 支持 PromQL 查询语法
- ✅ 兼容 Prometheus 指标格式

### Prometheus 配置

Prometheus 需要配置 remote_write 接收器。有两种方式：

#### 方式 1：使用 Prometheus Agent 模式（推荐）

Prometheus Agent 模式专门用于接收远程写入数据：

```yaml
# prometheus-agent.yml
global:
  external_labels:
    cluster: 'production'

# 启用远程写入接收器
remote_write:
  - url: "http://prometheus:9090/api/v1/write"
    basic_auth:
      username: "monitor"
      password: "secret"
```

启动 Prometheus Agent：

```bash
prometheus --config.file=prometheus-agent.yml --enable-feature=agent
```

#### 方式 2：使用 Prometheus 远程写入端点

如果使用标准 Prometheus，需要配置 remote_write 接收器或使用支持远程写入的 Prometheus 发行版（如 Prometheus Operator）。

**Prometheus 配置示例：**

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

# 远程写入配置（如果需要转发到其他系统）
remote_write:
  - url: "http://victoriametrics:8428/api/v1/write"

# 注意：标准 Prometheus 不直接支持接收远程写入
# 需要使用 Prometheus Agent 模式或第三方接收器
```

### 端点 URL 格式

两种系统使用相同的端点格式：

| 系统 | 默认端口 | 端点 URL 示例 |
|------|---------|-------------|
| VictoriaMetrics | 8428 | `http://victoriametrics:8428/api/v1/write` |
| Prometheus Agent | 9090 | `http://prometheus:9090/api/v1/write` |

### 认证支持

两种系统都支持 HTTP Basic Authentication：

```bash
# VictoriaMetrics 带认证
./process_exporter \
  --remote.url=https://vm.example.com/api/v1/write \
  --remote.username=monitor \
  --remote.password=secret

# Prometheus 带认证
./process_exporter \
  --remote.url=https://prometheus.example.com/api/v1/write \
  --remote.username=monitor \
  --remote.password=secret
```

### 查询兼容性

所有查询使用 **PromQL** 语法，在两种系统中都可以使用：

```promql
# 这些查询在 VictoriaMetrics 和 Prometheus 中都可以使用
process_memory_bytes{hostname="server01"}
topk(10, process_cpu_percent)
system_memory_used_percent > 80
```

### 选择建议

**选择 VictoriaMetrics 如果：**
- 需要高性能和大规模数据采集
- 需要原生支持远程写入
- 需要更好的压缩和存储效率
- 需要更快的查询性能

**选择 Prometheus 如果：**
- 已有 Prometheus 生态系统
- 需要使用 Prometheus Agent 模式
- 需要与现有 Prometheus 工具链集成
- 团队熟悉 Prometheus 运维

### 迁移指南

从 VictoriaMetrics 迁移到 Prometheus（或反之）非常简单，只需更改 `--remote.url` 参数：

```bash
# 从 VictoriaMetrics 切换到 Prometheus
# 只需修改 URL，其他配置保持不变
./process_exporter \
  --remote.url=http://prometheus:9090/api/v1/write \
  --interval=60s \
  --top=10 \
  --label env=production
```

