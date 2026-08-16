# 项目基准自检（评测用）

镜像：golang:1.22-bookworm
构建脚本：build_benzhi_docker.sh（用法见脚本头注释）

## 构建与自检（三件套验收）
1. docker build -f benzhi.Dockerfile -t benzhi/patrol-log .    预期：成功，无 downloading
2. docker run --rm -it benzhi/patrol-log sh
   - go version           预期：工具链完整
   - go build ./...       预期：成功
   - go test ./...        预期：成功，可离线测试
3. 双架构：分别以 linux/amd64 与 linux/arm64 构建并运行各一次
   ./build_benzhi_docker.sh patrol-log linux/amd64
   ./build_benzhi_docker.sh patrol-log linux/arm64
   （经 QEMU 模拟的架构不做 -race，TSan 会报 VMA range；用 go test/go vet/go build 验证，
    -race 只在原生架构跑）

## 项目说明
设备巡检记录服务：运维人员按巡检计划对设备逐项检查，记录读数与状态（正常/异常），
异常自动生成告警；支持按设备查询历史与导出最近结果。
数据存 SQLite（patrol.db，纯 Go 驱动 glebarez/sqlite，无 CGO）。
启动方式：go run . 或构建镜像后 docker run，监听 :8080。

## 对本仓库的修改约定
- 具备完整 Go 工具链；改代码/编译/跑测试均须在容器内进行
- 只改进公开行为，不断言私有实现
