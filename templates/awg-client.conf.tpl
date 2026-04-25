[Interface]
PrivateKey = {{ .Peer.PrivateKey }}
Address    = {{ .Peer.AddressV4 }}, {{ .PeerV6 }}/64
DNS        = 10.9.0.1, 2001:4860:4860::8888

Jc = {{ .Config.Obfuscation.Jc }}
Jmin = {{ .Config.Obfuscation.Jmin }}
Jmax = {{ .Config.Obfuscation.Jmax }}
S1 = {{ .Config.Obfuscation.S1 }}
S2 = {{ .Config.Obfuscation.S2 }}
H1 = {{ .Config.Obfuscation.H1 }}
H2 = {{ .Config.Obfuscation.H2 }}
H3 = {{ .Config.Obfuscation.H3 }}
H4 = {{ .Config.Obfuscation.H4 }}

[Peer]
PublicKey  = {{ .ServerPublicKey }}
Endpoint   = {{ .Endpoint }}
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
