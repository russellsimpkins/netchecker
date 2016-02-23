/*
The MIT License (MIT)
Copyright (c) 2016 Russell Simpkins <russellsimpkins @ gmail com>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

Netchecker is a simple program to see if you can connect to a given
hostname:port combination. The code attempts to support UDP tests.
Testing UDP connectivity is iffy at best and the test will only pass
if the UPD service responds.

Usage:

netchecker -f /path/to/config.yaml

Here is an example of the configuration.

$> cat config.yaml
tcp:
  - "192.168.33.10:8300"
  - "192.168.33.10:8301"
  - "192.168.33.10:8302"
  - "192.168.33.10:8400"
  - "192.168.33.10:8500"
  - "192.168.33.10:8600"
udp:
  - "192.168.33.10:8301"
  - "192.168.33.10:8302"
  - "192.168.33.10:8600"
timeout_seconds: 2
EOF
*/
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// beware, the yaml parser has rules about field names. Best off to
// name the fields by hand and avoid not getting the struct populated
type Config struct {
	TCP            []string      `yaml:"tcp"`
	UDP            []string      `yaml:"udp"`
	TimeoutSeconds time.Duration `yaml:"timeout_seconds"`
}

// this function reads a yaml file to populate the struct
func (csl *Config) ParseYaml(yamlFile string) (err error) {

	file, err := os.Open(yamlFile)
	if err != nil {
		return err
	}
	_ = file.Close()
	data, err := ioutil.ReadFile(yamlFile)
	if nil != err {
		panic(err)
	}

	err = yaml.Unmarshal(data, csl)
	if err != nil {
		panic(err)
	}
	return
}

func main() {
	var (
		c   net.Conn
		err error
	)

	yamlFile := flag.String("f", "", "/abs/path/to/file.yaml that has your configuration.")
	flag.Parse()

	if *yamlFile == "" {
		fmt.Println("You need to specify a valid yaml file with the -f flag. e.g. -f /path/to/config.yaml")
		return
	}

	config := Config{}

	err = config.ParseYaml(*yamlFile)
	if err != nil {
		fmt.Printf("Unable to parse the yaml file %s. Error %s\n", *yamlFile, err)
		return
	}

	t := config.TimeoutSeconds * time.Second

	// test tcp connectivity
	for _, host := range config.TCP {
		c, err = net.DialTimeout("tcp", host, t)
		if err != nil {
			fmt.Printf("%s TCP CLOSED\n", host, err)
		} else {
			fmt.Printf("%s TCP OPEN\n", host)
			c.Close()
		}
	}

	// test udp. Given that udp is connection less - this is
	// a "best effort". I've discovered that even if there's nothing
	// listening on a udp port, the code for connecting works. So
	// tests ONLY Pass if the UDP listener responds with anything.
	for _, host := range config.UDP {

		c, err = net.DialTimeout("udp", host, t)
		if err != nil {
			fmt.Printf("%s UDP CLOSED %s\n", host, err)
			continue
		}
		ra, _ := net.ResolveUDPAddr("udp", host)
		u, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
		u.SetWriteDeadline(time.Now().Add(config.TimeoutSeconds * time.Second))
		u.SetReadDeadline(time.Now().Add(config.TimeoutSeconds * time.Second))

		if err != nil {
			fmt.Printf("%s UDP CLOSED %s\n", host, err)
			continue
		}

		b := []byte("CONNECTED-MODE SOCKET")

		if _, err = u.WriteTo(b, ra); err != nil {
			fmt.Printf("%s UDP CLOSED %s\n", host, err)
			continue
		}

		// now make a write/read attempt
		c := make(chan int, 1)
		go func() {
			time.Sleep((config.TimeoutSeconds + 2) * time.Second)
			c <- -1
		}()

		in := make([]byte, 1500)
		go func() {
			n, _, _ := u.ReadFrom(in)
			c <- n
		}()

		v := <-c
		if v == 0 {
			fmt.Printf("%s UDP CLOSED %d\n", host, v)
		} else {
			fmt.Printf("%s UDP OPEN\n", host)
		}

		u.Close()
	}
}
