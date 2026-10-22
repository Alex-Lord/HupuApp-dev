# 开发与贡献

## 来源与许可

本项目派生自 [wudizhangzhi/HupuApp](https://github.com/wudizhangzhi/HupuApp)。保留上游 Git 历史、原有作者署名、原始 MIT LICENSE 和相关致谢。发布源码或二进制时，应随附 MIT 许可证及其中的版权声明。新增说明应清楚区分上游功能与本副本 的改动，不能将继承的代码描述为个人原创。

## 本地开发

- `origin` 指向 `Alex-Lord/HupuApp`，用于推送个人开发成果。
- `upstream` 指向 `wudizhangzhi/HupuApp`，用于读取和同步上游。
- 默认分支沿用上游的 `master`；每项改动在独立分支开发，通过 PR 合入自己的私有仓库。

```sh
git switch master
git pull --ff-only origin master
git switch -c feature/describe-the-change
go build -o bin/hupu ./cmd
./bin/hupu news nba
```

提交前检查 `git diff`，只暂存本次需要的文件。不要提交本地可执行文件、日志、Cookie、浏览器配置、访问令牌或个人账号数据。本项目沿用上游 Go module 路径；不要为了个人副本 的名称而批量改写归属信息。

## 验证改动

先运行构建，再针对改动验证功能。`test/` 中存在调用真实接口或交互菜单的测试，不能将未经检查的 `go test ./...` 当作完全离线的单元测试。增加新测试时优先使用固定样例或本地 HTTP 服务，并在 PR 中说明自动化测试与手动测试各自覆盖的范围。

## 同步上游

```sh
git fetch upstream
git switch -c sync/upstream-master origin/master
git merge upstream/master
```

解决冲突、验证后，将同步分支推送到自己的私有仓库，并通过 PR 合入 `master`。保留上游提交与作者信息，不使用重新初始化 Git、复制代码后重新署名或强制推送的方式掩盖来源。

## 向上游贡献

通用修复应保持范围清晰，提供复现方式、改动理由和验证结果，遵循上游已有约定。先在自己的私有仓库 评审和验证；准备好后，在获准公开相关改动的前提下，使用另建的公开 Fork 向上游发起 PR；私有仓库不直接向公开上游发起 PR。个人偏好或尚未成熟的实验保留在本副本，不将原作者列为维护或支持这些改动的责任人。
