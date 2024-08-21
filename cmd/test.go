/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type pepper struct {
	Name     string
	HeatUnit int
	Peppers  int
}

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		peppers := []pepper{
			{Name: "Bell Pepper", HeatUnit: 0, Peppers: 0},
			{Name: "Banana Pepper", HeatUnit: 100, Peppers: 1},
			{Name: "Poblano", HeatUnit: 1000, Peppers: 2},
			{Name: "Jalapeño", HeatUnit: 3500, Peppers: 3},
			{Name: "Aleppo", HeatUnit: 10000, Peppers: 4},
			{Name: "Tabasco", HeatUnit: 30000, Peppers: 5},
			{Name: "Malagueta", HeatUnit: 50000, Peppers: 6},
			{Name: "Habanero", HeatUnit: 100000, Peppers: 7},
			{Name: "Red Savina Habanero", HeatUnit: 350000, Peppers: 8},
			{Name: "Dragon’s Breath", HeatUnit: 855000, Peppers: 9},
		}

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "\U0001F336 {{ .Name | cyan }} ({{ .HeatUnit | red }})",
			Inactive: "  {{ .Name | cyan }} ({{ .HeatUnit | red }})",
			Selected: "\U0001F336 {{ .Name | red | cyan }}",
			Details: `
	--------- Pepper ----------
	{{ "Name:" | faint }}	{{ .Name }}
	{{ "Heat Unit:" | faint }}	{{ .HeatUnit }}
	{{ "Peppers:" | faint }}	{{ .Peppers }}`,
		}

		searcher := func(input string, index int) bool {
			pepper := peppers[index]
			name := strings.Replace(strings.ToLower(pepper.Name), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label:     "Spicy Level",
			Items:     peppers,
			Templates: templates,
			Size:      4,
			Searcher:  searcher,
		}

		i, _, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		fmt.Printf("You choose number %d: %s\n", i+1, peppers[i].Name)
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
