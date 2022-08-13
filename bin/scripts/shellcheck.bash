#!/usr/bin/env bash

exec shellcheck -s bash -x \
  bin/download \
  bin/install \
  bin/list-all \
  -P bin/scripts/
