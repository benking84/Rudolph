#!/usr/bin/env bash
set -euo pipefail

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$(cd -P "$(dirname "$SOURCE")/.." && pwd)"

if [ $# -eq 0 ]; then
  echo "usage: scripts/new_env.sh ENV_NAME" >&2
  exit 1
fi

NEW_ENV="$1"
if [[ ! "$NEW_ENV" =~ ^[a-zA-Z0-9_-]+$ ]] || [ "$NEW_ENV" = "example" ]; then
  echo "invalid environment name: $NEW_ENV (must match ^[a-zA-Z0-9_-]+$ and not be 'example')" >&2
  exit 1
fi
SRC="$DIR/deployments/environments/example"
DEST="$DIR/deployments/environments/$NEW_ENV"

if [ -d "$DEST" ]; then
  echo "environment $NEW_ENV already exists" >&2
  exit 1
fi

mkdir -p "$DEST"
cp "$SRC/main.tf" "$SRC/variables.tf" "$SRC/versions.tf" "$SRC/outputs.tf" "$SRC/config.auto.tfvars.json" "$DEST/"
echo "created $DEST"
echo "edit $DEST/config.auto.tfvars.json before terraform apply"
