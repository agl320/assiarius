package scraper

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: "scrape [link]",
		Short:	"Scarpe news from a given link",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ScrapeNewsFromLink(args[0])
		}
	}
}

func ScrapeNewsFromLink(link string) {

}

func fetchPage(url string) (string, error) {

}

func parsePage(content string) {


}