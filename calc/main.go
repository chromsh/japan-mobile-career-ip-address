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
	// ファイルから全IPを読み込む
	ips, err := readIPsFromFiles(os.Args[1:])
	if err != nil {
		panic(err)
	}

	// IPv4とIPv6に分離
	var v4s []*ipaddr.IPAddress
	var v6s []*ipaddr.IPAddress
	var subV4s []*ipaddr.IPAddress
	var subV6s []*ipaddr.IPAddress

	for _, item := range ips {
		ip := item.ip
		if ip.IsIPv4() {
			if item.isSub {
				subV4s = append(subV4s, ip)
			} else {
				v4s = append(v4s, ip)
			}
		} else if ip.IsIPv6() {
			if item.isSub {
				subV6s = append(subV6s, ip)
			} else {
				v6s = append(v6s, ip)
			}
		}
	}

	// マージと除外処理
	finalV4 := processIPs(v4s, subV4s)
	finalV6 := processIPs(v6s, subV6s)

	output := Output{V4: []string{}, V6: []string{}}
	for _, ip := range finalV4 {
		output.V4 = append(output.V4, ip.String())
	}
	for _, ip := range finalV6 {
		output.V6 = append(output.V6, ip.String())
	}

	// 整形して出力
	data, _ := json.MarshalIndent(output, "", "    ")
	fmt.Println(string(data))
}

type ipItem struct {
	ip    *ipaddr.IPAddress
	isSub bool
}

func readIPsFromFiles(files []string) ([]ipItem, error) {
	var items []ipItem
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
			isSub := false
			ipStr := line
			if strings.HasPrefix(line, "-") {
				isSub = true
				ipStr = line[1:]
			}

			ip := ipaddr.NewIPAddressString(ipStr).GetAddress()
			if ip == nil {
				// パースエラーの場合はスキップ
				continue
			}
			items = append(items, ipItem{ip: ip, isSub: isSub})
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func processIPs(mergeIPs []*ipaddr.IPAddress, subIPs []*ipaddr.IPAddress) []*ipaddr.IPAddress {
	if len(mergeIPs) == 0 {
		return []*ipaddr.IPAddress{}
	}

	// マージ
	mergedIPs := mergeIPs[0].MergeToPrefixBlocks(mergeIPs[1:]...)

	// 引き算
	var result []*ipaddr.IPAddress
	for _, mergedIP := range mergedIPs {
		currentIPs := []*ipaddr.IPAddress{mergedIP}

		for _, subIP := range subIPs {
			var nextIPs []*ipaddr.IPAddress
			for _, cip := range currentIPs {
				subtracted := cip.Subtract(subIP)
				if len(subtracted) > 0 {
					for _, s := range subtracted {
						nextIPs = append(nextIPs, s)
					}
				}
			}
			currentIPs = nextIPs
		}
		result = append(result, currentIPs...)
	}

	if len(result) == 0 {
		return []*ipaddr.IPAddress{}
	}

	// 最後に再度マージして整理
	return result[0].MergeToPrefixBlocks(result[1:]...)
}
