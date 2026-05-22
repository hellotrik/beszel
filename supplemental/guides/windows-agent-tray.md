# Windows 托盘版 Agent

## 程序图标

项目根目录 [`icon.ico`](../../icon.ico) 为 Windows Agent 图标源文件：

- **托盘图标**：构建时复制到 `internal/cmd/agent/assets/icon.ico` 并 `go:embed`
- **exe 图标**（资源管理器里看到的）：构建时由 `rsrc` 生成 `rsrc_windows_*.syso`

更新图标后执行：

```bash
make sync-agent-icon ARCH=amd64
# 或完整 Windows 构建
make build-agent OS=windows ARCH=amd64
```

## 两种运行方式

| 模式 | 启动方式 | 说明 |
|------|----------|------|
| **托盘** | 双击 `beszel-agent.exe` 或开始菜单快捷方式 | 任务栏托盘图标，可编辑配置 |
| **服务** | NSSM / `beszel-agent.exe -service` | 无托盘（Session 0），适合后台常驻 |

安装脚本 `install-agent.ps1` 注册的服务默认带 **`-service`** 参数。

## 配置文件

路径：`%APPDATA%\beszel-agent\config.json`

```json
{
  "hub_url": "https://your-hub.example",
  "token": "your-token",
  "key": "ssh-ed25519 AAAA...",
  "port": "45876",
  "listen": "",
  "log_level": ""
}
```

- 托盘：**优先**使用 `config.json`（可通过菜单「Settings…」或「Open config.json」编辑）。
- 服务：可使用 `config.json`（若完整）或 NSSM 环境变量 `HUB_URL` / `TOKEN` / `KEY`。

## 托盘菜单

- **启动 / 停止** — 启动或停止与 Hub 的连接
- **设置…** — 图形界面配置 Hub 地址、令牌、SSH 公钥等（首次运行会自动弹出）
- **打开 config.json** — 用记事本编辑（与「设置…」写入同一文件）
- **重新加载配置** — 重新读取 `config.json` 并重启 Agent
- **打开日志目录** — `%ProgramData%\beszel-agent\logs`

「设置…」在独立 UI 线程中打开原生对话框；若窗口未出现，请重新编译最新版 `beszel-agent.exe` 后再试。

## 命令行（与 Linux 相同）

```powershell
beszel-agent.exe health
beszel-agent.exe fingerprint view
beszel-agent.exe -service   # 无界面，供 NSSM 使用
```

## 桌面自启动（托盘）

将快捷方式放到：

`%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup`

目标：`beszel-agent.exe`（不要加 `-service`）。
