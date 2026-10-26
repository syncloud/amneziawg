package peers

import (
	"bufio"
	"net"
	"os"
	"strings"
)

type LanResolver struct {
	ResolvConf string
}

func (r *LanResolver) Resolve() (Resolver, error) {
	file, err := os.Open(r.ResolvConf)
	if os.IsNotExist(err) {
		return PublicResolver, nil
	}
	if err != nil {
		return Resolver{}, err
	}
	defer file.Close()

	var resolver Resolver
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "nameserver":
			ip := net.ParseIP(strings.SplitN(fields[1], "%", 2)[0])
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
				continue
			}
			resolver.Servers = append(resolver.Servers, ip.String())
		case "domain", "search":
			resolver.Domain = fields[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return Resolver{}, err
	}
	if len(resolver.Servers) == 0 {
		return PublicResolver, nil
	}
	return resolver, nil
}
