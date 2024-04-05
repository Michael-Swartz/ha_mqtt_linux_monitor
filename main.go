package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
)

var (
	broker = "10.0.0.170"
	port   = "1883"

	connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		fmt.Println("Connected")
	}
)

type DiskStats struct {
	All          uint64
	Free         uint64
	Used         uint64
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

	disk.Used_Percent = (float64(disk.Used) / float64(disk.All)) * 100
	return disk
}

func FloatToString(input float64) string {
	return strconv.FormatFloat(input, 'f', 2, 64)
}

func GetCPUUsage() string {
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
	fmt.Printf("CPU total use: %s%%\n", FloatToString(total_cpu_usage))

	return FloatToString(total_cpu_usage)
}

func GetMemoryUsage() string {
	memory, err := memory.Get()
	if err != nil {
		fmt.Fprint(os.Stderr, "%s\n", err)
	}
	total_mem_usage := (float64(memory.Used) / float64(memory.Total) * 100)
	fmt.Printf("memory total use: %s%%\n", FloatToString(total_mem_usage))

	return FloatToString(total_mem_usage)
}

func PublishMessage(channel, message string, client mqtt.Client) {
	token := client.Publish(channel, 0, false, message)
	token.Wait()
}

// This essentially just pareses the `sensor` command
func GetTemps(resource string) string {
	cmd := fmt.Sprintf("sensors | grep %s | sed 's/.*+//' | sed 's/°.*//'", resource)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Print("Error running shell command: ", err)
	}
	fmt.Printf("%s TEMP: %s", resource, out)
	return strings.TrimSuffix(string(out[:]), "\n")
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
		cpu_uage := GetCPUUsage()
		PublishMessage("/test/cpu", cpu_uage, client)

		mem_usage := GetMemoryUsage()
		PublishMessage("/test/memory", mem_usage, client)

		disk := GetDiskUsage("/")
		PublishMessage("/test/disk", FloatToString(disk.Used_Percent), client)

		cpu_temp := GetTemps("CPU")
		PublishMessage("/test/temp/cpu", cpu_temp, client)

		gpu_temp := GetTemps("GPU")
		PublishMessage("/test/temp/GPU", gpu_temp, client)

		//time.Sleep(time.Duration(1) * time.Second)

	}
}
