package esphomehomekit

import (
	"context"
	"fmt"
	"net/http"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/brutella/hap/service"
	"github.com/mycontroller-org/esphome_api/pkg/api"
	"github.com/sirupsen/logrus"
)

func (s *svc) createAccessory() (a *accessory.A, err error) {
	ver := fmt.Sprintf("%v.%v", s.esphomeInfo.ApiVersionMajor, s.esphomeInfo.ApiVersionMinor)

	a = accessory.New(accessory.Info{
		Name:         s.name,
		SerialNumber: s.esphomeInfo.ServerInfo,
		Manufacturer: "mligor",
		Model:        "esphome-homekit",
		Firmware:     ver,
	}, accessory.TypeOther) // TODO: choose the right type

	a.IdentifyFunc = func(r *http.Request) {
		logrus.Debug("identify") // TODO: do something usefull, maybe Ping
	}

	a.Id = 1
	return
}

func (s *svc) createSwichService(e *entity) (sv *service.S, err error) {

	k := service.NewSwitch()
	// k.On.Description = e.Name

	name := characteristic.NewName()
	name.SetValue(e.Name)
	k.AddC(name.C)

	// esphome -> homekit
	e.OnUpdate = func(newState interface{}) {
		msg, ok := newState.(*api.SwitchStateResponse)
		if ok {
			k.On.SetValue(msg.State)
		} else {
			logrus.Errorf("unexpected state for switch : %+v", newState)
		}
	}

	// homekit -> esphome
	k.On.OnSetRemoteValue(func(v bool) error {
		return s.esphomeClient.Send(&api.SwitchCommandRequest{
			Key:   e.Key,
			State: v,
		})
	})
	sv = k.S
	return
}

func (s *svc) createFanService(e *entity) (sv *service.S, err error) {

	//TODO: implement oscilating and speed
	k := service.NewFanV2()

	name := characteristic.NewName()
	name.SetValue(e.Name)
	k.AddC(name.C)

	// esphome -> homekit
	e.OnUpdate = func(newState interface{}) {
		msg, ok := newState.(*api.FanStateResponse)
		if ok {
			if msg.State {
				k.Active.SetValue(1)
			} else {
				k.Active.SetValue(0)
			}
		} else {
			logrus.Errorf("unexpected state for switch : %+v", newState)
		}
	}

	// homekit -> esphome
	k.Active.OnSetRemoteValue(func(v int) error {
		newState := v == 1

		return s.esphomeClient.Send(&api.FanCommandRequest{
			Key:      e.Key,
			State:    newState,
			HasState: true,
		})
	})
	sv = k.S
	return
}

func (s *svc) createLightService(e *entity) (sv *service.S, err error) {

	supportsBrightness := false
	//supportsRGB := false

	msg, ok := e.Info.(*api.ListEntitiesLightResponse)
	if ok {
		for _, v := range msg.SupportedColorModes {
			if v == api.ColorMode_COLOR_MODE_BRIGHTNESS {
				supportsBrightness = true
			}
			// else if v == api.ColorMode_COLOR_MODE_RGB {
			// 	supportsRGB = true
			// }
		}
	}

	k := service.NewLightbulb() // TODO: add support for RGBs

	name := characteristic.NewName()
	name.SetValue(e.Name)
	k.AddC(name.C)

	brightness := characteristic.NewBrightness()
	if supportsBrightness {
		k.AddC(brightness.C)
	}

	// esphome -> homekit
	e.OnUpdate = func(newState interface{}) {
		msg, ok := newState.(*api.LightStateResponse)
		if ok {
			k.On.SetValue(msg.State)

			if supportsBrightness {
				brightness.SetValue(int(msg.Brightness * 100))
			}

		} else {
			logrus.Errorf("unexpected state for light : %+v", newState)
		}

	}

	// homekit -> esphome
	k.On.OnSetRemoteValue(func(v bool) error {
		return s.esphomeClient.Send(&api.LightCommandRequest{
			Key:      e.Key,
			State:    v,
			HasState: true,
		})
	})

	brightness.OnSetRemoteValue(func(v int) error {
		return s.esphomeClient.Send(&api.LightCommandRequest{
			Key:           e.Key,
			Brightness:    float32(v) / 100.0,
			HasBrightness: true,
		})
	})

	sv = k.S
	return
}

func (s *svc) createProgrammableSwitchService(e *entity) (sv *service.S, err error) {

	k := service.NewStatelessProgrammableSwitch()
	// k.ProgrammableSwitchEvent.Description = e.Name
	k.ProgrammableSwitchEvent.MaxVal = 1 // allowed is only 0 and 1 (0 = single press (turn off), 1 = double (turn on))

	name := characteristic.NewName()
	name.SetValue(e.Name)
	k.AddC(name.C)

	// esphome -> homekit
	e.OnUpdate = func(newState interface{}) {
		msg, ok := newState.(*api.BinarySensorStateResponse)
		if ok {
			if msg.State {
				k.ProgrammableSwitchEvent.SetValue(0)
			} else {
				k.ProgrammableSwitchEvent.SetValue(1)
			}
		} else {
			logrus.Errorf("unexpected state for binary sensor : %+v", newState)
		}
	}

	// homekit -> esphome
	// nothing here, as homekit can not trigger this event
	sv = k.S
	return
}

func (s *svc) createTemperatureService(e *entity) (sv *service.S, err error) {

	k := service.NewTemperatureSensor()

	name := characteristic.NewName()
	name.SetValue(e.Name)
	k.AddC(name.C)

	// esphome -> homekit
	e.OnUpdate = func(newState interface{}) {
		msg, ok := newState.(*api.SensorStateResponse)
		if ok {

			k.CurrentTemperature.SetValue(float64(msg.State))
		} else {
			logrus.Errorf("unexpected state for sensor : %+v", newState)
		}
	}

	// homekit -> esphome
	// nothing here, as homekit can not trigger this event
	sv = k.S
	return
}

func (s *svc) createHumidityService(e *entity) (sv *service.S, err error) {
	k := service.NewHumiditySensor()

	name := characteristic.NewName()
	name.SetValue(e.Name)
	k.AddC(name.C)

	// esphome -> homekit
	e.OnUpdate = func(newState interface{}) {
		msg, ok := newState.(*api.SensorStateResponse)
		if ok {

			k.CurrentRelativeHumidity.SetValue(float64(msg.State))
		} else {
			logrus.Errorf("unexpected state for sensor : %+v", newState)
		}
	}

	// homekit -> esphome
	// nothing here, as homekit can not trigger this event
	sv = k.S
	return
}

func (s *svc) createSensorService(e *entity) (sv *service.S, err error) {

	msg, ok := e.Info.(*api.ListEntitiesSensorResponse)
	if ok {

		switch msg.DeviceClass {
		case "temperature":
			return s.createTemperatureService(e)
		case "humidity":
			return s.createHumidityService(e)
			//TODO: implement other device classes
		}
	}
	return
}

func (s *svc) createService(e *entity) (sv *service.S, err error) {

	switch e.Type {
	case EntityTypeSwitch:
		return s.createSwichService(e)
	case EntityTypeBinarySensor:
		return s.createProgrammableSwitchService(e)
	case EntityTypeFan:
		return s.createFanService(e)
	case EntityTypeLight:
		return s.createLightService(e)
	case EntityTypeSensor:
		return s.createSensorService(e)

		//TODO: implement other types
	}

	return
}

func (s *svc) initializeHomeKit(ctx context.Context) (err error) {

	a, err := s.createAccessory()
	if err != nil {
		logrus.WithError(err).Error("unable to create homekit accessory")
		return
	}

	entities := s.entities.sorted()

	for _, e := range entities {
		svc, err := s.createService(e)
		if err != nil {
			logrus.WithError(err).Error("unable to create service")
			continue
		}
		if svc == nil {
			continue
		}
		logrus.WithField("svc", svc.Type).Debug("added new service")
		a.AddS(svc)
	}

	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		logrus.Debug("starting homekit server")

		// Create the hap server.
		fs := hap.NewFsStore(s.homekitStorageDir)
		server, err := hap.NewServer(fs, a)
		if err != nil {
			logrus.WithError(err).Fatal("unable to create homekit server")
		}

		server.Pin = s.homekitPIN

		// Run the server.
		server.ListenAndServe(ctx)

		logrus.Debug("finishing homekit server")

	}()

	return
}
