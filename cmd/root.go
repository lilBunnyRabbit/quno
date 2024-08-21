/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"quno/internal/notes"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "quno",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// Title

		minChars := 2
		maxChars := 128

		titlePrompt := promptui.Prompt{
			Label: "Title",
			Validate: func(input string) error {
				inputLen := len(input)

				if inputLen < minChars {
					return fmt.Errorf("title must contain at least %d characters", minChars)
				}

				if inputLen > maxChars {
					return fmt.Errorf("title must contain at most %d characters", maxChars)
				}

				return nil
			},
			HideEntered: true,
		}

		title, err := titlePrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		fmt.Printf("%s%s\n", color.CyanString("Title: "), title)

		// Description

		reader := bufio.NewReader(os.Stdin)
		fmt.Println(strings.Repeat("-", len(title)+7))

		var lines []string
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading line: %v\n", err)
				break
			}

			trimmedLine := strings.TrimSpace(line)
			if trimmedLine == "" {
				break // Stop reading when an empty line is encountered
			}

			lines = append(lines, line)
		}

		content := strings.Join(lines, "")

		fileName, err := notes.CreateNote(title, content)
		if err != nil {
			fmt.Println("Failed to create note:", err)
			return
		}

		fmt.Printf("%s%s\n", color.GreenString("Not created at: "), color.CyanString(fileName))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.quno.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".quno" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".quno")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}