package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Kaomoji struct {
	Text       string   `json:"text"`
	Categories []string `json:"categories"`
	Meaning    string   `json:"meaning"`
}

type SearchResponse struct {
	Data []Kaomoji `json:"data"`
}

func categoriesToString(cats []string) string {
	if len(cats) == 0 {
		return ""
	}
	return strings.Join(cats, ", ")
}

func searchKaomoji(q string) ([]Kaomoji, error) {
	apiURL := fmt.Sprintf(
		"https://kaomojis.jp/api/v1/kaomojis/search?q=%s&locale=en",
		url.QueryEscape(q),
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api status: %s", resp.Status)
	}

	var out SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out.Data, nil
}

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN is required")
	}

	b, err := bot.New(token, bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.InlineQuery == nil {
			return
		}

		query := strings.TrimSpace(update.InlineQuery.Query)
		if query == "" {
			query = "smile"
		}

		items, err := searchKaomoji(query)
		if err != nil {
			log.Println("search error:", err)
			_, _ = b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
				InlineQueryID: update.InlineQuery.ID,
				Results:       []models.InlineQueryResult{},
				CacheTime:     1,
				IsPersonal:    true,
			})
			return
		}

		results := make([]models.InlineQueryResult, 0, len(items))
		for i, k := range items {
			desc := categoriesToString(k.Categories)

			results = append(results, &models.InlineQueryResultArticle{
				ID:          fmt.Sprintf("%d", i),
				Title:       fmt.Sprintf(" %s", k.Text),
				Description: desc,
				InputMessageContent: &models.InputTextMessageContent{
					MessageText: k.Text,
				},
			})
		}

		_, err = b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
			InlineQueryID: update.InlineQuery.ID,
			Results:       results,
			CacheTime:     10,
			IsPersonal:    true,
		})
		if err != nil {
			log.Println("answer inline query error:", err)
		}
	}))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("bot started")
	b.Start(context.Background())
}
