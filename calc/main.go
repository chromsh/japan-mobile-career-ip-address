package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/seancfoley/ipaddress-go/ipaddr"
)

type Output struct {
	V4 []string `json:"ipv4"`
	V6 []string `json:"ipv6"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: calc <file1> <file2> ...")
		os.Exit(1)
	}
	ips, err := mergeIPAddrFromFile(os.Args[1:])
	if err != nil {
		panic(err)
	}
	output := Output{V4: []string{}, V6: []string{}}
	for _, ip := range ips {
		if ip.IsIPv4() {
			output.V4 = append(output.V4, ip.String())
		} else if ip.IsIPv6() {
			output.V6 = append(output.V6, ip.String())
		} else {
			panic("Unknown IP type: " + ip.String())
		}
	}
	data, _ := json.Marshal(output)
	fmt.Println(string(data))
}

func mergeIPAddrFromFile(files []string) ([]*ipaddr.IPAddress, error) {
	var mergeIPs []*ipaddr.IPAddress
	var subIPs []*ipaddr.IPAddress
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			exclude := strings.HasPrefix(line, "-")
			if exclude {
				line = strings.TrimSpace(line[1:])
			}
			ip := ipaddr.NewIPAddressString(line).GetAddress()
			if ip == nil {
				return nil, fmt.Errorf("%s: invalid IP address %q", file, scanner.Text())
			}
			if exclude {
				// 削除対象アドレス
				subIPs = append(subIPs, ip)
				continue
			}
			mergeIPs = append(mergeIPs, ip)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	if len(mergeIPs) == 0 {
		return nil, nil
	}
	mergedIPs := mergeIPs[0].MergeToPrefixBlocks(mergeIPs[1:]...)
	// 引き算
	for _, subIP := range subIPs {
		var remaining []*ipaddr.IPAddress
		for _, mergedIP := range mergedIPs {
			remaining = append(remaining, mergedIP.Subtract(subIP)...)
		}
		mergedIPs = remaining
	}
	if len(mergedIPs) == 0 {
		return nil, nil
	}
	// 再度マージして返す。おそらくマージしなくて良いがとりあえず
	return mergedIPs[0].MergeToPrefixBlocks(mergedIPs[1:]...), nil
}
