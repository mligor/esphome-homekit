package esphomehomekit

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	esphome "github.com/mycontroller-org/esphome_api/pkg/client"
	"github.com/mycontroller-org/esphome_api/pkg/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ESPHomeService interface {
	Start() error
}

type svc struct {
	entities          EntryMap
	name              string
	homekitPIN        string
	homekitStorageDir string
	esphomeInfo       *model.HelloResponse
	esphomeClient     *esphome.Client
	ctx               context.Context
	cancel            context.CancelFunc
	wg                *sync.WaitGroup
}

func New() ESPHomeService {
	return &svc{
		entities: make(EntryMap),
	}
}

func (s *svc) connectToESPHome(subscribeStates bool) (err error) {

	if s.esphomeClient != nil {
		s.esphomeClient = nil
	}

	address := viper.GetString("address")
	s.esphomeClient, err = esphome.Init(s.name, address, time.Second*10, s.esphomeHandler)
	if err != nil {
		logrus.WithError(err).Error("unable to init client")
		return
	}

	helloResponse, err := s.esphomeClient.Hello()
	if err != nil {
		logrus.WithError(err).Error("no answer from hello")
		return
	}
	logrus.Debugf("hello response : %v", helloResponse)

	password := viper.GetString("password")
	err = s.esphomeClient.Login(password)
	if err != nil {
		logrus.WithError(err).Error("unable to login to client")
		return
	}

	if subscribeStates {
		err = s.esphomeClient.SubscribeStates()
		if err != nil {
			logrus.WithError(err).Error("unable to subscribe for states")
			s.esphomeClient.Close()
		}
	} else {
		s.esphomeInfo = helloResponse
	}

	return
}

func (s *svc) Start() (err error) {

	viper.SetDefault("log_level", "warning")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	pflag.String("log_level", "warning", "Log level")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	err = viper.ReadInConfig()
	if err != nil {
		logrus.WithError(err).Fatal("unable to read config")
	}

	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	customFormatter.ForceColors = true
	logrus.SetFormatter(customFormatter)

	logLevel, err := logrus.ParseLevel(viper.GetString("log_level"))
	if err != nil {
		logrus.WithError(err).Errorf("wrong log_level text : %s", viper.GetString("log_level"))
	}
	logrus.WithField("log_level", logLevel).Print("Log level set")
	logrus.SetLevel(logLevel)

	s.name = viper.GetString("name")
	s.homekitPIN = viper.GetString("homekit.pin")
	s.homekitStorageDir = viper.GetString("homekit.storage_dir")
	if s.homekitStorageDir == "" {
		s.homekitStorageDir = "./.homekit"
	}

	// Setup a listener for interrupts and SIGTERM signals
	// to stop the server.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	s.wg = new(sync.WaitGroup)

	s.ctx, s.cancel = context.WithCancel(context.Background())

	err = s.connectToESPHome(false)
	if err != nil {
		logrus.WithError(err).Error("unable to connect to esphome")
		return
	}
	defer s.esphomeClient.Close()

	err = s.esphomeClient.ListEntities()
	if err != nil {
		logrus.WithError(err).Error("error when listing entries")
		return
	}

	pingTicker := time.NewTicker(15 * time.Second)
	defer pingTicker.Stop()

	go func() {
		errorCounter := 0
		for _ = range pingTicker.C {

			if s.esphomeClient != nil {
				logrus.Debug("pinging esphome")
				pingError := s.esphomeClient.Ping()
				if pingError != nil {
					logrus.WithError(pingError).Errorf("error pinging esphome")
					errorCounter++
				} else {
					errorCounter = 0
				}

				if errorCounter >= 2 {
					errorCounter = 0
					// Try to reconnect
					logrus.Debug("reconnecting esphome")
					connectError := s.connectToESPHome(true)
					if connectError != nil {
						logrus.WithError(connectError).Errorf("error connecting to esphome")
					}
				}
			}

		}
	}()

	<-c // block until we got interupt signal
	// Stop delivering signals.
	signal.Stop(c)
	// Cancel the context to stop the server.
	s.cancel()

	logrus.Debug("shuting down esphome-homekit bridge")

	s.wg.Wait()
	return
}
