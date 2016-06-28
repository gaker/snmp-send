package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/parnurzeal/gorequest"
	g "github.com/soniah/gosnmp"
	"log"
	"os"
)

var (
	cmdConfigFile string
)

func init() {
	flag.StringVar(
		&cmdConfigFile, "conf",
		"example.config.json", "Location of config file")
	flag.Parse()
}

type Oids struct {
	Name string `json:"name"`
	Oid  string `json:"oid"`
}

type SNMPConfig struct {
	Community string `json:"community"`
	Hostname  string `json:"hostname"`
	IP        string `json:"ip"`
	Oids      struct {
		CPU  []Oids `json:"cpu"`
		Disk []struct {
			Name string `json:"name"`
			Oids []Oids `json:"oids"`
		} `json:"disk"`
		Interfaces []struct {
			Name string `json:"name"`
			Oids []Oids `json:"oids"`
		} `json:"interfaces"`
		LoadAverage []Oids `json:"load_average"`
		Memory      []Oids `json:"memory"`
	} `json:"oids"`
	ServerSubType string `json:"server_sub_type"`
	ServerType    string `json:"server_type"`
	Timeout       int    `json:"timeout"`
	Verbose       bool   `json:"verbose"`
	Database      string `json:"database"`
	ReceiverUrl   string `json:"receiver_url"`
	ReceiverToken string `json:"receiver_token"`
}

type Point struct {
	Measurement string `json:"measurement"`
	Fields      Field  `json:"fields"`
}

type Field struct {
	Value interface{} `json:"value"`
}

type MessagePayload struct {
	Database string `json:"db"`
	Data     []Item `json:"data"`
}

type Item struct {
	Tags struct {
		Hostname      string `json:"hostname"`
		Ip            string `json:"ip"`
		ServerType    string `json:"server_type"`
		ServerSubType string `json:"server_sub_type"`
		Sample        string `json:"sample"`
		NIC           string `json:"nic,omitempty"`
		Disk          string `json:"disk,omitempty"`
	} `json:"tags"`
	Points []Point `json:"points"`
}

/**
 * Get SNMP metric
 */
func getSimpleMetric(measurement string, oids []Oids, conf SNMPConfig) []Item {

	g.Default.Target = conf.IP
	g.Default.Community = conf.Community
	err := g.Default.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %s", err)
	}
	defer g.Default.Conn.Close()

	items := []Item{}

	lookupOids := []string{}

	for _, oid := range oids {
		lookupOids = append(lookupOids, oid.Oid)
	}

	result, err2 := g.Default.Get(lookupOids)

	if err2 != nil {
		log.Fatalf("Get() err: %v", err2)
	}

	for _, variable := range result.Variables {
		point := Point{}
		point.Measurement = measurement

		item := Item{}
		item.Tags.Hostname = conf.Hostname
		item.Tags.Ip = conf.IP
		item.Tags.ServerType = conf.ServerType
		item.Tags.ServerSubType = conf.ServerSubType

		// Which oid is this?
		for _, o := range oids {
			if o.Oid == variable.Name {
				item.Tags.Sample = o.Name
			}
		}

		m := Field{}

		switch variable.Type {
		case g.OctetString:
			m.Value = string(variable.Value.([]byte))
		default:
			m.Value = g.ToBigInt(variable.Value)
		}

		point.Fields = m

		item.Points = append(item.Points, point)
		items = append(items, item)
	}
	return items
}

/**
 * main()
 */
func main() {

	file, err := os.Open(cmdConfigFile)

	if err != nil {
		fmt.Println(err)
		return
	}

	decoder := json.NewDecoder(file)
	conf := SNMPConfig{}
	err = decoder.Decode(&conf)

	if err != nil {
		fmt.Println("error:", err)
		return
	}

	measurement := Point{}
	measurement.Measurement = conf.IP

	// Get all the values from the config file
	payload := MessagePayload{}
	payload.Database = conf.Database

	loadAverages := getSimpleMetric("load_average", conf.Oids.LoadAverage, conf)
	payload.Data = loadAverages

	cpu := getSimpleMetric("cpu", conf.Oids.CPU, conf)
	payload.Data = append(payload.Data, cpu...)

	memory := getSimpleMetric("memory", conf.Oids.Memory, conf)
	payload.Data = append(payload.Data, memory...)

	for _, iface := range conf.Oids.Interfaces {
		item := getSimpleMetric("interfaces", iface.Oids, conf)
		for o, _ := range item {
			item[o].Tags.NIC = iface.Name
		}
		payload.Data = append(payload.Data, item...)
	}

	for _, disk := range conf.Oids.Disk {
		item := getSimpleMetric("disk", disk.Oids, conf)
		for o, _ := range item {
			item[o].Tags.Disk = disk.Name
		}
		payload.Data = append(payload.Data, item...)
	}

	req := gorequest.New()
	req.Post(conf.ReceiverUrl).
		Set("Authorization", fmt.Sprintf("Bearer %s", conf.ReceiverToken)).
		Set("Content-Type", "application/json").
		Send(payload).
		End()
}
