package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

//content to look for at start of document
const original = `<!DOCTYPE html>
<html>`

func main() {
	ips := make(chan net.IP, 64)
	done := false

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)

		go func() {
			timeout := time.Duration(1 * time.Second)
			client := http.Client{
				Timeout: timeout,
			}

			processed := 0
			for !done || len(ips) > 0 {
				ip := <-ips
				processed++
				response, err := client.Get("http://" + ip.String())

				if err != nil {
					continue
				}

				buf := new(bytes.Buffer)
				buf.ReadFrom(response.Body)
				newStr := buf.String()

				if strings.Contains(newStr, original) {
					fmt.Printf("----- found: %s", newStr)
				}

			}
			wg.Done()
		}()
	}

	go func() {
		//generate some tasks

		//the ip range to search through
		ip, ipnet, err := net.ParseCIDR("162.243.0.0/16")
		if err != nil {
			log.Fatal(err)
		}
		for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
			ips <- ip
		}
		close(ips)
		done = true
	}()

	// wait for workers to finish
	wg.Wait()
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}