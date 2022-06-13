# ESPhome -> HomeKit Bridge

This library allows `esphome` device to be published and controlled over HomeKit. 

It required a small Linux server with local connection to `esphome` device. I'm using Raspberry Pi 3, but also older versions should work without any problem. One instance consumes about 15 MB of RAM. For every device, you have to run a separate instance that will publish a new HomeKit device.

## Fast Lane

- compile library for your architecture
- copy `esphome-homekit` executable on your server
- create `config.yaml` in the same directory – here is one example:
```yaml
name: mylight
address: 172.33.5.22:6053
password: myESPHomeAPIPassword

homekit:
  pin: "13062022"
  storage_dir: ./.homekit
```

- run `esphome-homekit` binary from the same directory

Application will create a new subdirectory and store HomeKit information there (private key, connections, etc...).

## What is supported?

This bridge is still in development phase and not all `esphome` features/types are not supported. Currently, supported types are:

- **Switch** - will create HomeKit switch (simple On/Off)
- **Binary Sensor** - will create Programmable Switch in HomeKit (single press will be mapped as On, double press as off). Using this, you can configure HomeKit devices to react on Binary Sensor from `esphome`
- **Fan** - will create Fan in HomeKit but only with On/Off support
- **Light** - will create Lightbulb in HomeKit. Only Brightness and On/Off is mapped
- **Sensor** with device class of `temperature` and `humidity` - will create Temperature or Humidity sensor in HomeKit

Will Always be created single accessory with multiple HomeKit services.

## Thanks to...
- [mycontroller-org/esphome_api](https://github.com/mycontroller-org/esphome_api) - `esphome` API library to connect with `esphome` device
- [brutella/hap](https://github.com/brutella/hap) - great library that makes possible creating HomeKit devices using Golang

## Contributions

Any kind of contributions/ideas are welcome.