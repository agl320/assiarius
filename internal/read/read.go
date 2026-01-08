package read

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-shiori/go-readability"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read [link]",
		Short: "Read news from a given link",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ReadNewsFromLink(args[0])
		},
	}

	return cmd
}

func ReadNewsFromLink(link string) error {
	resp, err := http.Get(link)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	
	parsedURL, err := url.Parse(link)
	
	if err != nil {
		return err
	}

	article, err := readability.FromReader(resp.Body, parsedURL)
	
	if err != nil {
		return nil
	}

	fmt.Println(article.TextContent)

	return nil
} 