package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main() {
	cmd := &cobra.Command{
		Use:          "backend",
		Short:        "AmneziaWG web backend — peer management API + OIDC auth",
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			logger, err := buildLogger()
			if err != nil {
				return err
			}
			b, err := newBackend()
			if err != nil {
				return err
			}
			return b.Serve(logger)
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use:          "render-server-conf",
		Short:        "Regenerate " + serverIface + ".conf from the peers database",
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			b, err := newBackend()
			if err != nil {
				return err
			}
			return b.RenderServerConf()
		},
	})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildLogger() (*zap.Logger, error) {
	c := zap.NewProductionConfig()
	c.Encoding = "console"
	c.EncoderConfig.TimeKey = ""
	c.OutputPaths = []string{"stdout"}
	c.ErrorOutputPaths = []string{"stderr"}
	return c.Build()
}
