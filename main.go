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
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	logLevel, err := logrus.ParseLevel(viper.GetString("log_level"))
	if err != nil {
		logrus.WithError(err).Errorf("wrong log_level text : %s", viper.GetString("log_level"))
	}
	logrus.WithField("log_level", logLevel).Print("Log level set")
	logrus.SetLevel(logLevel)

	s.name = viper.GetString("name")
	s.homekitPIN = viper.GetString("homekit.pin")
	address := viper.GetString("address")
	password := viper.GetString("password")
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

	s.esphomeClient, err = esphome.Init(s.name, address, time.Second*10, s.esphomeHandler)
	if err != nil {
		logrus.WithError(err).Error("unable to init client")
		return
	}

	defer s.esphomeClient.Close()

	helloResponse, err := s.esphomeClient.Hello()
	if err != nil {
		logrus.WithError(err).Error("no answer from hello")
		return
	}

	err = s.esphomeClient.Login(password)
	if err != nil {
		logrus.WithError(err).Error("unable to login to client")
		return
	}

	s.esphomeInfo = helloResponse

	logrus.Debugf("hello response : %v", helloResponse)

	err = s.esphomeClient.ListEntities()
	if err != nil {
		logrus.WithError(err).Error("error when listing entries")
		return
	}

	<-c // block until we got interupt signal
	// Stop delivering signals.
	signal.Stop(c)
	// Cancel the context to stop the server.
	s.cancel()

	logrus.Debug("shuting down esphome-homekit bridge")

	s.wg.Wait()
	return
}
