#!/bin/bash

#
# Copyright 2021 The CI/CD Operator Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

if [ "$#" -eq 0 ]; then
  VERSION=latest
else
  VERSION=$1
  REGISTRY=$2
fi

BASEDIR=$(dirname "$0")
PROJECT_DIR="$BASEDIR/.."
CONFIG_DIR="$PROJECT_DIR/config"

RELEASE_MANIFEST="$CONFIG_DIR/release.yaml"

TARGETS=("$CONFIG_DIR/controller/controller.yaml" "$CONFIG_DIR/blocker/blocker.yaml" "$CONFIG_DIR/webhook/webhook.yaml" "$CONFIG_DIR/apiserver/apiserver.yaml" "$CONFIG_DIR/rbac/role.yaml" "$CONFIG_DIR/rbac/role_binding.yaml" "$CONFIG_DIR/rbac/service_account.yaml" "$CONFIG_DIR/apiservice" "$CONFIG_DIR/templates")

function append_target(){
  local TARGET="$1"

  if [ "${TARGET: -5}" == ".yaml" ]; then
    cat "$TARGET" >> "$RELEASE_MANIFEST"
    echo "---" >> "$RELEASE_MANIFEST"
  else
    for f in "$TARGET"/*; do
      append_target "$f"
    done
  fi
}

rm -rf "$RELEASE_MANIFEST"

touch "$RELEASE_MANIFEST"

for target in "${TARGETS[@]}"; do
  append_target "$target"
done

sed -i "s/tmaxcloudck\/cicd-operator:latest/$REGISTRY\/cicd-operator:$VERSION/g" "$RELEASE_MANIFEST"
sed -i "s/tmaxcloudck\/cicd-blocker:latest/$REGISTRY\/cicd-blocker:$VERSION/g" "$RELEASE_MANIFEST"
sed -i "s/tmaxcloudck\/cicd-webhook:latest/$REGISTRY\/cicd-webhook:$VERSION/g" "$RELEASE_MANIFEST"
sed -i "s/tmaxcloudck\/cicd-api-server:latest/$REGISTRY\/cicd-api-server:$VERSION/g" "$RELEASE_MANIFEST"