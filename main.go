package main

import (
	"encoding/json"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/stianeikeland/go-rpio"
)

type MemoryObj struct {
	MemAvailable string
	MemFree      string
	MemTotal     string
}

func main() {

	sysMem := MemoryClean(Exec("cat", "/proc/meminfo"), "MemTotal:", "MemFree:", "MemAvailable:")
	if sysMem == nil {
		panic("null")
	}
	jSysMem, err := json.Marshal(sysMem)
	if err != nil {
		log.Println(err)
		panic("null")
	}

	var memObj MemoryObj

	err = json.Unmarshal(jSysMem, &memObj)
	if err != nil {
		log.Println(err)
		panic("null")
	}

	log.Println(memObj.MemFree)

	err = rpio.Open()
	if err != nil {
		log.Println(err)
		panic("null")
	}

	pin23 := rpio.Pin(23)
	pin23.Output()
	pin23.Low()
	pin24 := rpio.Pin(24)
	pin24.Output()
	pin24.Low()

	loc, err := time.LoadLocation("Europe/Kiev")
	if err != nil {
		log.Println(err)
		panic("null")
	}

	for {
		var timeNowHours = time.Now().In(loc).Hour()

		//turn off coller from 19:00 till 8:00
		if timeNowHours >= 19 || timeNowHours <= 7 {
			log.Print("Zzzzzz....")
			pin23.Low()
			time.Sleep(1 * time.Minute)
			continue
		}

		tmp := getTemp()
		log.Printf("TMP: %f C", tmp)
		if tmp > 42 {
			pin23.High()
			time.Sleep(1 * time.Minute)
		} else {
			pin23.Low()
			time.Sleep(1 * time.Second)
		}
	}
}

func getTemp() float64 {
	cpuTemp := CPUClean(Exec("vcgencmd", "measure_temp"), "temp=", "'C")
	if cpuTemp == "" {
		panic("null")
	}

	tmp, err := strconv.ParseFloat(cpuTemp, 32)
	if err != nil {
		panic(err)
	}
	return tmp
}

// Exec execute program and return stdout
func Exec(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(out)
}

// CPUClean return Cpu data
func CPUClean(str string, args ...string) string {
	for _, arg := range args {
		str = strings.Replace(str, arg, "", -1)
	}
	str = strings.TrimSpace(str)
	return str
}

// MemoryClean return MemTotal, MemFree, MemAvailable.
func MemoryClean(str string, args ...string) map[string]string {
	if str == "" {
		return nil
	}
	result := make(map[string]string)
	strArray := strings.Split(str, "\n")
	for _, arg := range args {
		for _, val := range strArray {
			if str := strings.Split(val, arg); len(str) == 2 {
				newStr := strings.TrimSpace(strings.Replace(str[1], "kB", "", -1))
				result[strings.Replace(arg, ":", "", -1)] = newStr
			}
		}
	}
	return result
}
