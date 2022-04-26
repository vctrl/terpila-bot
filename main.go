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
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var WebhookURL = "https://terpila-bot.herokuapp.com/"
var BotToken = os.Getenv("BOT_TOKEN")
var Port = ":" + os.Getenv("PORT")
var quotes = []string {
	"Основа всякой мудрости есть терпение. Платон\n",
	"Сила заключает в себе также терпение. В нетерпении проявляется слабость. Г. Гауптман\n",
	"Терпение — добродетель бессильного и украшение сильного. Древнеиндийский афоризм\n",
	"Терпение — добродетель нищих. Ф. Массинджер\n",
	"Терпение — единственное настоящее испытание цивилизации. А. Хелпс\n",
	"Терпение — лучшая религия. В. Гюго\n",
	"Терпение — опора слабости; нетерпение — гибель силы. Ч. Колтон\n",
	"Терпение — это дитя силы, упрямство — плод слабости, а именно слабости ума. М. Эбнер-Эшенбах\n",
	"Терпение — это то, без чего Вера, Надежда, Любовь ничто. Р. Зубкова\n",
	"Терпение и время дают больше, чем сила или страсть. Ж. Лафонтен\n",
}

type resp map[int64][]string

func r(receiverID int64, msg string) resp {
	return map[int64][]string {receiverID: {msg}}
}

type cmdHandler func(ctx context.Context, upd *tgbotapi.Update, params ...string) (resp, error)

type CmdNotSupportedErr struct {
	msg string
}

func (e *CmdNotSupportedErr) Error() string {
	return e.msg
}

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

func (tb *TerpilaBot) ExecuteCmd(upd *tgbotapi.Update) (resp, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if upd.MyChatMember != nil {
		return tb.AddToChat(ctx, upd)
	}
	if upd.Message != nil && upd.Message.NewChatMembers != nil && len(upd.Message.NewChatMembers) > 0 {
		return tb.InviteMember(ctx, upd)
	}

	cmdHandler, ok := tb.Cmds[upd.Message.Text]
	if !ok {
		return tb.Parse(ctx, upd)
	}

	return cmdHandler(ctx, upd)
}

func (tb *TerpilaBot) AddToChat(ctx context.Context, upd *tgbotapi.Update, params ...string) (resp, error) {
	return r(upd.MyChatMember.Chat.ID, "Всем чмоки в этом чате! Изберём же путь смирения, чтобы обрести вечную жизнь!"), nil
}

func (tb *TerpilaBot) InviteMember(ctx context.Context, upd *tgbotapi.Update, params ...string) (resp, error) {
	invited := upd.Message.NewChatMembers[0].UserName
	creator := upd.Message.From.UserName
	return r(upd.Message.Chat.ID,
			fmt.Sprintf("поздравляем %s с вступлением в ряд терпил! %s с нами уже давно, у него есть чему поучиться!",
				invited, creator)), nil
}

func (tb *TerpilaBot) Tolerate(ctx context.Context, upd *tgbotapi.Update, params ...string) (resp, error) {
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
		upd.Message.Chat.ID: msgs,
	}, nil
}

func (tb *TerpilaBot) GetStats(ctx context.Context, upd *tgbotapi.Update, params ...string) (resp, error) {
	cnt, err := tb.Tolerances.GetCountByUser(ctx, upd.Message.From.ID)
	if err != nil {
		return nil, errors.WithMessage(err, "get count by user id")
	}

	postfix := raz(cnt)
	result := r(upd.Message.Chat.ID, fmt.Sprintf("Ты затерпел %d %s", cnt, postfix))

	return result, nil
}

func (tb *TerpilaBot) Parse(ctx context.Context, upd *tgbotapi.Update, params ...string) (resp, error) {
	s := strings.ToLower(upd.Message.Text)
	if strings.Contains(s, "терпил") ||
		strings.Contains(s, "терпеть") ||
		strings.Contains(s, "терпение") ||
		strings.Contains(s, "терпения") ||
		strings.Contains(s, "терпел") {
		return r(upd.Message.Chat.ID, quotes[rand.Intn(len(quotes))]), nil
	}

	if strings.Contains(s, "вова спаси") {
		return r(upd.Message.Chat.ID, "https://www.youtube.com/watch?v=YdXhQXMvYuM"), nil
	}

	return map[int64][]string{}, nil
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
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

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

	//u := tgbotapi.NewUpdate(0)
	// u.Timeout = 60
	//updates := bot.GetUpdatesChan(u)

	updates := useWh(bot, WebhookURL)

	for update := range updates {
		result, err := tb.ExecuteCmd(&update)

		var chatID int64
		// new member status in group - not a message
		if update.MyChatMember == nil && update.Message == nil {
			sugar.Errorf("unknown message: %v", update)
			continue
		}

		if update.MyChatMember != nil {
			chatID = update.MyChatMember.Chat.ID
		} else {
			chatID = update.Message.Chat.ID
		}

		switch err1 := errors.Cause(err).(type) {
		case *CmdNotSupportedErr:
			bot.Send(tgbotapi.NewMessage(chatID, err1.Error()))
			if err != nil {
				sugar.Errorf("error executing command: %v", err)
			}

			if err != nil {
				sugar.Errorf("error executing command: %v", err)
			}

			continue
		default:
			sugar.Errorf("error executing command: %v", err)
		}
		if err != nil {
			sugar.Errorf("error executing command: %v", err)
		}

		for id, msgs := range result {
			for _, m := range msgs {
				_, err = bot.Send(tgbotapi.NewMessage(id, m))
				if err != nil {
					sugar.Errorf("error executing command: %v", err)
				}
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

func useWh(bot *tgbotapi.BotAPI, url string) tgbotapi.UpdatesChannel {
	wh, err := tgbotapi.NewWebhook(url)
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

	return bot.ListenForWebhook("/")
}