export type Peer = {
  id: number
  name: string
  public_key: string
  private_key?: string
  address_v4: string
  created_at: string
}

export type PeerStatus = {
  public_key: string
  endpoint: string
  allowed_ips: string
  latest_handshake: number
  rx_bytes: number
  tx_bytes: number
}

export type Status = {
  listen_port: number
  app_domain: string
  peers: PeerStatus[]
}

async function json<T>(r: Response): Promise<T> {
  if (r.status === 401) {
    window.location.href = '/auth/login'
    throw new Error('unauthenticated')
  }
  if (!r.ok) {
    throw new Error(`${r.status}: ${await r.text()}`)
  }
  return r.json() as Promise<T>
}

export const api = {
  listPeers: () => fetch('/api/peers').then(json<Peer[]>),
  createPeer: (name: string) =>
    fetch('/api/peers', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    }).then(json<Peer>),
  deletePeer: (id: number) =>
    fetch(`/api/peers/${id}`, { method: 'DELETE' }).then((r) => {
      if (!r.ok) throw new Error(`${r.status}`)
    }),
  peerConfigURL: (id: number) => `/api/peers/${id}/config`,
  peerConfigText: async (id: number): Promise<string> => {
    const r = await fetch(`/api/peers/${id}/config`)
    if (!r.ok) throw new Error(`${r.status}: ${await r.text()}`)
    return await r.text()
  },
  status: () => fetch('/api/status').then(json<Status>),
}
