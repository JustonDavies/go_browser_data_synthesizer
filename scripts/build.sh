#!/usr/bin/env bash
echo Building justondavies/go_activity_synthesizer:build

sudo docker build                                         \
  --network host                                          \
  --file dockerfiles/all.docker                           \
  --tag justondavies/go_browser_history_synthesizer:build \
  ./

sudo docker create                                   \
  --name build_extract                               \
  justondavies/go_browser_history_synthesizer:build

rm -rf ./build/browser*

sudo docker cp                                                                       \
  build_extract:/go/src/github.com/justondavies/go_browser_history_synthesizer/build \
  ./

sudo docker rm -f build_extract

sudo chown -R $USER:$USER ./build
chmod -R 700 ./build
