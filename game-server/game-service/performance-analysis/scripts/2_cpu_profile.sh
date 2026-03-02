#!/bin/bash

# ====================================================================
# CPU Profiling 腳本
# 用途: 分析 CPU 熱點，找出哪些函數佔用最多 CPU 時間
# ====================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROFILE_DIR="$SCRIPT_DIR/../profiles"

echo "╔════════════════════════════════════════════════════════════╗"
echo "║           🔥 CPU Profiling                                ║"
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

# 收集 CPU profile
DURATION=30
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
PROFILE_FILE="$PROFILE_DIR/cpu_${TIMESTAMP}.prof"

echo "🔍 收集 CPU profile (持續 ${DURATION} 秒)..."
echo "   在此期間請盡量模擬真實使用場景（例如：多個玩家進入遊戲）"
echo ""

go tool pprof -output="$PROFILE_FILE" \
    "http://localhost:8082/debug/pprof/profile?seconds=$DURATION"

echo ""
echo "✅ CPU profile 已儲存: $PROFILE_FILE"
echo ""

# 建立符號連結
ln -sf "cpu_${TIMESTAMP}.prof" "$PROFILE_DIR/cpu_latest.prof"

# 生成文字報告
REPORT_FILE="$PROFILE_DIR/cpu_report_${TIMESTAMP}.txt"

echo "📊 生成報告..."
{
    echo "======================================"
    echo "CPU Profiling 報告"
    echo "收集時間: $(date)"
    echo "持續時長: ${DURATION} 秒"
    echo "======================================"
    echo ""

    echo "📊 Top 20 Functions (by cumulative CPU):"
    echo "--------------------------------------"
    go tool pprof -top -cum "$PROFILE_FILE" 2>/dev/null | head -30
    echo ""

    echo "======================================"
    echo "🎮 Game Loop 相關函數:"
    echo "--------------------------------------"
    go tool pprof -top "$PROFILE_FILE" 2>/dev/null | grep -E "manageGameLoop|broadcastFullState|Update|GetAllEntities" || echo "  (未找到)"
    echo ""

} > "$REPORT_FILE"

cat "$REPORT_FILE"

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  📁 生成的檔案                                             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo "  1. $PROFILE_FILE"
echo "  2. $REPORT_FILE"
echo "  3. $PROFILE_DIR/cpu_latest.prof (符號連結)"
echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  🔧 進階分析命令                                           ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "# 互動式分析 (推薦)"
echo "go tool pprof $PROFILE_FILE"
echo ""
echo "# 在 pprof 互動模式中可用的命令:"
echo "  (pprof) top 20              # 顯示前 20 個函數"
echo "  (pprof) top 20 -cum         # 按累計時間排序"
echo "  (pprof) list manageGameLoop # 查看函數詳細 profile"
echo "  (pprof) web                 # 生成圖形視圖 (需要 graphviz)"
echo "  (pprof) pdf > cpu.pdf       # 匯出 PDF"
echo "  (pprof) help                # 查看所有命令"
echo ""
echo "# 生成圖形化報告 (需要先安裝 graphviz)"
echo "go tool pprof -http=:8081 $PROFILE_FILE"
echo "# 然後打開瀏覽器: http://localhost:8081"
