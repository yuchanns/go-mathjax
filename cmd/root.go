package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/yuchanns/yuchanns/pre-render/internal"
	"log"
)

var (
	dirs []string
	ext  string
)

var rootCmd = &cobra.Command{
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.Process(context.Background(), dirs, ext)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringSliceVarP(&dirs, "dir", "d", []string{}, "scan directory")
	rootCmd.PersistentFlags().StringVarP(&ext, "ext", "x", "", "extension of files")

	_ = cobra.MarkFlagRequired(rootCmd.PersistentFlags(), "dir")
}
