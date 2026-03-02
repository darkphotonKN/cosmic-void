#!/bin/bash

# ====================================================================
# Memory Profiling 腳本
# 用途: 分析記憶體分配熱點
# ====================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROFILE_DIR="$SCRIPT_DIR/../profiles"

echo "╔════════════════════════════════════════════════════════════╗"
echo "║           💾 Memory Profiling                             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# 檢查 game-service 是否運行
if ! curl -s http://localhost:8082/debug/pprof/ > /dev/null 2>&1; then
    echo "❌ 錯誤: game-service 沒有運行"
    echo ""
    echo "請先啟動 game-service:"
    echo "  cd game-server/game-service"
    echo "  go run ./cmd/server"
    echo ""
    exit 1
fi

echo "✅ game-service 正在運行"
echo ""

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# 收集 Heap profile
echo "🔍 收集 Heap profile..."
HEAP_FILE="$PROFILE_DIR/heap_${TIMESTAMP}.prof"
go tool pprof -output="$HEAP_FILE" \
    "http://localhost:8082/debug/pprof/heap"
echo "✅ Heap profile 已儲存: $HEAP_FILE"
echo ""

# 收集 Allocs profile
echo "🔍 收集 Allocs profile..."
ALLOC_FILE="$PROFILE_DIR/allocs_${TIMESTAMP}.prof"
go tool pprof -output="$ALLOC_FILE" \
    "http://localhost:8082/debug/pprof/allocs"
echo "✅ Allocs profile 已儲存: $ALLOC_FILE"
echo ""

# 建立符號連結
ln -sf "heap_${TIMESTAMP}.prof" "$PROFILE_DIR/heap_latest.prof"
ln -sf "allocs_${TIMESTAMP}.prof" "$PROFILE_DIR/allocs_latest.prof"

# 生成報告
REPORT_FILE="$PROFILE_DIR/memory_report_${TIMESTAMP}.txt"

echo "📊 生成報告..."
{
    echo "======================================"
    echo "Memory Profiling 報告"
    echo "收集時間: $(date)"
    echo "======================================"
    echo ""

    echo "======================================"
    echo "📊 Heap (當前使用) - Top 20"
    echo "======================================"
    go tool pprof -top -inuse_space "$HEAP_FILE" 2>/dev/null | head -30
    echo ""

    echo "======================================"
    echo "📊 Allocs (總分配) - Top 20"
    echo "======================================"
    go tool pprof -top -alloc_space "$ALLOC_FILE" 2>/dev/null | head -30
    echo ""

    echo "======================================"
    echo "🎮 Game Loop 相關分配:"
    echo "======================================"
    go tool pprof -top -alloc_space "$ALLOC_FILE" 2>/dev/null | \
        grep -E "manageGameLoop|broadcastFullState|GetAllEntities|Serialize" || echo "  (未找到)"
    echo ""

} > "$REPORT_FILE"

cat "$REPORT_FILE"

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  📁 生成的檔案                                             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo "  1. $HEAP_FILE"
echo "  2. $ALLOC_FILE"
echo "  3. $REPORT_FILE"
echo "  4. $PROFILE_DIR/heap_latest.prof (符號連結)"
echo "  5. $PROFILE_DIR/allocs_latest.prof (符號連結)"
echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  🔧 進階分析命令                                           ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "# 分析 Heap (當前使用的記憶體)"
echo "go tool pprof $HEAP_FILE"
echo "  (pprof) top -inuse_space     # 按當前使用量排序"
echo "  (pprof) list broadcastFullState"
echo ""
echo "# 分析 Allocs (總分配量，找熱點)"
echo "go tool pprof $ALLOC_FILE"
echo "  (pprof) top -alloc_space     # 按總分配量排序"
echo "  (pprof) list manageGameLoop"
echo ""
echo "# 生成圖形化報告"
echo "go tool pprof -http=:8081 $HEAP_FILE"
echo "go tool pprof -http=:8081 $ALLOC_FILE"
