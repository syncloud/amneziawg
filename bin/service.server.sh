#!/bin/bash -e

export PATH=$SNAP/amneziawg-tools/bin:$SNAP/bin:/usr/sbin:/sbin:$PATH
export WG_QUICK_USERSPACE_IMPLEMENTATION=$SNAP/amneziawg-go/amneziawg-go
export WG_SUDO=1
export LOG_LEVEL=verbose

INTERFACE=awg0
CONF=$SNAP_DATA/config/${INTERFACE}.conf

teardown() {
  $SNAP/amneziawg-tools/bin/awg-quick down $CONF || true
  $SNAP/bin/firewall teardown || true
}
trap teardown INT TERM EXIT

echo 1 > /proc/sys/net/ipv4/ip_forward || true
echo 1 > /proc/sys/net/ipv6/conf/all/forwarding || true

$SNAP/bin/firewall apply
$SNAP/amneziawg-tools/bin/awg-quick up $CONF

while true; do sleep 60; done
