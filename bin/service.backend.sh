#!/bin/bash -e
/bin/rm -f $SNAP_DATA/backend.sock
exec $SNAP/backend/backend \
  -socket $SNAP_DATA/backend.sock \
  -config-dir $SNAP_DATA/config \
  -data-dir $SNAP_DATA \
  -common-dir $SNAP_COMMON
