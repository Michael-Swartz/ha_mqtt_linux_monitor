package main

import (
	"fmt"
	"os"
	"time"
	"strconv"
	"syscall"

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

type DiskStats struct {
	All uint64
	Free uint64
	Used uint64
	Used_Percent float64
}

func GetDiskUsage(path string) (disk DiskStats) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}

	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free

	disk.Used_Percent = float64(disk.Used)/float64(disk.All)
	return
}

func FloatToString(input float64) string {
	return strconv.FormatFloat(input, 'f', 2, 64)
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

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
		fmt.Printf("cpu total: %s%%\n", FloatToString(total_cpu_usage))
		token := client.Publish("/test/cpu", 0, false, FloatToString(total_cpu_usage))
		token.Wait()
		memory, err := memory.Get()

		if err != nil {
			fmt.Fprint(os.Stderr, "%s\n", err)
		}
		total_mem_usage := (float64(memory.Used)/float64(memory.Total)*100)
		fmt.Printf("memory total use: %s%%\n", FloatToString(total_mem_usage))
		token_mem := client.Publish("/test/memory", 0, false, FloatToString(total_mem_usage))
		token_mem.Wait()

		disk := GetDiskUsage("/")
		total_disk_usage := (float64(disk.Used_Percent)*100)
		fmt.Printf("Disk Usage: %s%%\n", FloatToString(total_disk_usage))
		token_disk := client.Publish("/test/disk", 0, false, FloatToString(total_disk_usage))
		token_disk.Wait()
		time.Sleep(time.Duration(1) * time.Second)

	}
}
