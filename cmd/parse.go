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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// parseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// parseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
