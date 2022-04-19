package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/vctrl/terpila-bot/db"
	"github.com/vctrl/terpila-bot/db/memory"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var WebhookURL = "https://terpila-bot.herokuapp.com/"
var BotToken = os.Getenv("BOT_TOKEN")
var Port = ":" + os.Getenv("PORT")

type cmdHandler func(ctx context.Context, upd *tgbotapi.Update, params ...string) (map[int64][]string, error)

type TerpilaBot struct {
	Cmds map[string]cmdHandler

	Terpiloids db.Terpiloids

	Tolerances db.Tolerances

}

func NewTerpilaBot(ter db.Terpiloids, tol db.Tolerances) *TerpilaBot {
	tb := &TerpilaBot{
		Terpiloids: ter,
		Tolerances: tol,
	}

	tb.Cmds = map[string]cmdHandler {
		"/tolerate": tb.Tolerate,
		"/stats": tb.GetStats,
	}

	return tb
}

func (tb *TerpilaBot) ExecuteCmd(upd *tgbotapi.Update) (map[int64][]string, error) {
	cmdHandler, ok := tb.Cmds[upd.Message.Text]
	if !ok {
		return nil, errors.New("command not supported")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return cmdHandler(ctx, upd)
}

func (tb *TerpilaBot) Tolerate(ctx context.Context, upd *tgbotapi.Update, params ...string) (map[int64][]string, error) {
	// todo create new user if not exist
	err := tb.Tolerances.Add(ctx, db.NewTolerance(uuid.New(), upd.Message.From.ID))
	if err != nil {
		return nil, errors.WithMessage(err, "add tolerance")
	}

	cnt, err := tb.Tolerances.GetCountByUser(ctx, upd.Message.From.ID)
	if err != nil {
		return nil, errors.WithMessage(err, "get count by user")
	}

	msgs := make([]string, 0, 1)
	if cnt == 1 {
		msgs = append(msgs,"В первый раз может быть непривычно, но всё приходит с опытом!")
	}

	msgs = append(msgs, "Затерпел")
	return map[int64][]string {
		upd.Message.From.ID: msgs,
	}, nil
}

func (tb *TerpilaBot) GetStats(ctx context.Context, upd *tgbotapi.Update, params ...string) (map[int64][]string, error) {
	cnt, err := tb.Tolerances.GetCountByUser(ctx, upd.Message.From.ID)
	if err != nil {
		return nil, errors.WithMessage(err, "get count by user id")
	}

	postfix := raz(cnt)
	result := map[int64][]string{upd.Message.From.ID: {fmt.Sprintf("Ты затерпел %d %s", cnt, postfix)}}

	return result, nil
}

func raz(n int64) string {
	if n >= 10 && n <= 20 {
		return "раз"
	}

	d := n % 10

	if d == 2 || d == 3 || d == 4 {
		return "раза"
	}

	return "раз"
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatalf("failed to create new bot api: %v", err)
	}

	wh, err := tgbotapi.NewWebhook(WebhookURL)
	if err != nil {
		log.Fatalf("failed to create webhook: %v", err)
	}

	_, err = bot.Request(wh)
	if err != nil {
		log.Fatalf("error setting webhook: %v", err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/")
	server := &http.Server{
		Addr: Port,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to listen and serve: %+v", err)
		}
	}()

	fmt.Printf("server started at %s\n", Port)

	tol := memory.NewTolerancesMemory()
	tb := NewTerpilaBot(nil, tol)

	for update := range updates {
		result, err := tb.ExecuteCmd(&update)
		chatID := update.Message.Chat.ID
		if err != nil {
			sugar.Errorf("error executing command: %v", err)
			bot.Send(tgbotapi.NewMessage(chatID, "error happened"))
		}

		for id, msgs := range result {
			for _, m := range msgs {
				bot.Send(tgbotapi.NewMessage(id, m))
			}
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	err = server.Shutdown(ctx)
	if err != nil {
		server.Close()
		log.Fatal("error shutdown server")
	}
}
