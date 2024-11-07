package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	rawMode bool
	showMeta bool
)

type ArticleMetadata struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Author      string `json:"author,omitempty"`
	Published   string `json:"published,omitempty"`
}

type ArticleContent struct {
	Content  string          `json:"content"`
	Metadata ArticleMetadata `json:"metadata"`
}

func extractArticle(url string) (*ArticleContent, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	metadata := ArticleMetadata{
		Title:       doc.Find("title").Text(),
		Description: doc.Find("meta[name='description']").AttrOr("content", ""),
		Author:      doc.Find("meta[name='author']").AttrOr("content", ""),
		Published:   doc.Find("meta[property='article:published_time']").AttrOr("content", ""),
	}

	var articleContent string
	articleSelectors := []string{
		"article",
		"main",
		"div.content",
		"div.article-content",
		"div.main-content",
	}

	var contentNode *goquery.Selection
	for _, selector := range articleSelectors {
		contentNode = doc.Find(selector).First()
		if contentNode.Length() > 0 {
			break
		}
	}

	if contentNode.Length() > 0 {
		contentNode.Find("script, style, nav, header, footer, aside").Remove()
		articleContent, err = contentNode.Html()
		if err != nil {
			return nil, fmt.Errorf("error extracting HTML content: %v", err)
		}
	} else {
		return nil, fmt.Errorf("no article content found")
	}

	return &ArticleContent{
		Content:  strings.TrimSpace(articleContent),
		Metadata: metadata,
	}, nil
}

func convertToMarkdown(html string) (string, error) {
	converter := md.NewConverter("", true, nil)
	converter.Use(plugin.GitHubFlavored())
	converter.Use(plugin.Table())
	
	markdown, err := converter.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("error converting to markdown: %v", err)
	}
	
	return markdown, nil
}

func runExtract(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("please provide a URL to extract")
	}

	url := args[0]
	article, err := extractArticle(url)
	if err != nil {
		return fmt.Errorf("extraction failed: %v", err)
	}

	if showMeta {
		metadataJSON, err := json.MarshalIndent(article.Metadata, "", "  ")
		if err != nil {
			return fmt.Errorf("error formatting metadata: %v", err)
		}
		fmt.Fprintln(os.Stderr, string(metadataJSON))
	}

	if rawMode {
		fmt.Println(article.Content)
	} else {
		markdown, err := convertToMarkdown(article.Content)
		if err != nil {
			return fmt.Errorf("markdown conversion failed: %v", err)
		}
		fmt.Println(markdown)
	}

	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "extract [url]",
		Short: "Extract article content from any URL",
		Long: `Extract is a simple CLI tool that extracts article content from URLs and converts them to markdown.
It removes clutter like ads, navigation, and scripts, giving you clean, readable content.

Example:
  extract https://example.com/article.html > article.md
  extract --raw https://example.com/article.html > article.html
  extract --meta https://example.com/article.html | less`,
		RunE:    runExtract,
		Version: version,
	}

	rootCmd.Flags().BoolVarP(&rawMode, "raw", "r", false, "output raw HTML instead of markdown")
	rootCmd.Flags().BoolVarP(&showMeta, "meta", "m", false, "show article metadata")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
