# HupuApp · 个人开发副本

本仓库是 [wudizhangzhi/HupuApp](https://github.com/wudizhangzhi/HupuApp) 的独立私有开发副本，由 [Alex-Lord](https://github.com/Alex-Lord) 用于个人使用与后续开发。

**原项目由 wudizhangzhi 及其贡献者开发。** 本副本的初始程序代码直接继承上游，不是从零重写的项目。保留完整 Git 历史、原作者署名和 [MIT LICENSE](LICENSE)，并保留原作者对 [chenjiandongx/HupuLive](https://github.com/chenjiandongx/HupuLive) 的致谢。此副本不代表原作者，也不代表虎扑官方。

为满足私有开发需求，此仓库不属于 GitHub 的公开 Fork 网络，但保留上游代码及 Git 历史。私有仓库仅所有者和受邀协作者可访问，仅持有链接无法访问。

## 本副本的维护范围

- 起点：上游提交 `735288e988356305de690502c93fc8607718fcaf`。
- 本副本新增统一功能菜单、持久化网页登录、主队选择与新闻筛选，并改进返回导航和直播错误处理；这些改动由本副本维护，不属于上游原始功能。
- 后续改动通过独立提交记录；适合上游的通用修复，会在验证后考虑向上游提交 PR。
- 下方保留原作者 README 原文。原文中的开发经历、测试环境和待办属于上游，不表示本副本已重新验证或完成。

## 统一入口与自动登录

直接运行 `hupu`（或构建后的 `./bin/hupu`），自动验证本地 Cookie。有效时不打开浏览器；没有 Cookie 或确认失效时，打开持久化的 Chrome 登录窗口，尝试恢复会话，需要时由你在虎扑官网完成手机号/扫码验证。程序自动获取并验证 Cookie、保存、关闭窗口，然后显示功能菜单。网络错误不会清掉已有会话或强制重新登录。

- Live：NBA 文字直播，选择比赛后进入直播间。
- NBA 新闻：上游 NBA 新闻列表。
- 主队新闻：从 NBA 资讯标题筛选明确提及主队的“流言板”新闻，每页扫描 3 个源页面并去重；不会包含只提及球员、未写队名的新闻，也不保证覆盖所有主队新闻。
- 主队专区：该球队专区的新闻与讨论，不过滤讨论帖。
- 设置主队：支持 30 支 NBA 球队，兼容旧 CLI 的主队偏好；未设置时默认 76 人。

方向键选择、Enter 进入，选中项仅加粗。新闻正文中 Esc 返回列表、列表中 Esc 返回菜单；Ctrl+Q 切换面板，Ctrl+N / Ctrl+P 翻页。直播中 Ctrl+C 返回菜单，主菜单中 Esc / Ctrl+C 退出。

```sh
hupu                         # 自动登录后进入功能菜单
hupu --guest                 # 跳过登录体验公开内容
hupu --login-wait 10m         # 延长网页登录等待时间
hupu login                   # 复用/恢复登录后进入菜单
hupu auth                    # 仅检查会话
hupu news nba                # 直接看 NBA 新闻
hupu news news:sixers         # 直接看 76 人相关新闻
hupu news sixers              # 直接看 76 人专区
hupu live nba                # 直接进入直播入口
```

Cookie 与浏览器配置沿用系统用户配置目录下的 `hupu/session.cookie`、`hupu/browser-profile`，主队为 `hupu/preferred-team`。macOS 位于 `~/Library/Application Support/hupu/`。Cookie 文件权限为 0600，专用浏览器目录权限为 0700，不上传 GitHub；BBS Cookie 不发送给直播 API 或站外跳转。自动登录需要已安装 Chrome / Chromium。

## 从本副本构建

需要 Go 1.24 或更新版本。

```sh
git clone https://github.com/Alex-Lord/HupuApp-dev.git
cd HupuApp-dev
git remote add upstream https://github.com/wudizhangzhi/HupuApp.git
go build -o bin/hupu ./cmd
./bin/hupu news nba
```

已在 macOS Apple Silicon、Go 1.27.3 上构建并手动验证 NBA 新闻列表、正文和评论；已验证 NBA 赛程接口；实时比赛持续推送与其他平台尚未完整验证。需在支持光标控制的交互终端运行。

本副本尚未发布自己的二进制安装包。下方原文中的 Releases 链接属于上游，其版本不包含本副本后续改动。

离线验证：`go test ./cmd ./gohupu/...`、`go vet ./cmd ./gohupu/...`。上游 `test/` 含交互及联网测试，未包含在离线检查中。

开发分支、上游同步和贡献约定见 [CONTRIBUTING.md](CONTRIBUTING.md)。

---

## 原作者 README（保留原文）

# 虎扑篮球命令行版  

![Duncan](images/Duncan.jpg)

灵感来自另一位[JR](https://github.com/chenjiandongx/HupuLive), 本来想fork他的工程继续改的,但是发现自己都没用到他的代码。所以就自己新建一个项目了.

~~目前大部分功能支持python3.x， python2.x, 部分功能python2.x的适配还在加油中...~~

~~window的适配也在努力中...~~

# go 版本的架构

- cmd 入口
- gohupu
  - api
    - live.go 直播接口
    - base.go 接口基础
  - message
    - ~~ws.go   websocket消息~~
    - http.go http消息
  - live
    - client.go 直播客户端
  - menu
    - base.go   基础
    - live.go   直播菜单
  - logger  日志
- constant.go  常量，设置等
- utils.go     公用方法

### 测试环境

- go version go1.16.4 windows/amd64

## 安装

在[release](https://github.com/wudizhangzhi/HupuApp/releases)页面中选择符合自己系统版本的下载，
解压缩后进入文件夹直接执行`./hupu`

## 使用指南

```bash
Usage: Hupu <command>

A command line tool for Hupu.

Commands:
  live            比赛直播.
    <gameType>    比赛类型(nba/cba).

  news            新闻、帖子.
    <region>      领域(nba/cba/湿乎乎/步行街).

Flags:
  -h, --help     Show context-sensitive help.
      --debug    Enable debug mode.

Run "Hupu <command> --help" for more information on a command.
```

方向键上下控制选项，回车选择，ctrl+c中途退出。

## 虎扑直播

```bash
Hupu live nba
```

![hupu_live](images/hupu_live.gif)

![hupu_live_contrast](images/hupu_live_contrast.gif)

## 虎扑新闻

```bash
Hupu news 步行街
```

![hupu_news](images/news_v2.gif)

## To-do list

- [ ] go版本其他功能(新闻，数据等)
  - [x] 新闻
  - [ ] 数据
