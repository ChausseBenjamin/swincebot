#!/bin/sh

localsecrets=".secrets"
localruntime=".runtime"

binpath="build/swincebot"

mkdir -p "$localruntime" || exit 1
[ -f "$localruntime/.gitignore" ] || echo '*' > "$localruntime/.gitignore"

# Values here are not production ready, they are meant to ease development
env_vars=$(cat << EOF
  LOG_LEVEL=debug
  LOG_FORMAT=plain
  LOG_OUTPUT=stdout
  LISTEN_PORT=1157
  DATABASE_PATH=$localruntime/store.db
  GRACEFUL_TIMEOUT=200ms
  DISCORD_GUILD_ID=$(pass swincebot/guild-id)
  DISCORD_CHANNEL_ID=$(pass swincebot/channel-id)
  CGO_ENABLED=1
EOF
)


case "$1" in
  --print-config)
    echo "$env_vars"
    ;;
  --bin)
    clear && shift && env $env_vars "$binpath" $@
    ;;
  *)
    clear && env $env_vars go run . $@
    ;;
esac
