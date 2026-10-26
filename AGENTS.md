# 项目约定

- 遵守 CONTRIBUTING.md，保留原作者署名、MIT LICENSE 和上游历史。
- 本仓库为用户个人开发的私有仓库。用户已授权：完成修改并验证通过后，默认合并、推送到 origin/master，无需重复确认，不要仅停留在开发分支或草稿 PR。不要绕过失败的检查或强制推送。
- Cookie、浏览器配置、令牌、日志和编译产物不得提交。
- 验证使用 go test ./cmd ./gohupu/...、go vet ./cmd ./gohupu/... 和 go build -o bin/hupu ./cmd；上游 test/ 包含交互或联网测试。
