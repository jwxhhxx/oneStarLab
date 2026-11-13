#!/usr/bin/env bash
# SQLite 在线备份：依赖宿主机 sqlite3（apt install sqlite3）。
# 用法：./backup.sh [保留天数]，默认保留 14 天。
set -euo pipefail

APP_DIR="${APP_DIR:-/opt/onestarlab}"
DB_PATH="${DB_PATH:-$APP_DIR/data/onestar.db}"
BACKUP_DIR="${BACKUP_DIR:-$APP_DIR/backups}"
KEEP_DAYS="${1:-14}"

log() { printf '[backup] %s\n' "$*"; }
die() { printf '[backup] ERROR: %s\n' "$*" >&2; exit 1; }

[ -f "$DB_PATH" ] || die "数据库不存在：$DB_PATH"
command -v sqlite3 >/dev/null 2>&1 || die "缺少 sqlite3，请先安装：sudo apt install -y sqlite3"

mkdir -p "$BACKUP_DIR"
stamp="$(date +%Y%m%d-%H%M%S)"
out="$BACKUP_DIR/onestar-$stamp.db"

sqlite3 "$DB_PATH" ".backup '$out'"
gzip -f "$out"

find "$BACKUP_DIR" -name 'onestar-*.db.gz' -type f -mtime +"$KEEP_DAYS" -delete
log "完成：$out.gz（保留 $KEEP_DAYS 天）"
