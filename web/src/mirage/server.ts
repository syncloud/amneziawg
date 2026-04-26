import { createServer, Response } from 'miragejs'
import type { Peer, Status } from '@/api'

type MirageRequest = {
  params: Record<string, string>
  requestBody?: string
  url: string
}

const state = {
  listenPort: 51234,
  appDomain: 'amneziawg.example.com',
  serverPublicKey: 'FAKE_SERVER_PUBLIC_KEY_FOR_DEV_PREVIEW====',
  peers: [
    {
      id: 1,
      name: 'laptop',
      public_key: 'AAAA_FAKE_LAPTOP_PUBKEY_FOR_DEV====',
      private_key: '',
      address_v4: '10.9.0.2/32',
      created_at: '2026-04-18 12:30:00',
    },
    {
      id: 2,
      name: 'phone',
      public_key: 'BBBB_FAKE_PHONE_PUBKEY_FOR_DEV======',
      private_key: '',
      address_v4: '10.9.0.3/32',
      created_at: '2026-04-19 09:12:00',
    },
  ] as Peer[],
  nextId: 3,
}

function nextFreeAddress(): string {
  const used = new Set(state.peers.map((p) => p.address_v4))
  for (let i = 2; i < 254; i++) {
    const addr = `10.9.0.${i}/32`
    if (!used.has(addr)) return addr
  }
  throw new Error('address pool exhausted')
}

function sampleConf(peer: Peer): string {
  return `[Interface]
PrivateKey = ${peer.private_key || '<redacted>'}
Address    = ${peer.address_v4}
DNS        = 10.9.0.1

Jc = 5
Jmin = 50
Jmax = 1000
S1 = 42
S2 = 91
H1 = 1234567890
H2 = 987654321
H3 = 1357924680
H4 = 2468013579

[Peer]
PublicKey  = ${state.serverPublicKey}
Endpoint   = ${state.appDomain}:${state.listenPort}
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
`
}

function randKey(prefix: string): string {
  return `${prefix}_${Math.random().toString(36).slice(2, 10).toUpperCase()}_STUB====`
}

function statusPeers(): Status['peers'] {
  return state.peers.map((p, idx) => ({
    public_key: p.public_key,
    endpoint: `203.0.113.${100 + p.id}:${40000 + p.id}`,
    allowed_ips: p.address_v4,
    latest_handshake:
      idx === state.peers.length - 1
        ? Math.floor(Date.now() / 1000) - 30
        : Math.floor(Date.now() / 1000) - (p.id * 120),
    rx_bytes: p.id * 123_456,
    tx_bytes: p.id * 45_678,
  }))
}

export function makeServer() {
  return createServer({
    routes() {
      this.get('/api/peers', () => state.peers)

      this.post('/api/peers', (_: unknown, request: MirageRequest) => {
        const body = JSON.parse(request.requestBody ?? '{}') as { name?: string }
        if (!body.name) {
          return new Response(400, {}, { error: 'name is required' })
        }
        const priv = randKey('PRIV')
        const peer: Peer = {
          id: state.nextId++,
          name: body.name,
          public_key: randKey('PUB'),
          private_key: priv,
          address_v4: nextFreeAddress(),
          created_at: new Date().toISOString().replace('T', ' ').slice(0, 19),
        }
        state.peers.push({ ...peer, private_key: '' })
        return peer
      })

      this.delete('/api/peers/:id', (_: unknown, request: MirageRequest) => {
        const id = Number(request.params.id)
        state.peers = state.peers.filter((p) => p.id !== id)
        return new Response(204)
      })

      this.get('/api/peers/:id/config', (_: unknown, request: MirageRequest) => {
        const peer = state.peers.find((p) => p.id === Number(request.params.id))
        if (!peer) return new Response(404, {}, 'not found')
        return new Response(200, { 'Content-Type': 'application/octet-stream' }, sampleConf(peer))
      })

this.get('/api/status', (): Status => ({
        listen_port: state.listenPort,
        app_domain: state.appDomain,
        peers: statusPeers(),
      }))
    },
  })
}
