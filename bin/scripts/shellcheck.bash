#!/usr/bin/env bash

exec shellcheck -s bash -x \
  bin/download bin/install bin/list-all bin/scripts/utils.bash
