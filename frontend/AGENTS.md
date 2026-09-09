# 前端子目錄

規則與導覽在 repo 根目錄:[../AGENTS.md](../AGENTS.md)、[../CONTRIBUTING.md](../CONTRIBUTING.md)。
dev container 的 workspace 雖是本目錄,repo 根一律掛載於 `/app`,上述檔案都讀得到。

本目錄最容易踩的兩件事:`src/api/gen/` 不入版控(缺了跑 `pnpm gen:api-client`,
需 submodule 已 checkout);API **規格**變更(端點、輸入輸出、模型)一律從 `../typespec/` 改起,純實作變更
不必;不要手改 `src/api/openapi.yaml`。
