#!/bin/bash

# Quick Escape Analysis Script
# Usage: ./quick_escape_analysis.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPORT_DIR="$SCRIPT_DIR/../reports"
ESCAPE_FILE="$REPORT_DIR/escape_analysis_latest.txt"

# 如果沒有 latest，嘗試找最新的檔案
if [ ! -f "$ESCAPE_FILE" ]; then
    LATEST_FILE=$(ls -t "$REPORT_DIR"/escape_analysis_*.txt 2>/dev/null | head -1)
    if [ -n "$LATEST_FILE" ]; then
        ESCAPE_FILE="$LATEST_FILE"
    else
        echo "❌ Error: No escape analysis report found!"
        echo ""
        echo "請先執行逃逸分析:"
        echo "  ./scripts/1_escape_analysis.sh"
        echo ""
        echo "或手動執行:"
        echo "  cd ../.."
        echo "  go build -a -gcflags='all=-m' ./internal/game 2>&1 | tee performance-analysis/reports/escape_analysis.txt"
        exit 1
    fi
fi

echo "📁 使用報告: $ESCAPE_FILE"
echo ""

echo "╔════════════════════════════════════════════════════════════╗"
echo "║         🔍 Game Loop Escape Analysis Report              ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# 基本統計
echo "📊 基本統計"
echo "────────────────────────────────────────────────────────────"
echo "  總行數:        $(wc -l < $ESCAPE_FILE)"
echo "  逃逸總數:      $(grep -c 'escapes to heap' $ESCAPE_FILE)"
echo "  強制移到 heap: $(grep -c 'moved to heap' $ESCAPE_FILE)"
echo "  參數逃逸:      $(grep -c 'leaking param' $ESCAPE_FILE)"
echo ""

# Session.go 分析
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  🎮 session.go (Game Loop Core)                           ║"
echo "╚════════════════════════════════════════════════════════════╝"
grep "session.go" $ESCAPE_FILE | grep "escapes to heap" | head -15
echo ""

# 關鍵函數分析
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  ⚡ Critical Functions                                     ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "📌 manageGameLoop():"
grep "manageGameLoop" $ESCAPE_FILE | grep -E "escapes|moved" || echo "  (無逃逸)"
echo ""

echo "📌 broadcastFullState():"
grep "broadcastFullState" $ESCAPE_FILE | grep -E "escapes|moved" || echo "  (無逃逸)"
echo ""

echo "📌 GetAllEntities():"
grep "GetAllEntities" $ESCAPE_FILE | grep -E "escapes|moved" || echo "  (無逃逸)"
echo ""

# Metrics 分析
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  📊 Metrics Recording                                      ║"
echo "╚════════════════════════════════════════════════════════════╝"
grep -E "TickDuration|EntityCount" $ESCAPE_FILE | grep -E "Record|escapes" | head -5
echo ""

# Movement System
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  🏃 Movement System                                        ║"
echo "╚════════════════════════════════════════════════════════════╝"
grep "movement.go" $ESCAPE_FILE | grep "escapes to heap" | head -8
echo ""

# 最嚴重的逃逸 (moved to heap)
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  ⚠️  Most Critical (moved to heap)                         ║"
echo "╚════════════════════════════════════════════════════════════╝"
grep "moved to heap" $ESCAPE_FILE | grep -E "session.go|movement.go|metrics.go" | head -10
echo ""

echo "╔════════════════════════════════════════════════════════════╗"
echo "║  💾 Output Files Created                                   ║"
echo "╚════════════════════════════════════════════════════════════╝"

# 建立分類文件
grep "session.go" $ESCAPE_FILE | grep "escapes" > session_escapes.txt
grep "movement.go" $ESCAPE_FILE | grep "escapes" > movement_escapes.txt
grep "moved to heap" $ESCAPE_FILE > critical_moved_to_heap.txt

echo "  ✅ session_escapes.txt       - Session 逃逸詳情"
echo "  ✅ movement_escapes.txt      - Movement 逃逸詳情"
echo "  ✅ critical_moved_to_heap.txt - 最嚴重的逃逸"
echo ""

echo "📖 查看詳細內容："
echo "   cat session_escapes.txt"
echo "   cat movement_escapes.txt"
echo "   cat critical_moved_to_heap.txt"
