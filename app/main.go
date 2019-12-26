// jobs-queue is the service for running jobs in background with queue limiter.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/hashicorp/logutils"
	"github.com/tierpod/jobs-queue/app/cache"
	"github.com/tierpod/jobs-queue/app/config"
	"github.com/tierpod/jobs-queue/app/worker"
)

const bufSize = 1024

const usage = `Run jobs in background with queue limiter.

Put job to queue:
  printf 'sleep 10' | nc /path/to/socket -Uu

Usage:
`

// Set version on compile time: -ldflags "-X main.version=0.1-git6e38624"
var version = "unset"

func main() {
	// Command line flags
	var (
		flagConfig  string
		flagVersion bool
	)

	flag.Usage = func() {
		fmt.Printf(usage)
		flag.PrintDefaults()
	}

	flag.StringVar(&flagConfig, "config", "./config.yaml", "path to config file")
	flag.BoolVar(&flagVersion, "version", false, "show version and exit")
	flag.Parse()

	if flagVersion {
		fmt.Printf("Version: %v\n", version)
		return
	}

	// load service configuration
	cfg, err := config.Load(flagConfig)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[INFO]  loaded config from %v (%d jobs)", flagConfig, len(cfg.Jobs))

	// configure logger
	setupLog(cfg.LogDebug, cfg.LogDatetime)

	// configure cache
	inCache, err := cache.New(cfg.CacheExpire, cfg.CacheDeleteMode, cfg.CacheExcludesRe)
	if err != nil {
		log.Fatal(err)
	}

	// start server
	log.Printf("[INFO]  start server on: %v, version: %v", cfg.Socket, version)

	quit := make(chan bool)
	listen, err := net.ListenPacket("unixgram", cfg.Socket)
	if err != nil {
		log.Fatalf("[ERROR] listen error: %v", err)
	}
	defer listen.Close()

	// configure signal handler
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		log.Printf("[INFO]  shutdown (signal: %v)", sig)
		close(quit)
		listen.Close()
		os.Remove(cfg.Socket)
		os.Exit(0)
	}()

	inCh := make(chan worker.Command, cfg.QueueSize)
	for i := 1; i <= cfg.Workers; i++ {
		worker := worker.New(inCh, inCache, "worker-"+strconv.Itoa(i))
		go worker.Start()
	}

	for {
		handleConnection(listen, inCh, inCache, quit, cfg)
	}
}

func handleConnection(conn net.PacketConn, inCh chan<- worker.Command, inCache *cache.Cache, quitCh chan bool, cfg *config.Config) {
	buf := make([]byte, bufSize)

	n, _, err := conn.ReadFrom(buf)
	if err != nil {
		select {
		case <-quitCh:
			return
		default:
		}
		log.Fatalf("[ERROR] %v", err)
	}

	cmdline := strings.TrimSpace(string(buf[:n]))
	log.Printf("[DEBUG] receive string: %v", cmdline)

	cmd, err := worker.NewCommand(cmdline, cfg)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		return
	}

	if inCache.Contains(cmd.Key()) {
		log.Printf("[INFO]  skip cmd: already in cache: %v", cmd)
		return
	}

	select {
	case inCh <- cmd:
	default:
		log.Printf("[WARN] queue limit is reached, drop job")
	}
}

func setupLog(debug, datetime bool) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stdout,
	}

	if debug {
		filter.MinLevel = logutils.LogLevel("DEBUG")
	}

	if datetime {
		log.SetFlags(log.LstdFlags)
	} else {
		log.SetFlags(0)
	}

	log.SetOutput(filter)
}
