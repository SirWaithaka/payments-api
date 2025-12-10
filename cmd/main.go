/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/SirWaithaka/payments-api/cmd/payments"
)

func main() {
	os.Exit(payments.Execute())
}
