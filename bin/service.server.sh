#!/bin/bash -e

export PATH=$SNAP/amneziawg-tools/bin:$SNAP/bin:/usr/sbin:/sbin:$PATH
export WG_QUICK_USERSPACE_IMPLEMENTATION=$SNAP/amneziawg-go/amneziawg-go
export WG_SUDO=1

INTERFACE=awg0
CONF=$SNAP_DATA/config/${INTERFACE}.conf

teardown() {
  $SNAP/amneziawg-tools/bin/awg-quick down $CONF || true
  $SNAP/bin/firewall teardown || true
}
trap teardown INT TERM EXIT

$SNAP/bin/firewall apply
$SNAP/amneziawg-tools/bin/awg-quick up $CONF

RUN=/var/snap/amneziawg/current/run/amneziawg
SOCK=$RUN/${INTERFACE}.sock
if [ -S "$SOCK" ]; then
  chgrp amneziawg "$RUN" "$SOCK"
  chmod 750 "$RUN"
  chmod 660 "$SOCK"
fi

while true; do sleep 60; done
