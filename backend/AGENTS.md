# 後端子目錄

規則與導覽在 repo 根目錄:[../AGENTS.md](../AGENTS.md)、[../CONTRIBUTING.md](../CONTRIBUTING.md)。
dev container 的 workspace 雖是本目錄,repo 根一律掛載於 `/app`,上述檔案都讀得到。

本目錄最容易踩的兩件事:`api/api.gen.go` 與 `internal/**` 各層的 `mocks/` 不入版控
(缺了跑 `make api` / `make mocks`);API **規格**變更(端點、輸入輸出、模型)一律從 `../typespec/` 改起,純實作變更不必。
