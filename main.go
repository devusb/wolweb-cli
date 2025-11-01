package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var server string

var rootCmd = &cobra.Command{
	Use:   "wolweb-cli",
	Short: "A command line interface to wake devices with wolweb",
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List available devices",
	Run:     listDevices,
}

var wakeCmd = &cobra.Command{
	Use:     "wake [device]",
	Aliases: []string{"w"},
	Short:   "Wake device",
	Long:    "Wake device matching name from [list]",
	Run:     wakeDevice,
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
		fmt.Printf("Failure reading config: %s\n", err)
		os.Exit(1)
	}

	server = viper.GetString("server")
}

func listDevices(cmd *cobra.Command, args []string) {
	hosts, err := getDevices()
	if err != nil {
		fmt.Printf("failed to get hosts: %s\n", err)
		os.Exit(1)
	}

	lw := list.NewWriter()
	for _, v := range hosts {
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
		fmt.Printf("Failed to wake device: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to wake device: %s\n", err)
		os.Exit(1)
	}

	response := HTTPResponseObject{}
	json.Unmarshal(body, &response)

	if !response.Success {
		fmt.Printf("Failed to wake device: %s\n", response.Message)
		os.Exit(1)
	}

	fmt.Println(response.Message)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
