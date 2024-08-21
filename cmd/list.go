/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"quno/internal/notes"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		list, err := notes.GetNotes()
		if err != nil {
			fmt.Printf("Failed to get notes %v\n", err)
			return
		}

		templates := &promptui.SelectTemplates{
			Label:    "{{ .Title }}?",
			Active:   "  {{ .Title | green }} ({{ .CreatedAt | red }})",
			Inactive: "{{ .Title | cyan }} ({{ .CreatedAt | red }})",
		}

		prompt := promptui.Select{
			Label:        "Notes",
			Items:        list,
			Templates:    templates,
			Size:         16,
			HideSelected: true,
		}

		i, _, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		note := notes.GetNote(list[i])

		fmt.Println(note.ToString())
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
