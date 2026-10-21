package telegram

import (
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	bot *tgbotapi.BotAPI
}

func NewClient(token string) (*Client, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telegram bot api: %w", err)
	}
	return &Client{
		bot: bot,
	}, nil
}

func (c *Client) SendFile(chatID int64, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for telegram upload: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to ead file stats: %w", err)
	}

	fileFileReader := tgbotapi.FileReader{
		Name:   fileInfo.Name(),
		Reader: file,
	}

	docMsg := tgbotapi.NewDocument(chatID, fileFileReader)

	_, err = c.bot.Send(docMsg)
	if err != nil {
		return fmt.Errorf("telegram api dispatch error: %w", err)
	}
	return nil
}
