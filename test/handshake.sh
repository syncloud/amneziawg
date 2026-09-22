#!/bin/bash
set -e

SNAP=/snap/amneziawg/current
AWG=$SNAP/amneziawg-tools/bin/awg
AWGQUICK=$SNAP/amneziawg-tools/bin/awg-quick
AWGGO=$SNAP/amneziawg-go/amneziawg-go
SRVCONF=/var/snap/amneziawg/current/config/awg0.conf
NS=awghs
CLICONF=/tmp/awg-handshake-client.conf
export PATH=$SNAP/amneziawg-tools/bin:$SNAP/bin:/usr/sbin:/sbin:$PATH
export WG_QUICK_USERSPACE_IMPLEMENTATION=$AWGGO
export WG_SUDO=1

for b in "$AWG" "$AWGQUICK" "$AWGGO"; do [ -x "$b" ] || { echo "missing $b"; exit 1; }; done

SERVER_PUB=$(cat /var/snap/amneziawg/current/server.pub)
PORT=$($AWG show awg0 listen-port)
OBF=$(grep -E '^(Jc|Jmin|Jmax|S1|S2|S3|S4|H1|H2|H3|H4|I1|I2|I3|I4|I5) ' $SRVCONF)
echo "server pub=$SERVER_PUB port=$PORT"
echo "obfuscation params mirrored to client:"; echo "$OBF"

cleanup() {
  ip netns exec $NS env PATH=$PATH WG_QUICK_USERSPACE_IMPLEMENTATION=$AWGGO $AWGQUICK down $CLICONF 2>/dev/null || true
  ip netns pids $NS 2>/dev/null | xargs -r kill 2>/dev/null || true
  ip netns del $NS 2>/dev/null || true
  ip link del veth-h 2>/dev/null || true
  [ -n "$CLIENT_PUB" ] && $AWG set awg0 peer $CLIENT_PUB remove 2>/dev/null || true
  rm -f $CK $CLICONF
}
trap cleanup EXIT

# fresh state
ip netns del $NS 2>/dev/null || true
ip link del veth-h 2>/dev/null || true

CK=$(mktemp)
$AWG genkey > $CK
CLIENT_PUB=$(cat $CK | $AWG pubkey)
echo "client pub=$CLIENT_PUB"

# veth carries the encapsulated UDP between the client netns and the server
ip netns add $NS
ip link add veth-h type veth peer name veth-c
ip link set veth-c netns $NS
ip addr add 172.31.0.1/30 dev veth-h
ip link set veth-h up
ip netns exec $NS ip addr add 172.31.0.2/30 dev veth-c
ip netns exec $NS ip link set veth-c up
ip netns exec $NS ip link set lo up

cat > $CLICONF <<EOF
[Interface]
PrivateKey = $(cat $CK)
Address = 10.9.0.77/24
$OBF

[Peer]
PublicKey = $SERVER_PUB
Endpoint = 172.31.0.1:$PORT
AllowedIPs = 10.9.0.1/32
PersistentKeepalive = 15
EOF
chmod 600 $CLICONF

$AWG set awg0 peer $CLIENT_PUB allowed-ips 10.9.0.77/32

ip netns exec $NS env WG_QUICK_USERSPACE_IMPLEMENTATION=$AWGGO WG_SUDO=1 PATH=$PATH \
  $AWGQUICK up $CLICONF 2>&1 | sed 's/^/  [awg-quick] /'
CLIIF=$(basename $CLICONF .conf)

HS=0
for i in $(seq 1 20); do
  sleep 2
  HS=$(ip netns exec $NS $AWG show $CLIIF latest-handshakes | awk '{print $2}' | head -1)
  [ "${HS:-0}" -gt 0 ] 2>/dev/null && { echo "handshake completed at t+$((i*2))s"; break; }
done

ip netns exec $NS $AWG show $CLIIF

PING_OK=skip
if ip netns exec $NS sh -c 'command -v ping' >/dev/null 2>&1; then
  if ip netns exec $NS ping -c 4 -W 2 10.9.0.1; then PING_OK=1; else PING_OK=0; fi
else
  echo "ping not available on device; gating on handshake + transfer only"
fi

RX=$(ip netns exec $NS $AWG show $CLIIF transfer | awk '{print $2}' | head -1)
echo "RESULT handshake=${HS:-0} rx_bytes=${RX:-0} ping=${PING_OK}"

# gate: handshake must complete AND the client must have received encrypted
# bytes back (proves a bidirectional tunnel, not just a sent initiation).
# When ping exists it must also succeed.
FAIL=0
[ "${HS:-0}" -gt 0 ] || { echo "no handshake"; FAIL=1; }
[ "${RX:-0}" -gt 0 ] || { echo "no bytes received"; FAIL=1; }
[ "$PING_OK" = "0" ] && { echo "ping failed"; FAIL=1; }
[ "$FAIL" = "0" ] && echo "CONNECTION TEST PASSED" || { echo "CONNECTION TEST FAILED"; exit 1; }
