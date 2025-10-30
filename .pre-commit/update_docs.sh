#! /bin/bash

set -euo pipefail

script_path_relative="${BASH_SOURCE[0]:-$0}"
script_dir_relative="$(dirname -- $script_path_relative)"
workspace_root="$(cd -- "$script_dir_relative"/.. && pwd)"
snaggle_src="$workspace_root/cmd/snaggle"
snaggle_bin="$workspace_root/.build/snaggle"

echo "Removing old $snaggle_bin..."
rm "$snaggle_bin"

echo "Building new $snaggle_bin from $snaggle_src..."
go build -o "$snaggle_bin" "$snaggle_src"

helptext="$($workspace_root/.build/snaggle --help)"
echo "$helptext"

perl -g -i -e 's/(?<header>^\/\*\n)(?<doccomment>\X*)(?<code>\*\/\npackage main\X*)/$header foo $code/gsm' "$snaggle_src/main.go"
