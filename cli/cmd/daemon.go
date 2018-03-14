package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kinvolk/lerobot/pkg/daemon"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "responsible for requesting and renewing certificates",
	Run:   daemonRun,
}

func init() {
	RootCmd.AddCommand(daemonCmd)

	daemonCmd.Flags().String("authorized-keys-file", "", "Path to authorized_keys file to write (disabled when not set)")
	daemonCmd.Flags().String("le-config", "./lets-encrypt.yaml", "Path to lerobot's Let's Encrypt config file")
	daemonCmd.Flags().String("le-api", "https://acme-v01.api.letsencrypt.org/directory", "Let's Encrypt API URL")
	daemonCmd.Flags().String("account-dir", "./accounts", "Path to directory where to store account data")
	daemonCmd.Flags().String("certificate-dir", "./certificates", "Path to directory where to store certificate data")
	daemonCmd.Flags().Duration("interval-seconds", 300, "Seconds to sleep between doing work")

	viper.BindPFlags(daemonCmd.Flags())
}

func daemonRun(cmd *cobra.Command, args []string) {
	log.Printf("lerobot version %s", version)

	daemon, err := daemon.New(&daemon.Options{
		AuthorizedKeysPath: viper.GetString("authorized-keys-file"),
		LEConfigPath:       viper.GetString("le-config"),
		LEAPI:              viper.GetString("le-api"),
		AccountDir:         viper.GetString("account-dir"),
		CertificateDir:     viper.GetString("certificate-dir"),
		Interval:           viper.GetDuration("interval-seconds") * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to init daemon: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	log.Println("Running daemon ...")

	go daemon.Run()

	<-sigChan

	log.Println("Shutting down ...")
	log.Println("^C again to force stop")

	go func() {
		<-sigChan
		log.Fatalln("Aborted")
	}()

	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer timeoutCancel()

	if err := daemon.Shutdown(timeoutCtx); err != nil {
		log.Fatalf("Graceful shutdown failed: %v", err)
	}

	log.Println("Stopped")
}
