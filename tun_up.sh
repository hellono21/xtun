#!/bin/bash

ip link set dev tun0 mtu 1300
ip addr add 10.55.0.1/24 dev tun0
ip link set dev tun0 up
