/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"log"

	"github.com/lllrdgz/cc-hyperpay-go/hyperpay-transfer/client"
	"github.com/spf13/cobra"
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Reads the details of the given account",
	Long: `Reads the details of the given account.
			Receives an id transaction and reads its value`,
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		contract, err := client.NewHyperPayContract()
		if err != nil {
			log.Fatalf("Failed to create contract client: %v", err)
		}
		log.Println("--> Evaluate Transaction: ReadAccount, function reads the value of an account")
		acc, err := contract.Read(id)
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %v", err)
		}
		accBytes, err := json.Marshal(*acc)
		if err != nil {
			panic(err)
		}
		log.Println(string(accBytes))
	},
}

func init() {
	rootCmd.AddCommand(readCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// readCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// readCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
