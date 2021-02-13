package cmd

import (
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/irmf/reflector.go/db"
	"github.com/irmf/reflector.go/internal/metrics"
	"github.com/irmf/reflector.go/meta"
	"github.com/irmf/reflector.go/peer"
	"github.com/irmf/reflector.go/peer/http3"
	"github.com/irmf/reflector.go/reflector"
	"github.com/irmf/reflector.go/store"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

var (
	tcpPeerPort           int
	http3PeerPort         int
	receiverPort          int
	metricsPort           int
	disableUploads        bool
	disableBlocklist      bool
	proxyAddress          string
	proxyPort             string
	proxyProtocol         string
	useDB                 bool
	cloudFrontEndpoint    string
	reflectorCmdDiskCache string
	reflectorCmdMemCache  int
)

func init() {
	var cmd = &cobra.Command{
		Use:   "reflector",
		Short: "Run reflector server",
		Run:   reflectorCmd,
	}
	cmd.Flags().StringVar(&proxyAddress, "proxy-address", "", "address of another reflector server where blobs are fetched from")
	cmd.Flags().StringVar(&proxyPort, "proxy-port", "5567", "port of another reflector server where blobs are fetched from")
	cmd.Flags().StringVar(&proxyProtocol, "proxy-protocol", "http3", "protocol used to fetch blobs from another reflector server (tcp/http3)")
	cmd.Flags().StringVar(&cloudFrontEndpoint, "cloudfront-endpoint", "", "CloudFront edge endpoint for standard HTTP retrieval")
	cmd.Flags().IntVar(&tcpPeerPort, "tcp-peer-port", 5567, "The port reflector will distribute content from")
	cmd.Flags().IntVar(&http3PeerPort, "http3-peer-port", 5568, "The port reflector will distribute content from over HTTP3 protocol")
	cmd.Flags().IntVar(&receiverPort, "receiver-port", 5566, "The port reflector will receive content from")
	cmd.Flags().IntVar(&metricsPort, "metrics-port", 2112, "The port reflector will use for metrics")
	cmd.Flags().BoolVar(&disableUploads, "disable-uploads", false, "Disable uploads to this reflector server")
	cmd.Flags().BoolVar(&disableBlocklist, "disable-blocklist", false, "Disable blocklist watching/updating")
	cmd.Flags().BoolVar(&useDB, "use-db", true, "whether to connect to the reflector db or not")
	cmd.Flags().StringVar(&reflectorCmdDiskCache, "disk-cache", "",
		"enable disk cache, setting max size and path where to store blobs. format is 'MAX_BLOBS:CACHE_PATH'")
	cmd.Flags().IntVar(&reflectorCmdMemCache, "mem-cache", 0, "enable in-memory cache with a max size of this many blobs")
	rootCmd.AddCommand(cmd)
}

func reflectorCmd(cmd *cobra.Command, args []string) {
	log.Printf("reflector %s", meta.VersionString())

	// the blocklist logic requires the db backed store to be the outer-most store
	underlyingStore := setupStore()
	outerStore := wrapWithCache(underlyingStore)

	if !disableUploads {
		reflectorServer := reflector.NewServer(underlyingStore)
		reflectorServer.Timeout = 3 * time.Minute
		reflectorServer.EnableBlocklist = !disableBlocklist

		err := reflectorServer.Start(":" + strconv.Itoa(receiverPort))
		if err != nil {
			log.Fatal(err)
		}
		defer reflectorServer.Shutdown()
	}

	peerServer := peer.NewServer(outerStore)
	err := peerServer.Start(":" + strconv.Itoa(tcpPeerPort))
	if err != nil {
		log.Fatal(err)
	}
	defer peerServer.Shutdown()

	http3PeerServer := http3.NewServer(outerStore)
	err = http3PeerServer.Start(":" + strconv.Itoa(http3PeerPort))
	if err != nil {
		log.Fatal(err)
	}
	defer http3PeerServer.Shutdown()

	metricsServer := metrics.NewServer(":"+strconv.Itoa(metricsPort), "/metrics")
	metricsServer.Start()
	defer metricsServer.Shutdown()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
	<-interruptChan
	// deferred shutdowns happen now
}

func setupStore() store.BlobStore {
	var s store.BlobStore

	if proxyAddress != "" {
		switch proxyProtocol {
		case "tcp":
			s = peer.NewStore(peer.StoreOpts{
				Address: proxyAddress + ":" + proxyPort,
				Timeout: 30 * time.Second,
			})
		case "http3":
			s = http3.NewStore(http3.StoreOpts{
				Address: proxyAddress + ":" + proxyPort,
				Timeout: 30 * time.Second,
			})
		default:
			log.Fatalf("protocol is not recognized: %s", proxyProtocol)
		}
	} else {
		s3Store := store.NewS3Store(globalConfig.AwsID, globalConfig.AwsSecret, globalConfig.BucketRegion, globalConfig.BucketName)
		if cloudFrontEndpoint != "" {
			s = store.NewCloudFrontRWStore(store.NewCloudFrontROStore(cloudFrontEndpoint), s3Store)
		} else {
			s = s3Store
		}
	}

	if useDB {
		db := new(db.SQL)
		db.TrackAccessTime = true
		err := db.Connect(globalConfig.DBConn)
		if err != nil {
			log.Fatal(err)
		}

		s = store.NewDBBackedStore(s, db)
	}

	return s
}

func wrapWithCache(s store.BlobStore) store.BlobStore {
	wrapped := s

	diskCacheMaxSize, diskCachePath := diskCacheParams()
	if diskCacheMaxSize > 0 {
		err := os.MkdirAll(diskCachePath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		wrapped = store.NewCachingStore(
			"reflector",
			wrapped,
			store.NewLRUStore("peer_server", store.NewDiskStore(diskCachePath, 2), diskCacheMaxSize),
		)
	}

	if reflectorCmdMemCache > 0 {
		wrapped = store.NewCachingStore(
			"reflector",
			wrapped,
			store.NewLRUStore("peer_server", store.NewMemStore(), reflectorCmdMemCache),
		)
	}

	return wrapped
}

func diskCacheParams() (int, string) {
	if reflectorCmdDiskCache == "" {
		return 0, ""
	}

	parts := strings.Split(reflectorCmdDiskCache, ":")
	if len(parts) != 2 {
		log.Fatalf("--disk-cache must be a number, followed by ':', followed by a string")
	}

	maxSize := cast.ToInt(parts[0])
	if maxSize <= 0 {
		log.Fatalf("--disk-cache max size must be more than 0")
	}

	path := parts[1]
	if len(path) == 0 || path[0] != '/' {
		log.Fatalf("--disk-cache path must start with '/'")
	}

	return maxSize, path
}
