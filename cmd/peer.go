package cmd

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/irmf/reflector.go/db"
	"github.com/irmf/reflector.go/peer"
	"github.com/irmf/reflector.go/store"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var peerNoDB bool

func init() {
	var cmd = &cobra.Command{
		Use:   "peer",
		Short: "Run peer server",
		Run:   peerCmd,
	}
	cmd.Flags().BoolVar(&peerNoDB, "nodb", false, "Don't connect to a db and don't use a db-backed blob store")
	rootCmd.AddCommand(cmd)
}

func peerCmd(cmd *cobra.Command, args []string) {
	var err error

	s3 := store.NewS3Store(globalConfig.AwsID, globalConfig.AwsSecret, globalConfig.BucketRegion, globalConfig.BucketName)
	peerServer := peer.NewServer(s3)

	if !peerNoDB {
		db := new(db.SQL)
		err = db.Connect(globalConfig.DBConn)
		checkErr(err)

		combo := store.NewDBBackedStore(s3, db)
		peerServer = peer.NewServer(combo)
	}

	err = peerServer.Start(":" + strconv.Itoa(peer.DefaultPort))
	if err != nil {
		log.Fatal(err)
	}

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
	<-interruptChan
	peerServer.Shutdown()
}
