package main

import (
	"fmt"
	"os"
	"time"
	"strconv"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	broker = "10.0.0.170"
	port   = "1883"

	connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		fmt.Println("Connected")
	}
)

func FloatToString(input float64) string {
	return strconv.FormatFloat(input, 'f', 2, 64)
}

func main() {

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://10.0.0.170:1883"))
	opts.OnConnect = connectHandler
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(fmt.Sprintf("Error connecting to MQTT broker:", token.Error()))
	}

	for {
		before, err := cpu.Get()
		if err != nil {
			fmt.Fprint(os.Stderr, "%s\n", err)
		}
		time.Sleep(time.Duration(1) * time.Second)
		after, err := cpu.Get()

		if err != nil {
			fmt.Fprint(os.Stderr, "%s\n", err)
		}
		total := float64(after.Total - before.Total)
		total_cpu_usage := (float64(float64(after.System-before.System)+float64(after.User-before.User)) / total * 100)
		//fmt.Printf("cpu total: %f%%\n", (float64(float64(after.System-before.System)+float64(after.User-before.User)) / total * 100))
		fmt.Printf("cpu total: %s%%\n", FloatToString(total_cpu_usage))
		token := client.Publish("/test/cpu", 0, false, FloatToString(total_cpu_usage))
		token.Wait()
		memory, err := memory.Get()

		if err != nil {
			fmt.Fprint(os.Stderr, "%s\n", err)
		}
		total_mem_usage := (float64(memory.Used)/float64(memory.Total)*100)
		fmt.Printf("memory total use: %s%%\n", FloatToString(total_mem_usage))
		tokenn := client.Publish("/test/memory", 0, false, FloatToString(total_mem_usage))
		tokenn.Wait()
		time.Sleep(time.Duration(1) * time.Second)
	}
}
