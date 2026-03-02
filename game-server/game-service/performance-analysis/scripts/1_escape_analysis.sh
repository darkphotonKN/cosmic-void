#!/bin/bash

# ====================================================================
# 逃逸分析腳本 (Escape Analysis)
# 用途: 分析哪些變數從 stack 逃逸到 heap
# ====================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
REPORT_DIR="$SCRIPT_DIR/../reports"

echo "╔════════════════════════════════════════════════════════════╗"
echo "║           🔍 Go Escape Analysis                           ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "📁 項目路徑: $PROJECT_ROOT"
echo "📊 報告路徑: $REPORT_DIR"
echo ""

cd "$PROJECT_ROOT"

# 清理舊的 build cache
echo "🧹 清理 build cache..."
go clean -cache
echo ""

# 執行逃逸分析
echo "🔍 執行逃逸分析..."
echo "   這可能需要幾秒鐘..."
echo ""

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$REPORT_DIR/escape_analysis_${TIMESTAMP}.txt"

# 分析所有相關 packages
go build -a -gcflags='all=-m -m' \
    ./internal/game \
    ./internal/systems \
    ./internal/components/metrics \
    > "$REPORT_FILE" 2>&1

echo "✅ 完整報告已儲存: $REPORT_FILE"
echo ""

# 建立符號連結指向最新報告
ln -sf "escape_analysis_${TIMESTAMP}.txt" "$REPORT_DIR/escape_analysis_latest.txt"

# 生成摘要報告
echo "📊 生成摘要報告..."
echo ""

SUMMARY_FILE="$REPORT_DIR/escape_summary_${TIMESTAMP}.txt"

{
    echo "======================================"
    echo "逃逸分析摘要報告"
    echo "生成時間: $(date)"
    echo "======================================"
    echo ""

    echo "📊 統計資訊:"
    echo "  總行數: $(wc -l < "$REPORT_FILE")"
    echo "  逃逸到 heap: $(grep -c 'escapes to heap' "$REPORT_FILE")"
    echo "  移動到 heap: $(grep -c 'moved to heap' "$REPORT_FILE")"
    echo "  參數逃逸: $(grep -c 'leaking param' "$REPORT_FILE")"
    echo ""

    echo "======================================"
    echo "🎮 Session.go (前 20 個逃逸)"
    echo "======================================"
    grep "session.go" "$REPORT_FILE" | grep "escapes to heap" | head -20
    echo ""

    echo "======================================"
    echo "⚠️  嚴重逃逸 (moved to heap)"
    echo "======================================"
    grep "moved to heap" "$REPORT_FILE" | grep -E "session.go|movement.go|metrics.go" | head -15
    echo ""

} > "$SUMMARY_FILE"

echo "✅ 摘要報告已儲存: $SUMMARY_FILE"
echo ""

# 顯示摘要
cat "$SUMMARY_FILE"

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  📁 生成的檔案                                             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo "  1. $REPORT_FILE"
echo "  2. $SUMMARY_FILE"
echo "  3. $REPORT_DIR/escape_analysis_latest.txt (符號連結)"
echo ""
echo "💡 提示:"
echo "   查看完整報告: cat $REPORT_FILE"
echo "   查看摘要:     cat $SUMMARY_FILE"
echo "   執行快速分析: ./scripts/quick_escape_analysis.sh"
