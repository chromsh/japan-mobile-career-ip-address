package main

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMergeIPAddrFromFile(t *testing.T) {
	tests := []struct {
		name, input string
		want        []string
		wantErr     bool
	}{
		{"multiple exclusions", "106.128.0.0/13\n-106.134.0.0/16\n-106.135.0.0/16\n", []string{"106.128.0.0/14", "106.132.0.0/15"}, false},
		{"disjoint exclusions", "192.0.2.0/24\n-192.0.2.0/26\n-192.0.2.128/26\n", []string{"192.0.2.64/26", "192.0.2.192/26"}, false},
		{"whole block", "192.0.2.0/24\n-192.0.2.0/24\n", nil, false},
		{"containing block", "192.0.2.0/24\n-192.0.0.0/16\n", nil, false},
		{"outside block", "192.0.2.0/24\n-198.51.100.0/24\n", []string{"192.0.2.0/24"}, false},
		{"nested exclusions", "192.0.2.0/24\n-192.0.2.0/25\n-192.0.2.0/26\n", []string{"192.0.2.128/25"}, false},
		{"empty", "\n\t\n", nil, false},
		{"exclusions only", "-192.0.2.0/24\n", nil, false},
		{"zero address", "0.0.0.0\n", []string{"0.0.0.0/32"}, false},
		{"invalid", "not-an-ip\n", nil, true},
		{"invalid exclusion", "192.0.2.0/24\n-invalid\n", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input.ip")
			if err := os.WriteFile(path, []byte(tt.input), 0600); err != nil {
				t.Fatal(err)
			}
			ips, err := mergeIPAddrFromFile([]string{path})
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}
			var got []string
			for _, ip := range ips {
				got = append(got, ip.String())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublishedIPv4MatchesInputs(t *testing.T) {
	for _, carrier := range []string{"docomo", "au"} {
		t.Run(carrier, func(t *testing.T) {
			input := carrier + "_v4.ip"
			if carrier == "docomo" {
				input = "docomo_v4.json" // Historical filename; contents are line-delimited CIDRs.
			}
			ips, err := mergeIPAddrFromFile([]string{input})
			if err != nil {
				t.Fatal(err)
			}
			var want []string
			for _, ip := range ips {
				want = append(want, ip.String())
			}
			content, err := os.ReadFile(filepath.Join("..", "data", carrier+".json"))
			if err != nil {
				t.Fatal(err)
			}
			var published Output
			if err := json.Unmarshal(content, &published); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(published.V4, want) {
				t.Fatalf("published IPv4 does not match source inputs: got %v, want %v", published.V4, want)
			}
			if carrier == "au" {
				for _, cidr := range published.V4 {
					_, block, err := net.ParseCIDR(cidr)
					if err != nil {
						t.Fatal(err)
					}
					for _, excluded := range []string{"106.134.0.0/16", "106.135.0.0/16"} {
						ip, exclusion, _ := net.ParseCIDR(excluded)
						if block.Contains(ip) || exclusion.Contains(block.IP) {
							t.Fatalf("%s intersects excluded %s", cidr, excluded)
						}
					}
				}
			}
		})
	}
}
