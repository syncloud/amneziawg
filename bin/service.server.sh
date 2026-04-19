#!/bin/bash -e

export PATH=$SNAP/amneziawg-tools/bin:$SNAP/bin:$PATH
export WG_QUICK_USERSPACE_IMPLEMENTATION=$SNAP/amneziawg-go/amneziawg-go
export WG_SUDO=1

INTERFACE=awg0
CONF=$SNAP_DATA/config/${INTERFACE}.conf

trap "$SNAP/amneziawg-tools/bin/awg-quick down $CONF" INT TERM EXIT

$SNAP/amneziawg-tools/bin/awg-quick up $CONF

# awg-quick brings up the interface then returns; keep the unit alive so
# the tun device and routes persist for the lifetime of the service.
while true; do sleep 60; done
