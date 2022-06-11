package esphomehomekit

import (
	"os"
	"os/signal"
	"strings"
	"time"

	esphome "github.com/mycontroller-org/esphome_api/pkg/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
)

type ESPHomeService interface {
	Start() error
}

type service struct{}

func New() ESPHomeService {
	return &service{}
}

func (s *service) Start() (err error) {

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

	name := viper.GetString("name")
	address := viper.GetString("address")
	password := viper.GetString("password")

	client, err := esphome.Init(name, address, time.Second*10, func(m proto.Message) {

		logrus.Debugf("message received : %v", m)

		//TODO;
	})
	if err != nil {
		logrus.WithError(err).Error("unable to init client")
		return
	}

	defer client.Close()

	helloResponse, err := client.Hello()
	if err != nil {
		logrus.WithError(err).Error("no answer from hello")
		return
	}

	err = client.Login(password)
	if err != nil {
		logrus.WithError(err).Error("unable to login to client")
		return
	}

	logrus.Debugf("hello response : %v", helloResponse)

	err = client.ListEntities()
	if err != nil {
		logrus.WithError(err).Error("error when listing entries")
		return
	}

	// err = client.SubscribeStates()
	// if err != nil {
	// 	logrus.WithError(err).Error("error when subscribing states")
	// 	return
	// }

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c // block until we got interupt signal

	return
}
