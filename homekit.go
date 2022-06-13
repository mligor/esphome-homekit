package esphomehomekit

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

func (s *svc) initializeHomeKit() (err error) {

	a, err := s.createAccessory()
	if err != nil {
		logrus.WithError(err).Error("unable to create homekit accessory")
		return
	}

	for _, e := range s.entities {
		svc, err := s.createService(e)
		if err != nil {
			logrus.WithError(err).Error("unable to create service")
			continue
		}
		if svc == nil {
			continue
		}
		svc.Id = uint64(e.Key)
		a.AddS(svc)
	}

	// log.Info.Enable()
	// log.Debug.Enable()

	// lightID := uint32(1935470627)
	// switchID := uint32(3202316517)
	// powerID := uint32(2391494160)

	go func() {

		// powerCo := characteristic.NewInt("1F10D5F1-01F5-470B-AE88-34E34EFA19B7")
		// powerCo.Format = characteristic.FormatUInt8
		// powerCo.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
		// powerCo.SetMinValue(0)
		// powerCo.SetMaxValue(100)
		// powerCo.SetStepValue(1)
		// powerCo.SetValue(0)
		// powerCo.Description = "Power Consumption"
		// powerCo.Unit = "W" //characteristic.UnitPercentage
		// powerCoLastValue := int(0)
		// powerCo.Id = 10001

		// batLevel := characteristic.NewBatteryLevel()
		// batLevel.Id = 10002

		// powerEntry := s.entities[uint32(powerID)]
		// powerEntry.OnUpdate = func(newState interface{}) {
		// 	s, ok := newState.(*api.SensorStateResponse)
		// 	if ok {
		// 		//newState := math.Floor(float64(s.State)*100) / 100
		// 		newState := int(math.Floor(float64(s.State)))

		// 		if powerCoLastValue != newState {
		// 			batLevel.SetValue(newState)

		// 			logrus.Debugf("new power consuption: %+v", newState)
		// 			powerCo.SetValue(newState)
		// 			powerCoLastValue = newState
		// 		}
		// 	} else {
		// 		logrus.Errorf("invalid power state: %+v", newState)
		// 	}
		// }

		// powerCo.OnValueRemoteUpdate(func(v int) {
		// 	logrus.Debugf("set new value : %v", v)
		// })

		// s.entities[uint32(powerID)] = powerEntry
		// a.Lightbulb.AddC(powerCo.C)
		// a.Lightbulb.AddC(batLevel.C)

		// service.NewSwitch()

		// sw := service.NewStatelessProgrammableSwitch()
		// sw.ProgrammableSwitchEvent.MaxVal = 1
		// a.AddS(sw.S)

		// // sw2 := service.NewSwitch()

		// // sw2.On.Permissions = []string{characteristic.PermissionEvents}
		// // a.AddS(sw2.S)
		// // sw2.On.SetValue(true)

		// //sw.Id = 2

		// fs := hap.NewFsStore("./hk")
		// lightEntity := s.entities[uint32(lightID)]

		// lightEntity.OnUpdate = func(newState interface{}) {
		// 	s, ok := newState.(*api.SwitchStateResponse)
		// 	if ok {
		// 		a.Lightbulb.On.SetValue(s.State)
		// 	} else {

		// 		logrus.Errorf("invalid state: %+v", newState)
		// 	}
		// }

		// s.entities[uint32(lightID)] = lightEntity

		// a.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		// 	if on {
		// 		logrus.Debug("Client changed light to on")

		// 	} else {
		// 		logrus.Debug("Client changed light to off")
		// 	}

		// 	err := s.esphomeClient.Send(&api.SwitchCommandRequest{
		// 		Key:   uint32(lightID),
		// 		State: on,
		// 	})

		// 	if err != nil {
		// 		logrus.WithError(err).Error("error sending request to homekit")
		// 	}

		// })

		// switchEntity := s.entities[uint32(switchID)]

		// switchEntity.OnUpdate = func(newState interface{}) {
		// 	s, ok := newState.(*api.BinarySensorStateResponse)
		// 	if ok {
		// 		//sw2.On.SetValue(s.State)
		// 		if s.State {
		// 			sw.ProgrammableSwitchEvent.SetValue(0)
		// 		} else {
		// 			sw.ProgrammableSwitchEvent.SetValue(1)

		// 		}
		// 	} else {
		// 		logrus.Errorf("invalid state for switch: %+v", newState)
		// 	}
		// }

		// test := accessory.NewAirPurifier(accessory.Info{
		// 	Name:         "Test",
		// 	Manufacturer: "mligor",
		// 	Model:        "esphome-homekit",
		// 	Firmware:     ver,
		// })

		// test.Id = 4

		// swingMode := characteristic.NewSwingMode()
		// test.AirPurifier.AddC(swingMode.C)
		// test.AirPurifier.AddC(characteristic.NewRotationSpeed().C)

		// door := accessory.New(accessory.Info{
		// 	Name:         "Door Bell",
		// 	Manufacturer: "mligor",
		// 	Model:        "esphome-homekit",
		// 	Firmware:     ver,
		// }, accessory.TypeVideoDoorbell)

		// door.Id = 6
		// doorBell := service.NewDoorbell()
		// doorBell.Primary = true
		// nameC := characteristic.NewName()
		// nameC.SetValue("Mladenovic")
		// doorBell.AddC(nameC.C)
		// doorBell.ProgrammableSwitchEvent.MinVal = 0
		// doorBell.ProgrammableSwitchEvent.MaxVal = 0
		// doorBell.Hidden = false

		// door.AddS(doorBell.S)
		// door.AddS(service.NewCameraRTPStreamManagement().S)
		// door.AddS(service.NewSpeaker().S)
		// door.AddS(service.NewMicrophone().S)

		// Create the hap server.
		fs := hap.NewFsStore(s.homekitStorageDir)
		server, err := hap.NewServer(fs, a)
		if err != nil {
			logrus.WithError(err).Fatal("unable to create homekit server")
		}

		server.Pin = s.homekitPIN

		// Setup a listener for interrupts and SIGTERM signals
		// to stop the server.
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-c
			// Stop delivering signals.
			signal.Stop(c)
			// Cancel the context to stop the server.
			cancel()
		}()

		// Run the server.
		server.ListenAndServe(ctx)

	}()

	return
}
