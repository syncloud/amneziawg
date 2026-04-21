[Interface]
PrivateKey = {{ .ServerPrivateKey }}
ListenPort = {{ .ListenPort }}
Address    = 10.9.0.1/24

Jc = {{ .Jc }}
Jmin = {{ .Jmin }}
Jmax = {{ .Jmax }}
S1 = {{ .S1 }}
S2 = {{ .S2 }}
H1 = {{ .H1 }}
H2 = {{ .H2 }}
H3 = {{ .H3 }}
H4 = {{ .H4 }}

{{ range .Peers }}
[Peer]
# name: {{ .Name }}
PublicKey  = {{ .PublicKey }}
AllowedIPs = {{ .AllowedIPs }}
{{ end }}
