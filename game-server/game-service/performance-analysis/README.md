# 🎮 Game Service Performance Analysis

這個資料夾包含所有效能分析相關的工具、腳本和報告。

## 📁 資料夾結構

```
performance-analysis/
├── README.md           # 本文檔（使用說明）
├── scripts/            # 分析腳本
│   ├── 1_escape_analysis.sh      # 逃逸分析
│   ├── 2_cpu_profile.sh          # CPU profiling
│   ├── 3_memory_profile.sh       # Memory profiling
│   ├── 4_benchmark.sh            # Benchmark 測試
│   └── quick_escape_analysis.sh  # 快速逃逸分析報告
├── reports/            # 分析報告輸出
│   ├── escape_analysis.txt       # 完整逃逸分析
│   ├── session_escapes.txt       # Session 逃逸
│   ├── movement_escapes.txt      # Movement 逃逸
│   └── critical_moved_to_heap.txt # 嚴重逃逸
└── profiles/           # Profiling 數據
    ├── cpu.prof
    ├── mem.prof
    └── goroutine.prof
```

## 🚀 快速開始

### 1. 逃逸分析（Escape Analysis）

檢查哪些變數逃逸到 heap，影響 GC 效能：

```bash
# 執行完整逃逸分析
./scripts/1_escape_analysis.sh

# 查看快速報告
./scripts/quick_escape_analysis.sh
```

**查看報告：**
```bash
cat reports/session_escapes.txt        # Session.go 的逃逸
cat reports/movement_escapes.txt       # Movement system 的逃逸
cat reports/critical_moved_to_heap.txt # 最嚴重的逃逸
```

### 2. CPU Profiling

分析 CPU 熱點：

```bash
# 啟動 game-service
go run ./cmd/server

# 在另一個 terminal 執行
./scripts/2_cpu_profile.sh

# 會自動打開互動式分析工具
```

### 3. Memory Profiling

分析記憶體分配：

```bash
# 啟動 game-service
go run ./cmd/server

# 在另一個 terminal 執行
./scripts/3_memory_profile.sh
```

### 4. Benchmark

執行效能基準測試：

```bash
./scripts/4_benchmark.sh
```

---

## 📊 各類分析說明

### 逃逸分析（Escape Analysis）

**目的：** 找出哪些變數從 stack 逃逸到 heap，造成 GC 壓力

**關鍵指標：**
- `escapes to heap` - 變數逃逸到 heap
- `moved to heap` - 變數被強制移到 heap（更嚴重）
- `leaking param` - 參數逃逸

**重點關注：**
1. `manageGameLoop()` - 每秒執行 30 次的熱路徑
2. `broadcastFullState()` - 狀態序列化和廣播
3. `GetAllEntities()` - Entity 管理
4. `metrics.TickDuration.Record()` - Metrics 記錄

**優化目標：**
- 減少 hot path 中的 heap allocation
- 盡量讓 Game Loop 中的變數保持在 stack

---

### CPU Profiling

**目的：** 找出 CPU 熱點，哪些函數佔用最多 CPU 時間

**常用命令：**
```bash
# 在 pprof 互動模式中
(pprof) top 20              # 顯示前 20 個 CPU 佔用高的函數
(pprof) list manageGameLoop # 查看 manageGameLoop 的詳細 profile
(pprof) web                 # 生成圖形化視圖（需要 graphviz）
(pprof) pdf > cpu.pdf       # 匯出 PDF 報告
```

**重點關注：**
- `flat%` - 函數本身的 CPU 佔用
- `cum%` - 函數及其調用的所有函數的總 CPU 佔用

---

### Memory Profiling

**目的：** 找出記憶體分配熱點

**兩種模式：**
1. **alloc_space** - 總分配量（預設）
2. **inuse_space** - 當前使用量

**常用命令：**
```bash
(pprof) top 20 -alloc_space    # 按總分配排序
(pprof) top 20 -inuse_space    # 按當前使用排序
(pprof) list broadcastFullState # 查看函數的詳細分配
```

---

## 🎯 效能優化 Workflow

### Step 1: Baseline 測量

```bash
# 1. 執行逃逸分析，記錄當前狀態
./scripts/1_escape_analysis.sh > reports/baseline_escape.txt

# 2. 執行 benchmark，記錄基準效能
./scripts/4_benchmark.sh > reports/baseline_bench.txt

# 3. 記錄當前的 Grafana P95 數值
```

### Step 2: 識別瓶頸

```bash
# 1. 查看逃逸報告，找出 hot path 的逃逸
cat reports/session_escapes.txt

# 2. CPU profiling，找出 CPU 熱點
./scripts/2_cpu_profile.sh

# 3. Memory profiling，找出記憶體分配熱點
./scripts/3_memory_profile.sh
```

### Step 3: 優化

根據分析結果進行優化（例如：）
- 減少不必要的 heap allocation
- 使用 sync.Pool 復用對象
- 優化序列化邏輯
- 批次處理

### Step 4: 驗證

```bash
# 1. 重新執行逃逸分析
./scripts/1_escape_analysis.sh > reports/optimized_escape.txt

# 2. 重新執行 benchmark
./scripts/4_benchmark.sh > reports/optimized_bench.txt

# 3. 比較結果
diff reports/baseline_bench.txt reports/optimized_bench.txt

# 4. 檢查 Grafana P95 是否改善
```

---

## 📈 Grafana Monitoring

除了本地分析，還要監控生產環境的效能：

**Dashboard:** `http://localhost:3030`
- **Game Loop Performance - P95 Dashboard**

**關鍵指標：**
- **P95 Tick Duration** - 應該 < 33ms
- **Tick Rate** - 應該約 30 ticks/sec
- **Entity Count** - 與效能的相關性

**文檔：**
- `/game-server/observability/GAME_LOOP_P95_GUIDE.md`
- `/game-server/observability/QUICKSTART.md`

---

## 🔧 常用命令速查

### 逃逸分析
```bash
# 完整分析
go build -a -gcflags='all=-m' ./internal/game 2>&1 | tee reports/escape_analysis.txt

# 只看 session.go
grep "session.go" reports/escape_analysis.txt | grep "escapes"

# 找出最嚴重的
grep "moved to heap" reports/escape_analysis.txt
```

### CPU Profiling
```bash
# 收集 30 秒的 CPU profile
go tool pprof http://localhost:8082/debug/pprof/profile?seconds=30

# 或使用腳本
./scripts/2_cpu_profile.sh
```

### Memory Profiling
```bash
# Heap profiling
go tool pprof http://localhost:8082/debug/pprof/heap

# Alloc profiling
go tool pprof http://localhost:8082/debug/pprof/allocs
```

### Benchmark
```bash
# 執行所有 benchmark
go test -bench=. -benchmem ./...

# 只測試特定函數
go test -bench=BenchmarkGameLoop -benchmem ./internal/game

# 多次執行取平均
go test -bench=. -benchmem -count=10 ./internal/game
```

---

## 📚 相關資源

### 內部文檔
- [Game Loop P95 監控指南](../../observability/GAME_LOOP_P95_GUIDE.md)
- [Observability 快速開始](../../observability/QUICKSTART.md)

### 外部資源
- [Go Profiling Guide](https://go.dev/blog/pprof)
- [Escape Analysis](https://medium.com/@yulang.chu/go-escape-analysis-b07fc3b15c0d)
- [High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html)

---

## 🔄 更新歷史

| 日期 | 內容 | 作者 |
|------|------|------|
| 2026-02-26 | 初始建立 performance-analysis 資料夾結構 | Team |
| 2026-02-26 | 新增逃逸分析腳本和報告 | Team |

---

## 💡 小技巧

1. **定期分析** - 建議每次重大更新後都執行完整分析
2. **版本控制** - 報告加上日期標籤（例如：`escape_2026-02-26.txt`）
3. **比較分析** - 保留優化前後的報告做對比
4. **自動化** - 可以整合到 CI/CD 中自動執行

---

**維護者**: Game Service Team
**最後更新**: 2026-02-26
