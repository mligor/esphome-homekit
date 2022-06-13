package esphomehomekit

import (
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/mycontroller-org/esphome_api/pkg/api"
	esphome "github.com/mycontroller-org/esphome_api/pkg/client"
	"github.com/mycontroller-org/esphome_api/pkg/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
)

type ESPHomeService interface {
	Start() error
}

type svc struct {
	entities          map[uint32]*entity
	name              string
	homekitPIN        string
	homekitStorageDir string
	esphomeInfo       *model.HelloResponse
	esphomeClient     *esphome.Client
}

type entity struct {
	Key       uint32
	ID        string
	Name      string
	Type      EntityType
	Info      interface{}
	LastState interface{}
	OnUpdate  func(newState interface{})
}

type EntityType int

const (
	EntityTypeUnknown EntityType = iota
	EntityTypeBinarySensor
	EntityTypeCover
	EntityTypeFan
	EntityTypeLight
	EntityTypeSensor
	EntityTypeSwitch
	EntityTypeTextSensor
	EntityTypeCamera
	EntityTypeClimate
	EntityTypeNumber
	EntityTypeSelect
	EntityTypeLock
	EntityTypeButton
	EntityTypeMediaPlayer
)

func New() ESPHomeService {
	return &svc{
		entities: make(map[uint32]*entity),
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

	s.esphomeClient, err = esphome.Init(s.name, address, time.Second*10, func(m proto.Message) {

		logrus.Debugf("message received : %+v", m)

		switch api.TypeID(m) {

		// List response

		case api.ListEntitiesBinarySensorResponseTypeID:
			msg := m.(*api.ListEntitiesBinarySensorResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeBinarySensor,
				Info: msg,
			}

		case api.ListEntitiesCoverResponseTypeID:
			msg := m.(*api.ListEntitiesCoverResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeCover,
				Info: msg,
			}

		case api.ListEntitiesFanResponseTypeID:
			msg := m.(*api.ListEntitiesFanResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeFan,
				Info: msg,
			}

		case api.ListEntitiesLightResponseTypeID:
			msg := m.(*api.ListEntitiesLightResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeLight,
				Info: msg,
			}

		case api.ListEntitiesSensorResponseTypeID:
			msg := m.(*api.ListEntitiesSensorResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeSensor,
				Info: msg,
			}

		case api.ListEntitiesSwitchResponseTypeID:
			msg := m.(*api.ListEntitiesSwitchResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeSwitch,
				Info: msg,
			}

		case api.ListEntitiesTextSensorResponseTypeID:
			msg := m.(*api.ListEntitiesTextSensorResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeTextSensor,
				Info: msg,
			}

		case api.ListEntitiesCameraResponseTypeID:
			msg := m.(*api.ListEntitiesCameraResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeCamera,
				Info: msg,
			}

		case api.ListEntitiesClimateResponseTypeID:
			msg := m.(*api.ListEntitiesClimateResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeClimate,
				Info: msg,
			}

		case api.ListEntitiesNumberResponseTypeID:
			msg := m.(*api.ListEntitiesNumberResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeNumber,
				Info: msg,
			}

		case api.ListEntitiesSelectResponseTypeID:
			msg := m.(*api.ListEntitiesSelectResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeSelect,
				Info: msg,
			}

		case api.ListEntitiesLockResponseTypeID:
			msg := m.(*api.ListEntitiesLockResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeLock,
				Info: msg,
			}

		case api.ListEntitiesButtonResponseTypeID:
			msg := m.(*api.ListEntitiesButtonResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeButton,
				Info: msg,
			}

		case api.ListEntitiesMediaPlayerResponseTypeID:
			msg := m.(*api.ListEntitiesMediaPlayerResponse)
			s.entities[msg.Key] = &entity{
				Key:  msg.Key,
				ID:   msg.ObjectId,
				Name: msg.Name,
				Type: EntityTypeMediaPlayer,
				Info: msg,
			}

		// List Done

		case api.ListEntitiesDoneResponseTypeID:
			{
				logrus.Tracef("entities: %+v", s.entities)
				logrus.Debug("start subscribe for states")

				err := s.initializeHomeKit()
				if err != nil {
					logrus.WithError(err).Error("unable to initialize homekit")
				}

				err = s.esphomeClient.SubscribeStates()
				if err != nil {
					logrus.WithError(err).Error("unable to subscribe for states")
				}
			}

		// States

		case api.BinarySensorStateResponseTypeID:
			msg := m.(*api.BinarySensorStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			if entity.OnUpdate != nil {
				entity.OnUpdate(msg)
			}
			s.entities[msg.Key] = entity

		case api.CoverStateResponseTypeID:
			msg := m.(*api.CoverStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			if entity.OnUpdate != nil {
				entity.OnUpdate(msg)
			}
			s.entities[msg.Key] = entity

		case api.FanStateResponseTypeID:
			msg := m.(*api.FanStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			if entity.OnUpdate != nil {
				entity.OnUpdate(msg)
			}
			s.entities[msg.Key] = entity

		case api.LightStateResponseTypeID:
			msg := m.(*api.LightStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			if entity.OnUpdate != nil {
				entity.OnUpdate(msg)
			}
			s.entities[msg.Key] = entity

		case api.SensorStateResponseTypeID:
			msg := m.(*api.SensorStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			if entity.OnUpdate != nil {
				entity.OnUpdate(msg)
			}
			s.entities[msg.Key] = entity

		case api.SwitchStateResponseTypeID:
			msg := m.(*api.SwitchStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			if entity.OnUpdate != nil {
				entity.OnUpdate(msg)
			}
			s.entities[msg.Key] = entity

		case api.TextSensorStateResponseTypeID:
			msg := m.(*api.TextSensorStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			if entity.OnUpdate != nil {
				entity.OnUpdate(msg)
			}
			s.entities[msg.Key] = entity

		case api.ClimateStateResponseTypeID:
			msg := m.(*api.ClimateStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			s.entities[msg.Key] = entity

		case api.NumberStateResponseTypeID:
			msg := m.(*api.NumberStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			if entity.OnUpdate != nil {
				entity.OnUpdate(msg)
			}
			s.entities[msg.Key] = entity

		case api.SelectStateResponseTypeID:
			msg := m.(*api.SelectStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			if entity.OnUpdate != nil {
				entity.OnUpdate(msg)
			}
			s.entities[msg.Key] = entity

		case api.LockStateResponseTypeID:
			msg := m.(*api.LockStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			if entity.OnUpdate != nil {
				entity.OnUpdate(msg)
			}
			s.entities[msg.Key] = entity

		case api.MediaPlayerStateResponseTypeID:
			msg := m.(*api.MediaPlayerStateResponse)
			entity, ok := s.entities[msg.Key]
			if !ok {
				logrus.Errorf("received state for unknown key: %s", msg.Key)
				break
			}
			entity.LastState = msg
			if entity.OnUpdate != nil {
				entity.OnUpdate(msg)
			}
			s.entities[msg.Key] = entity

		}

	})
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
