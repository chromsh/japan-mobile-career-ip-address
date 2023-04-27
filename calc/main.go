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
			line := scanner.Text()
			ip := ipaddr.NewIPAddressString(line).GetAddress()
			if ip == nil || ip.IsZero() {
				// println("nil ip " + line)
				continue
			}
			if strings.HasPrefix(line, "-") {
				ip = ipaddr.NewIPAddressString(line[1:]).GetAddress()
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

	mergedIPs := mergeIPs[0].MergeToPrefixBlocks(mergeIPs[1:]...)
	// 引き算
	var ips []*ipaddr.IPAddress
	for _, mergedIP := range mergedIPs {
		isSubtracted := false
		for _, subIP := range subIPs {
			println(subIP.String())
			subtracted := mergedIP.Subtract(subIP)
			if len(subtracted) == 0 || mergedIP.Equal(subtracted[0]) {
				continue
			}
			ips = append(ips, subtracted...)
			isSubtracted = true
		}
		if !isSubtracted {
			ips = append(ips, mergedIP)
		}
	}
	// 再度マージして返す。おそらくマージしなくて良いがとりあえず
	return ips[0].MergeToPrefixBlocks(ips[1:]...), nil
}
