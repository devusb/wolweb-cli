package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var server string

var rootCmd = &cobra.Command{
	Use:   "wolweb-cli",
	Short: "A command line interface to trigger devices with wolweb",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available devices",
	Run:   listDevices,
}

var wakeCmd = &cobra.Command{
	Use:   "wake",
	Short: "Wake device",
	Run:   wakeDevice,
}

func init() {
	rootCmd.AddCommand(listCmd, wakeCmd)

	setupConfig()
}

func setupConfig() {
	viper.SetConfigName("wolweb-cli.yaml")
	viper.SetConfigType("yaml")

	viper.AddConfigPath("$HOME/.config")
	
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Failure reading config: %s", err)
	}

	server = viper.GetString("server")
}

func listDevices(cmd *cobra.Command, args []string) {
	hosts, err := getDevices()
	if err != nil {
		log.Fatalf("failed to get hosts: %s", err)
	}

	lw := list.NewWriter()
	for _, v := range(hosts) {
		lw.AppendItem(fmt.Sprintf("%s: %s", v.Name, v.Mac))
	}

	fmt.Println("Devices:")
	fmt.Println(lw.Render())
}

func getDevices() ([]Device, error) {
	resp, err := http.Get(fmt.Sprintf("%s/wolweb/data/get", server))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	devices := AppData{}
	json.Unmarshal(body, &devices)

	return devices.Devices, nil
}

func wakeDevice(cmd *cobra.Command, args []string) {
	target := args[0]

	resp, err := http.Get(fmt.Sprintf("%s/wolweb/wake/%s", server, target))
	if err != nil {
		log.Fatalf("Failed to wake device: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to wake device: %s", err)
	}

	response := HTTPResponseObject{}
	json.Unmarshal(body, &response)

	if !response.Success {
		log.Fatalf("Failed to wake device")
	}

	fmt.Println(response.Message)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
