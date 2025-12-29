package commands

import (
	"github.com/onedr0p/exportarr/internal/sabnzbd/collector"
	"github.com/onedr0p/exportarr/internal/sabnzbd/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

func init() {
	config.RegisterSabnzbdFlags(sabnzbdCmd.PersistentFlags())
	rootCmd.AddCommand(sabnzbdCmd)
}

var sabnzbdCmd = &cobra.Command{
	Use:     "sabnzbd",
	Aliases: []string{"sab"},
	Short:   "Prometheus Exporter for Sabnzbd",
	Long:    "Prometheus Exporter for Sabnzbd.",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.LoadSabnzbdConfig(*conf, cmd.PersistentFlags())
		if err != nil {
			return err
		}
		if err := c.Validate(); err != nil {
			return err
		}

		collector, err := collector.NewSabnzbdCollector(c)
		if err != nil {
			return err
		}
		serveHttp(func(r prometheus.Registerer) {
			r.MustRegister(collector)
		})
		return nil
	},
}
