package main

import (
	"fmt"

	"tinygo.org/x/bluetooth"
)

const (
	BLE_SERVICE_GOPRO                         = 0xFEA6
	BLE_CHARACTERISTIC_GOPRO_COMMAND          = "b5f90072-aa8d-11e3-9046-0002a5d5c51b"
	BLE_CHARACTERISTIC_GOPRO_COMMAND_RESPONSE = "b5f90073-aa8d-11e3-9046-0002a5d5c51b"
)

var adapter = bluetooth.DefaultAdapter

func main() {
	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	// Start scanning.
	println("scanning...")
	err := adapter.Scan(onScan)
	must("start scan", err)
}

func onScan(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	if device.HasServiceUUID(bluetooth.New16BitUUID(BLE_SERVICE_GOPRO)) {
		println("found device:", device.Address.String(), device.RSSI, device.LocalName())

		go func() {
			res, err := adapter.Connect(device.Address, bluetooth.ConnectionParams{})
			if err != nil {
				println("error connection:", err.Error())
				return
			}
			onConnect(device, res)
		}()
	}
}

func onConnect(scanResult bluetooth.ScanResult, device bluetooth.Device) {
	fmt.Println("connected:", scanResult.Address.String(), scanResult.LocalName())

	// Get a list of services
	services, err := device.DiscoverServices([]bluetooth.UUID{
		bluetooth.New16BitUUID(BLE_SERVICE_GOPRO),
	})
	if err != nil {
		fmt.Println("error getting services:", err.Error())
		return
	}

	for _, service := range services {
		if service.UUID() == bluetooth.New16BitUUID(BLE_SERVICE_GOPRO) {
			sendCommand, _ := bluetooth.ParseUUID(BLE_CHARACTERISTIC_GOPRO_COMMAND)
			commandRes, _ := bluetooth.ParseUUID(BLE_CHARACTERISTIC_GOPRO_COMMAND_RESPONSE)
			characteristics, err := service.DiscoverCharacteristics([]bluetooth.UUID{
				sendCommand, commandRes,
			})
			if err != nil {
				fmt.Println("error getting characteristics:", err.Error())
				return
			}

			for _, characteristic := range characteristics {
				if characteristic.UUID() == sendCommand {
					shutterOn := []byte{0x03, 0x01, 0x01, 0x01}
					_, err := characteristic.WriteWithoutResponse(shutterOn)
					if err != nil {
						fmt.Println("send command", err.Error())
						return
					}
				}
				if characteristic.UUID() == commandRes {
					err := characteristic.EnableNotifications(notification)
					if err != nil {
						fmt.Println("error enabling notifications:", err.Error())
						return
					}
				}
			}
		}
	}
}

func notification(buf []byte) {
	fmt.Println("received: %x", buf)
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}

