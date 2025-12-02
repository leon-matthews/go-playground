package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	// Pflag replaces for Go's flag package, implementing GNU-style --flags.
	pflag.StringP("name", "n", "Leon", "Name parameter")
	pflag.StringP("password", "p", "hardToGuess", "Password")
	pflag.Parse()

	// Viper's handles configuration from env, files, and/or command-line, etc.
	viper.BindPFlags(pflag.CommandLine)
	fmt.Println(viper.GetString("password"))
}

func aliasNormalise(f *pflag.FlagSet, n string) pflag.NormalizedName {
	switch n {
	case "pass":
		n = "password"
	case "ps":
		n = "password"
	}
	return pflag.NormalizedName(n)
}
