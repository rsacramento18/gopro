package bleController

import (
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

const (
	BLE_SERVICE_GOPRO                         = 0xFEA6
	BLE_CHARACTERISTIC_GOPRO_COMMAND          = "b5f90072-aa8d-11e3-9046-0002a5d5c51b"
	BLE_CHARACTERISTIC_GOPRO_COMMAND_RESPONSE = "b5f90073-aa8d-11e3-9046-0002a5d5c51b"
)

type Controller struct {
	Adapter bluetooth.Adapter
	Scans   []bluetooth.Address
}

func NewController() Controller {
	adapter := bluetooth.DefaultAdapter
	must("enable BLE stack", adapter.Enable())
	return Controller{*adapter, []bluetooth.Address{}}
}

func (controller *Controller) Scan() {
	fmt.Println("scanning...")
	count := 0
	err := controller.Adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if device.HasServiceUUID(bluetooth.New16BitUUID(BLE_SERVICE_GOPRO)) {
			println("found device:", device.Address.String(), device.LocalName())
			controller.Scans = append(controller.Scans, device.Address)
		}
		count++
		time.Sleep(1 * time.Second)
		if count == 5 {
			controller.Adapter.StopScan()
		}
	})
	must("start scan", err)
	fmt.Println("finished scanning")
}

func (controller Controller) PrintScans() {
	for index, node := range controller.Scans {
		fmt.Println(index, "-", node.String())
	}
}

func (controller Controller) Connect(address bluetooth.Address) bluetooth.Device {
	res, err := controller.Adapter.Connect(address, bluetooth.ConnectionParams{})
	go func() {
		if err != nil {
			println("error connection:", err.Error())
			return
		}
	}()

	return res
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
