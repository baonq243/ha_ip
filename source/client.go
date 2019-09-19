package main

import (
	"bufio"
	"fmt"
	"github.com/sparrc/go-ping"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)
func checkClient(ip string) bool {
	a, err := ping.NewPinger(ip)
	if err != nil {
		fmt.Println(err)
	}
	a.Count = 3
	a.Timeout = time.Second * 3
	a.SetPrivileged(true)
	a.Run()
	stats := a.Statistics()
	if stats.PacketLoss > 98 {
		return false
	} else {
		return true
	}
}
func getConfig() (client string, server string, port string, listIP []string) {
	viper.SetConfigFile("/opt/ha_ip/config.yml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Không tìm thấy file config, %s", err)
	}
	client = viper.GetString("client")
	server = viper.GetString("server")
	listIP = viper.GetStringSlice("list_ip")
	port = viper.GetString("port")
	return
}
func getLoIP() (listIP string) {
	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {
		if addrs, err := interf.Addrs(); err == nil {
			for _, addr := range addrs {
				currentIP := addr.(*net.IPNet)
				if currentIP.IP.To4() != nil && currentIP.IP.String() != "127.0.0.1" && interf.Name == "lo" {
					listIP = listIP + currentIP.IP.String() + "\n"
				}
			}
		}
	}
	return strings.TrimSuffix(listIP, "\n")
}
func checkIPExist(ip string, a []string) bool {
	for _, v := range a {
		if ip == v {
			return true
		}
	}
	return false
}
func setIP(act string, add string) {
	cmd := exec.Command("sudo", "ip", "addr", act, add, "dev", "lo")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	stdout, err := cmd.Output()
	if err != nil {
		println(err.Error())
		return
	}
	print(string(stdout))
}
func main() {
	time.Sleep(time.Second * 5)
	for {
		fmt.Println("_______________________")
		ipClient, ipServer, port, listIPConfig := getConfig()
		fmt.Println("Get config file")
		fmt.Println("IP_Client: ", ipClient)
		fmt.Println("IP_Server: ", ipServer)
		fmt.Println("Port_Server: ", port)
		fmt.Println("VIP_IP: ", listIPConfig)
		fmt.Println("_______________________")
		ipServerPort := ipServer + ":" +port
		d := net.Dialer{Timeout: time.Second * 2}
		conn, check := d.Dial("tcp", ipServerPort)
		if check != nil && checkClient(ipServer) == false{
			fmt.Println(check)
			fmt.Println("Connection Error")
			for _,v := range listIPConfig {
				fmt.Println("Add ip: ",v)
				setIP("add",v)
			}
		} else {
			fmt.Println("_______________________")
			fmt.Println("Connection OK")
			msg := ""
			msg = getLoIP()
			fmt.Println(msg)
			conn.Write([]byte(msg))
			conn.Write([]byte("-"))

			reader := bufio.NewReader(conn)
			g, _ := reader.ReadString('-')
			g = strings.TrimSuffix(g, "-")
			listIPServer := strings.Split(g, "\n")

			listIPClient := strings.Split(msg, "\n")
			fmt.Println("_______________________")
			fmt.Println("listIPClient")
			for _, v := range listIPClient {
				fmt.Println(v)
			}
			fmt.Println("_______________________")
			fmt.Println("listIPServer")
			for _, v := range listIPServer {
				fmt.Println(v)
			}
			fmt.Println("_______________________")
			fmt.Println("ListConfig")
			for _, v := range listIPConfig {
				fmt.Println(v)
			}
			fmt.Println("_______________________")
			for _, v := range listIPConfig {
				if checkIPExist(v, listIPClient) && checkIPExist(v, listIPServer) {
					fmt.Printf("Del IP: %v \n",v)
					setIP("del",v)
					fmt.Println("_______________________")
				}
			}
			conn.Close()
		}
		time.Sleep(time.Second * 5)
	}
}
