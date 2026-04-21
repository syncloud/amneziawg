package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"hooks/firewall"
)

const (
	tableName     = "amneziawg"
	internalIface = "awg0"
	externalIface = "eth0"
)

func main() {
	fw := &firewall.Firewall{
		TableName:     tableName,
		InternalIface: internalIface,
		ExternalIface: externalIface,
	}

	cmd := &cobra.Command{
		Use:          "firewall",
		SilenceUsage: true,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "apply",
		Short: "Install the forward-accept and postrouting-masquerade rules for the VPN interface",
		RunE:  func(*cobra.Command, []string) error { return fw.Apply() },
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "teardown",
		Short: "Remove the rules installed by 'apply'",
		RunE:  func(*cobra.Command, []string) error { return fw.Teardown() },
	})

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
