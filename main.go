package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tkanos/konsumerou"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/vladimir-klymniuk/notification-service-original/config"
	"github.com/vladimir-klymniuk/notification-service-original/httpget"
	"github.com/vladimir-klymniuk/notification-service-original/kafka"
	"github.com/vladimir-klymniuk/notification-service-original/message"
	"github.com/vladimir-klymniuk/notification-service-original/metrics"
	"github.com/vladimir-klymniuk/notification-service-original/notify"
	"github.com/vladimir-klymniuk/notification-service-original/producer"
	"github.com/vladimir-klymniuk/notification-service-original/runner"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cfg := config.GetConfig()
	loadOSArgs(cfg)

	// zerolog.TimeFieldFormat = zerolog.TimeFieldFormat

	zerolog.SetGlobalLevel(zerolog.Level(cfg.Log.Level))

	log.Info().Msg(fmt.Sprintf("starting %s port %d", appName, cfg.App.Port))

	httpAddr := ":" + strconv.Itoa(cfg.App.Port)
	mux := http.NewServeMux()

	if cfg.App.EnablePprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	ctx := context.Background()

	sconfig := sarama.NewConfig()
	sconfig.Version = sarama.V2_4_0_0
	sconfig.Consumer.Offsets.CommitInterval = time.Second

	if cfg.Kafka.UseCredentials {
		sconfig.Net.SASL.Enable = cfg.Kafka.UseCredentials
		sconfig.Net.SASL.User = cfg.Kafka.Username
		sconfig.Net.SASL.Password = cfg.Kafka.Password
		sconfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		sconfig.Net.SASL.Handshake = true
		sconfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &kafka.XDGSCRAMClient{HashGeneratorFcn: kafka.SHA512} }
		sconfig.Producer.RequiredAcks = sarama.WaitForLocal
		sconfig.Producer.Compression = sarama.CompressionSnappy
		// sconfig.Producer.Flush.Frequency = 50 * time.Millisecond
	}

	service := cfg.Services[0]

	log.Info().Msgf("service: %v", service)

	dspErr, err := producer.NewPublisher(service.Topic, service.Error, cfg.Kafka.Brokers, sconfig)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("error creating kafka producer: %v", err))
	}

	// metrics
	dspErr = metrics.NewPublisher(dspErr, service.Error, service.Name)

	// message decoder
	decoder := message.NewDecoder()

	rb := runner.NewBuilder(service.Retry, service.RetryDelay)
	mrb := metrics.NewRunnerBuilder(rb, service.Name)
	// TODO: here create 3 listeners
	handler := httpget.MakeWorkerEndpoint(httpget.NewWorker(
		getHttpClient(service.Timeout),
		dspErr,
		decoder,
		service.MaxRequests,
		mrb,
	))

	listener, err := konsumerou.NewListener(ctx, // ultimately we should start one listener per service
		cfg.Kafka.Brokers, // kafka brokers
		service.GroupID,   // group id
		service.Topic,     // the topic name
		metrics.NewMetricssWorker(service.Name, handler), // the handler
		sconfig)

	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("listener not starting, %v", err))
	}
	// Subscribe your service to the topic
	listener.Subscribe()
	defer listener.Close()

	bsp, err := producer.NewPublisher("", service.Topic, cfg.Kafka.Brokers, sconfig)
	if err != nil {
		log.Error().Err(err).Msg("error creating kafka producer")
	}

	// message encoder
	enc := message.NewEncoder()

	// metrics
	bsp = metrics.NewPublisher(bsp, service.Topic, service.Name)

	bs := notify.NewService(bsp, enc)

	notifyEndpoint := notify.NewEndpoints(bs)
	notifyHandler := notify.NewHTTPHandler(notifyEndpoint).ServeHTTP
	mux.HandleFunc("/notify", metrics.NewHTTPMiddleware("notify", notifyHandler))

	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/removelb", removeLBHandler)
	mux.Handle("/metrics", promhttp.Handler())

	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	log.Info().Err(httpServer.ListenAndServe()).Msg(fmt.Sprintf("exit %s", appName))
}

var (
	appName    = "notification-service"
	version    = ""
	gitversion = ""
	buildtime  = ""
)

func loadOSArgs(_ *config.Configuration) {
	if len(os.Args) == 2 {
		if os.Args[1] == "version" {
			fmt.Printf("%s %s\n", appName, version)
			fmt.Printf("git commit : %s\n", gitversion)
			fmt.Printf("build time : %s\n", buildtime)
			os.Exit(0)
		}
	}
}

var status = http.StatusOK

// HealthzHandler returns HTTP Status 200 when the application is running with no issues
func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(status)
}

// RemoveLBHandler forces the HealthzHandler to return 403 on all subsequent requests so
// the loadbalancer to removes this server from its available pool.
func removeLBHandler(w http.ResponseWriter, _ *http.Request) {
	status = http.StatusForbidden
	w.WriteHeader(http.StatusOK)
}

func getHttpClient(timeout time.Duration) *http.Client {
	netTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   time.Second,
			KeepAlive: 0,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		IdleConnTimeout:     0,
		MaxIdleConnsPerHost: 50000,
		MaxIdleConns:        50000,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: netTransport,
	}
}
