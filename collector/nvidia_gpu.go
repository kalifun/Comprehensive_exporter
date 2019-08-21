package collector

import (
	"fmt"
	"github.com/mindprince/gonvml"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"strconv"
	"time"
)

type nvidaCollector struct {
	deviceCount *prometheus.Desc
	deviceInfo  *prometheus.Desc
	memoryUsed  *prometheus.Desc
	memoryTotal  *prometheus.Desc
	powerUsage  *prometheus.Desc
	temperature *prometheus.Desc
	processInfo *prometheus.Desc
	utilizationGPU   *prometheus.Desc
	utilizationMemory *prometheus.Desc
	utilizationGPUAverage *prometheus.Desc
}

const (
	space = "nvidia"
)

var (
	labels = []string{"minor_number", "uuid", "name"}
	averageDuration = 10 * time.Second
)

func init() {
	registerCollector("nvidia",defaultEnabled,NewCollector)
}

func NewCollector() (Collector,error) {
	return &nvidaCollector{
		deviceCount: prometheus.NewDesc(
			prometheus.BuildFQName(space, "", "device_count"),
			"Number of GPU devices.",
			nil, nil,
		),
		deviceInfo:  prometheus.NewDesc(
			prometheus.BuildFQName(space, "", "driver_info"),
			"driver_info.",
			[]string{"minor_number", "uuid", "name","version"}, nil,
		),
		memoryUsed: prometheus.NewDesc(
			prometheus.BuildFQName(space, "", "memory_used"),
			"Memory used by the GPU device in bytes.",
			labels, nil,
		),
		memoryTotal: prometheus.NewDesc(
			prometheus.BuildFQName(space, "", "memory_total"),
			"Total memory of the GPU device in bytes.",
			labels, nil,
		),
		powerUsage: prometheus.NewDesc(
			prometheus.BuildFQName(space, "", "power_usage"),
			"Power usage of the GPU device in milliwatts.",
			labels, nil,
		),
		temperature: prometheus.NewDesc(
			prometheus.BuildFQName(space, "", "temperatures"),
			"Temperature of the GPU device in celsius.",
			labels, nil,
		),
		processInfo: prometheus.NewDesc(
			prometheus.BuildFQName(space, "", "processInfo"),
			"Structure to store utilization value and process Id.",
			[]string{"minor_number", "Pid", "name"}, nil,
		),
		utilizationGPU: prometheus.NewDesc(
			prometheus.BuildFQName(space, "", "utilization_gpu"),
			"Percent of time over the past sample period during which one or more kernels were executing on the GPU device.",
			labels, nil,
		),
		utilizationMemory: prometheus.NewDesc(
			prometheus.BuildFQName(space, "", "utilization_memory"),
			"Percent of time over the past sample period during which one or more kernels were executing on the GPU device.",
			labels, nil,
		),
		utilizationGPUAverage: prometheus.NewDesc(
			prometheus.BuildFQName(space, "", "utilization_gpu_average"),
			"Used memory as reported by the device averraged over 10s.",
			labels, nil,
		),
	},nil
}

func (c *nvidaCollector) Update(ch chan<- prometheus.Metric) error {
	if err := gonvml.Initialize(); err != nil {
		log.Fatalf("Couldn't initialize gonvml: %v. Make sure NVML is in the shared library search path.", err)
	}
	defer gonvml.Shutdown()

	deviceCount, err := gonvml.DeviceCount()
	if err != nil {
		log.Printf("DeviceCount() error: %v", err)
		return err
	} else {
		ch <- prometheus.MustNewConstMetric(c.deviceCount, prometheus.CounterValue, float64(deviceCount))
	}
	for i := 0; i < int(deviceCount); i++ {
		dev, err := gonvml.DeviceHandleByIndex(uint(i))
		if err != nil {
			log.Printf("DeviceHandleByIndex(%d) error: %v", i, err)
			continue
		}
		minorNumber, err := dev.MinorNumber()
		if err != nil {
			log.Printf("MinorNumber() error: %v", err)
			continue
		}
		minor := strconv.Itoa(int(minorNumber))

		uuid, err := dev.UUID()
		if err != nil {
			log.Printf("UUID() error: %v", err)
			continue
		}

		name, err := dev.Name()
		if err != nil {
			log.Printf("Name() error: %v", err)
			continue
		}

		deviceInfo,err := gonvml.SystemDriverVersion()
		if err != nil {
			log.Printf("SystemDriverVersion() error: %v",err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.deviceInfo, prometheus.CounterValue, float64(0), minor, uuid, name,deviceInfo)
		}

		memoryTotal, memoryUsed, err := dev.MemoryInfo()
		if err != nil {
			log.Printf("MemoryInfo() error: %v", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.memoryUsed, prometheus.GaugeValue, float64(memoryUsed), minor, uuid, name)

			ch <- prometheus.MustNewConstMetric(
				c.memoryTotal, prometheus.GaugeValue, float64(memoryTotal), minor, uuid, name)
		}

		powerUsage, err := dev.PowerUsage()
		if err != nil {
			log.Printf("PowerUsage() error: %v", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.powerUsage, prometheus.GaugeValue, float64(powerUsage), minor, uuid, name)
		}

		temperature, err := dev.Temperature()
		if err != nil {
			log.Printf("Temperature() error: %v", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.temperature, prometheus.GaugeValue, float64(temperature), minor, uuid, name)
		}

		processInfo, err := dev.GetProcessInfo()
		if err != nil {
			log.Printf("ProcessInfo() error: %v", err)
		} else {
			for _,process := range processInfo {
				PID := fmt.Sprint(process.PID)
				Name := process.Name
				UsedMemory := process.UsedMemory
				ch <- prometheus.MustNewConstMetric(
					c.processInfo, prometheus.GaugeValue, float64(UsedMemory), minor, PID, Name)
			}
		}

		utilizationGPU, utilizationMemory, err := dev.UtilizationRates()
		if err != nil {
			log.Printf("UtilizationRates() error: %v", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.utilizationGPU, prometheus.GaugeValue, float64(utilizationGPU), minor, uuid, name)
			ch <- prometheus.MustNewConstMetric(
				c.utilizationMemory, prometheus.GaugeValue, float64(utilizationMemory), minor, uuid, name)
		}

		utilizationGPUAverage, err := dev.AverageGPUUtilization(averageDuration)
		if err != nil {
			log.Printf("AverageGPUUtilization() error: %v", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.utilizationGPUAverage, prometheus.GaugeValue, float64(utilizationGPUAverage), minor, uuid, name)
		}
	}
	return nil
}