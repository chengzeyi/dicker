package network

import (
	"io/ioutil"
	"net"
	"reflect"
	"testing"
)

func TestIpAddrManager_Alloc(t *testing.T) {
	tmpAllocatorFile, err := ioutil.TempFile("", "dicker_ip_test")
	tmpAllocatorPath := tmpAllocatorFile.Name()
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		SubnetAllocatorPath string
		Subnets             map[string][]bool
	}
	type args struct {
		subnet *net.IPNet
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    net.IP
		wantErr bool
	}{
		{
			name: "192.168.1.0/24",
			fields: fields{
				SubnetAllocatorPath: tmpAllocatorPath,
			},
			args: args{
				subnet: getIpNetFromCIDR("192.168.1.0/24"),
			},
			want:    net.ParseIP("192.168.1.1"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipam := &IpAddrManager{
				SubnetAllocatorPath: tt.fields.SubnetAllocatorPath,
				Subnets:             tt.fields.Subnets,
			}
			got, err := ipam.Alloc(tt.args.subnet)
			if (err != nil) != tt.wantErr {
				t.Errorf("IpAddrManager.Alloc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IpAddrManager.Alloc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIpAddrManager_Release(t *testing.T) {
	tmpAllocatorFile, err := ioutil.TempFile("", "dicker_ip_test")
	tmpAllocatorPath := tmpAllocatorFile.Name()
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		SubnetAllocatorPath string
		Subnets             map[string][]bool
	}
	type args struct {
		subnet *net.IPNet
		ip     net.IP
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "192.168.1.0/24",
			fields: fields{
				SubnetAllocatorPath: tmpAllocatorPath,
			},
			args: args{
				subnet: getIpNetFromCIDR("192.168.1.0/24"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipam := &IpAddrManager{
				SubnetAllocatorPath: tt.fields.SubnetAllocatorPath,
				Subnets:             tt.fields.Subnets,
			}
			// Prepare IP to be released.
			tt.args.ip, err = ipam.Alloc(tt.args.subnet)
			if err != nil {
				t.Fatalf("IpAddrManager.Alloc() error = %v", err)
			}
			if err := ipam.Release(tt.args.subnet, tt.args.ip); (err != nil) != tt.wantErr {
				t.Errorf("IpAddrManager.Release() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func getIpFromCIDR(s string) net.IP {
	ip, _, _ := net.ParseCIDR(s)
	return ip
}

func getIpNetFromCIDR(s string) *net.IPNet {
	_, ipNet, _ := net.ParseCIDR(s)
	return ipNet
}
