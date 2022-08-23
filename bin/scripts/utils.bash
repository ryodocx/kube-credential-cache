#!/usr/bin/env bash

set -euo pipefail

GH_REPO="https://github.com/ryodocx/kube-credential-cache"
TOOL_NAME="kube-credential-cache"

fail() {
  echo -e "asdf-$TOOL_NAME: $*"
  exit 1
}

curl_opts=(-fsSL)

if [ -n "${GITHUB_API_TOKEN:-}" ]; then
  curl_opts=("${curl_opts[@]}" -H "Authorization: token $GITHUB_API_TOKEN")
fi

sort_versions() {
  sed 'h; s/[+-]/./g; s/.p\([[:digit:]]\)/.z\1/; s/$/.z/; G; s/\n/ /' |
    LC_ALL=C sort -t. -k 1,1 -k 2,2n -k 3,3n -k 4,4n -k 5,5n | awk '{print $2}'
}

list_github_tags() {
  if which jq &>/dev/null; then
    # TODO: support pagination
    curl "${curl_opts[@]}" "https://api.github.com/repos/ryodocx/kube-credential-cache/releases?per_page=100" |
      jq -r '.[] | select(.prerelease == false) | .tag_name' |
      sed 's/^v//'
  else
    git ls-remote --tags --refs "$GH_REPO" |
      grep -o 'refs/tags/.*' | cut -d/ -f3- |
      sed 's/^v//'
  fi
}

list_all_versions() {
  list_github_tags
}

download_release() {
  local version filename url
  version="$1"
  filename="$2"

  url="$GH_REPO/releases/download/v${version}/kube-credential-cache_${version}_$(uname -s)_$(uname -m).tar.gz"

  echo "* Downloading $TOOL_NAME release $version..."
  curl "${curl_opts[@]}" -o "$filename" -C - "$url" || fail "Could not download $url"
}

install_version() {
  local install_type="$1"
  local version="$2"
  local install_path="$3"

  if [ "$install_type" != "version" ]; then
    fail "asdf-$TOOL_NAME supports release installs only"
  fi

  (
    mkdir -p "$install_path/bin"
    cp "$ASDF_DOWNLOAD_PATH/kcc-cache" "$install_path/bin/"
    cp "$ASDF_DOWNLOAD_PATH/kcc-injector" "$install_path/bin/"

    local tool_cmd="kcc-cache"
    test -x "$install_path/bin/$tool_cmd" || fail "Expected $install_path/bin/$tool_cmd to be executable."
    local tool_cmd2="kcc-injector"
    test -x "$install_path/bin/$tool_cmd2" || fail "Expected $install_path/bin/$tool_cmd2 to be executable."

    echo "$TOOL_NAME $version installation was successful!"
  ) || (
    rm -rf "$install_path"
    fail "An error ocurred while installing $TOOL_NAME $version."
  )
}
