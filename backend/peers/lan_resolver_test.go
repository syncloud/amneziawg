package peers

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func resolve(t *testing.T, content string) Resolver {
	t.Helper()
	file := filepath.Join(t.TempDir(), "resolv.conf")
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write resolv.conf: %v", err)
	}
	resolver, err := (&LanResolver{ResolvConf: file}).Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	return resolver
}

func TestLanResolverRouterAndDomain(t *testing.T) {
	got := resolve(t, "domain home.arpa\nsearch home.arpa\nnameserver 192.168.1.1\n")
	want := Resolver{Servers: []string{"192.168.1.1"}, Domain: "home.arpa"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLanResolverKeepsGlobalIPv6SkipsLinkLocal(t *testing.T) {
	got := resolve(t, "nameserver 192.168.1.1\nnameserver 2a00:23c5::1\nnameserver fe80::1%eth0\n")
	want := []string{"192.168.1.1", "2a00:23c5::1"}
	if !reflect.DeepEqual(got.Servers, want) {
		t.Fatalf("got %v, want %v", got.Servers, want)
	}
}

func TestLanResolverSearchListUsesFirstDomain(t *testing.T) {
	if got := resolve(t, "search lan example.com\nnameserver 10.0.0.1\n").Domain; got != "lan" {
		t.Fatalf("got %q, want lan", got)
	}
}

func TestLanResolverLoopbackOnlyFallsBackToPublic(t *testing.T) {
	for _, content := range []string{"nameserver 127.0.0.11\n", "nameserver 127.0.0.53\nsearch lan\n", "nameserver ::1\n"} {
		if got := resolve(t, content); !reflect.DeepEqual(got, PublicResolver) {
			t.Fatalf("%q: got %+v, want public resolver", content, got)
		}
	}
}

func TestLanResolverMissingFileFallsBackToPublic(t *testing.T) {
	got, err := (&LanResolver{ResolvConf: filepath.Join(t.TempDir(), "missing")}).Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !reflect.DeepEqual(got, PublicResolver) {
		t.Fatalf("got %+v, want public resolver", got)
	}
}

func TestLanResolverUnreadableFileIsAnError(t *testing.T) {
	if _, err := (&LanResolver{ResolvConf: t.TempDir()}).Resolve(); err == nil {
		t.Fatal("expected an error reading a directory as resolv.conf")
	}
}
