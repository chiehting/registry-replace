# registry-replace

## 專案概覽

  registry-replace 是一個專為解決受限的網路環境下容器映像檢索問題而設計的工具。它能夠自動將 docker.io 等受限來源切換至指定的鏡像來源，大幅提升映像檢索效率和成功率。

## 主要特點

- 自動檢測並切換至最佳鏡像來源
- 無縫集成至 Kubernetes 環境
- 高度可配置，適應不同需求
- 提升容器部署效率，減少網路問題造成的延遲

## 快速開始

### 前置條件

- Kubernetes 集群 (版本 1.16+)
- kubectl 命令行工具
- 管理員權限

### 安裝

  1. 克隆儲存庫：

     ```bash
     git clone https://github.com/chiehting/registry-replace.git
     cd registry-replace
     ```

  2. 部署 registry-replace：

     ```bash
     make apply
     ```

### 配置

  1. 檢查並根據需要修改配置：

     ```bash
     kubectl describe configmap -n admission registry-replace
     ```

  2. 編輯配置以指定首選的鏡像來源。

### 使用

  部署完成後，registryReplace 將自動攔截並處理 Kubernetes 的映像拉取請求，無需額外操作。

### 卸載

  如需移除 registryReplace：

  ```bash
  make delete
  ```

<!-- ## 進階配置

  有關高級設置和自定義選項，請參閱我們的 [配置指南](docs/configuration.md)。 -->

<!-- ## 故障排除

  常見問題和解決方案可在 [故障排除指南](docs/troubleshooting.md) 中找到。 -->

<!-- ## 貢獻

  我們歡迎社區貢獻！請查看 [貢獻指南](CONTRIBUTING.md) 了解如何參與項目開發。 -->

## 授權

  本專案採用 [MIT 授權](LICENSE)。

## 聯繫我們

  如有任何問題或建議，請 [開啟一個 issue](https://github.com/chiehting/registry-replace/issues) 或發送郵件至 <ting911111@gmail.com>。
