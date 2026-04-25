package firewall

import (
	"fmt"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
)

type Firewall struct {
	TableName     string
	InternalIface string
	ExternalIface string
}

var families = []nftables.TableFamily{
	nftables.TableFamilyIPv4,
	nftables.TableFamilyIPv6,
}

func (f *Firewall) Apply() error {
	var lastErr error
	applied := 0
	for _, family := range families {
		if err := f.applyFamily(family); err != nil {
			lastErr = err
			continue
		}
		applied++
	}
	if applied == 0 {
		return fmt.Errorf("no nftables family usable: %w", lastErr)
	}
	return nil
}

func (f *Firewall) applyFamily(family nftables.TableFamily) error {
	c, err := nftables.New()
	if err != nil {
		return fmt.Errorf("nftables conn: %w", err)
	}

	table := &nftables.Table{
		Family: family,
		Name:   f.TableName,
	}
	c.AddTable(table)
	c.FlushTable(table)

	forward := c.AddChain(&nftables.Chain{
		Name:     "forward",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookForward,
		Priority: nftables.ChainPriorityFilter,
	})
	c.AddRule(&nftables.Rule{
		Table: table,
		Chain: forward,
		Exprs: []expr.Any{
			&expr.Meta{Key: expr.MetaKeyIIFNAME, Register: 1},
			&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: ifname(f.InternalIface)},
			&expr.Verdict{Kind: expr.VerdictAccept},
		},
	})

	postrouting := c.AddChain(&nftables.Chain{
		Name:     "postrouting",
		Table:    table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
	})
	c.AddRule(&nftables.Rule{
		Table: table,
		Chain: postrouting,
		Exprs: []expr.Any{
			&expr.Meta{Key: expr.MetaKeyOIFNAME, Register: 1},
			&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: ifname(f.ExternalIface)},
			&expr.Masq{},
		},
	})

	return c.Flush()
}

func (f *Firewall) Teardown() error {
	c, err := nftables.New()
	if err != nil {
		return fmt.Errorf("nftables conn: %w", err)
	}
	for _, family := range families {
		c.DelTable(&nftables.Table{
			Family: family,
			Name:   f.TableName,
		})
	}
	return c.Flush()
}

func ifname(n string) []byte {
	b := make([]byte, 16)
	copy(b, n+"\x00")
	return b
}
