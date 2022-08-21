package cmd

import (
	"fmt"
	"github.com/LeweyM/blogparser/internal/parse"

	"github.com/spf13/cobra"
)

// parseCmd represents the parse command
var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "parse the blog posts",
	Long:  `Parse the blog posts.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("parse called")
		path, err := cmd.Flags().GetString("path")
		if err != nil {
			fmt.Printf("Error parsing 'path' flag: %v", err)
			return
		}

		imgPath, err := cmd.Flags().GetString("imgPath")
		if err != nil {
			fmt.Printf("Error parsing 'imgPath' flag: %v", err)
			return
		}

		handler := parse.NewHandler(path, imgPath)
		err = handler.Handle()
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().StringP("path", "p", "", "directory to parse")
	_ = parseCmd.MarkFlagRequired("path")
	_ = parseCmd.MarkFlagDirname("path")

	parseCmd.Flags().StringP("imgPath", "i", "", "directory to parse image files")
	_ = parseCmd.MarkFlagDirname("imgPath")
}
