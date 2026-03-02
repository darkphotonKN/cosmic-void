#!/bin/bash

# ====================================================================
# Benchmark 測試腳本
# 用途: 執行效能基準測試
# ====================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
REPORT_DIR="$SCRIPT_DIR/../reports"

echo "╔════════════════════════════════════════════════════════════╗"
echo "║           📊 Benchmark Testing                            ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

cd "$PROJECT_ROOT"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$REPORT_DIR/benchmark_${TIMESTAMP}.txt"

echo "🏃 執行 Benchmark 測試..."
echo "   這可能需要幾分鐘..."
echo ""

# 執行所有 benchmark
{
    echo "======================================"
    echo "Benchmark 報告"
    echo "執行時間: $(date)"
    echo "======================================"
    echo ""

    # 執行 game package 的 benchmark
    echo "🎮 Game Package Benchmarks:"
    echo "--------------------------------------"
    go test -bench=. -benchmem -benchtime=3s ./internal/game/... 2>&1 || echo "  (無 benchmark)"
    echo ""

    # 執行 systems package 的 benchmark
    echo "⚡ Systems Package Benchmarks:"
    echo "--------------------------------------"
    go test -bench=. -benchmem -benchtime=3s ./internal/systems/... 2>&1 || echo "  (無 benchmark)"
    echo ""

    # 執行 serializer package 的 benchmark
    echo "📦 Serializer Package Benchmarks:"
    echo "--------------------------------------"
    go test -bench=. -benchmem -benchtime=3s ./internal/serializer/... 2>&1 || echo "  (無 benchmark)"
    echo ""

} | tee "$REPORT_FILE"

# 建立符號連結
ln -sf "benchmark_${TIMESTAMP}.txt" "$REPORT_DIR/benchmark_latest.txt"

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  📁 生成的檔案                                             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo "  1. $REPORT_FILE"
echo "  2. $REPORT_DIR/benchmark_latest.txt (符號連結)"
echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  📖 如何解讀 Benchmark 結果                                ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "範例輸出:"
echo "  BenchmarkGameLoop-8    50000    35421 ns/op    4096 B/op    12 allocs/op"
echo ""
echo "說明:"
echo "  BenchmarkGameLoop-8  # 測試名稱 (-8 表示 8 個 CPU)"
echo "  50000                # 執行次數"
echo "  35421 ns/op          # 每次操作耗時 (越小越好)"
echo "  4096 B/op            # 每次操作分配的記憶體 (越小越好)"
echo "  12 allocs/op         # 每次操作的分配次數 (越少越好)"
echo ""
echo "🎯 優化目標:"
echo "  • ns/op: 減少執行時間"
echo "  • B/op: 減少記憶體分配"
echo "  • allocs/op: 減少分配次數 (最重要，影響 GC)"
echo ""
echo "💡 比較優化前後:"
echo "  benchstat reports/benchmark_before.txt reports/benchmark_after.txt"
