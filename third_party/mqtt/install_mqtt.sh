#!/bin/bash
#
# Copyright (c) 2022 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
set -exo pipefail
#===================================================================================================
# Option parsing

os=${os:-auto}
mqtt_branch=${mqtt_branch:-v1.3.12}
work_dir=${work_dir:-/opt}


#===================================================================================================
# OS detection

if [ "$os" == "auto" ] ; then
    os=$( . /etc/os-release ; echo "${ID}${VERSION_ID}" )
    if [[ "$os" =~ "rhel8".* ]] ; then
      os="rhel8"
    fi
    case $os in
        rhel8|ubuntu18.04|ubuntu20.04|ubuntu21.10|ubuntu22.04) [ -z "$print" ] && echo "Detected OS: ${os}" ;;
        *) echo "Unsupported OS: ${os:-detection failed}" >&2 ; exit 1 ;;
    esac
fi

#===================================================================================================
# MQTT installation

if [ "$os" == "ubuntu20.04" ] ; then
    export DEBIAN_FRONTEND=noninteractive
    apt update && apt install -y build-essential git cmake libssl-dev \
        && rm -rf /var/lib/apt/lists/*
elif [ "$os" == "rhel8" ] ; then
    yum install -d6 -y git cmake gcc-c++
else
    echo "Internal script error: unsupported OS" >&2
    exit 3
fi

current_working_dir=$(pwd)

cd $work_dir
git clone https://github.com/eclipse/paho.mqtt.c.git
cd paho.mqtt.c
git checkout $mqtt_branch
mkdir -p build
cd build
cmake -DCMAKE_PREFIX_PATH=/usr/local/lib/ .. && \
    make "-j$(nproc)" && \
    make install

#===================================================================================================
# end

