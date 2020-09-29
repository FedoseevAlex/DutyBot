package bot

import (
	db "dutybot/internal/database"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

var handlers map[string]func(*tgbot.BotAPI, *tgbot.Message)

func initHandlers() {
	handlers = make(map[string]func(*tgbot.BotAPI, *tgbot.Message))
	handlers["help"] = help
	handlers["assign"] = assign
	handlers["show"] = show
	handlers["operator"] = operator
}

func processCommands(bot *tgbot.BotAPI, command *tgbot.Message) {
	handler, ok := handlers[command.Command()]
	if !ok {
		msg := tgbot.NewMessage(command.Chat.ID, "Unknown command. Try /help")
		_, err := bot.Send(msg)
		if err != nil {
			log.Print(err)
		}
		return
	}
	handler(bot, command)
}

func help(bot *tgbot.BotAPI, msg *tgbot.Message) {
	helpString := `Usage:
/help - look at this message again
/operator - tag current duty
/show [weeks (default=2)] - show duty schedule for some weeks ahead
/assign [date] - assign yourself for duty. Date should be in format DD-MM-YYYY`
	answer := tgbot.NewMessage(msg.Chat.ID, helpString)
	_, err := bot.Send(answer)
	if err != nil {
		log.Print(err)
	}
}

func operator(bot *tgbot.BotAPI, msg *tgbot.Message) {
	ass, err := db.GetTodaysOperator(msg.Chat.ID)
	if err != nil {
		reply := tgbot.NewMessage(msg.Chat.ID, "Я обосрался")
		log.Print(err)
		_, err := bot.Send(reply)
		if err != nil {
			log.Print(err)
		}
		return
	}

	var b strings.Builder
	for i, as := range ass {
		log.Printf("%d. %+v, %+v", i, as, as.Operator)
		b.WriteString(fmt.Sprintf("@%s", as.Operator.UserName))
	}

	reply := tgbot.NewMessage(msg.Chat.ID, b.String())
	_, err = bot.Send(reply)
	if err != nil {
		log.Print(err)
	}
}

func assign(bot *tgbot.BotAPI, msg *tgbot.Message) {
	dutydate, err := time.Parse("02-01-2006", msg.CommandArguments())
	if err != nil {
		reply := tgbot.NewMessage(msg.Chat.ID, "Something wrong with date.")
		_, err := bot.Send(reply)
		if err != nil {
			log.Print(err)
		}
		return
	}
	op := &db.Operator{
		UserName:  msg.From.UserName,
		FirstName: msg.From.FirstName,
		LastName:  msg.From.LastName,
	}
	err = op.Get()
	if err != nil {
		err = op.Insert()
		if err != nil {
			log.Print(err)
			return
		}
	}
	a := &db.Assignment{ChatID: msg.Chat.ID, DutyDate: dutydate.Unix(), Operator: op}
	err = a.Insert()
	if err != nil {
		log.Print(err)
		return
	}
	reply := tgbot.NewMessage(msg.Chat.ID, "Assigned")
	_, err = bot.Send(reply)
	if err != nil {
		log.Print(err)
	}
}

func show(bot *tgbot.BotAPI, msg *tgbot.Message) {
	weeks, err := strconv.Atoi(msg.CommandArguments())
	if err != nil {
		weeks = 2
	}

	assignments, err := db.GetAssignmentSchedule(weeks)
	if err != nil {
		reply := tgbot.NewMessage(msg.Chat.ID, "Couldn't get assignments.")
		_, err := bot.Send(reply)
		if err != nil {
			log.Print(err)
		}
		return
	}

	var b strings.Builder
	for i, ass := range assignments {
		log.Printf("%d. %+v, %+v", i, ass, ass.Operator)
		b.WriteString(fmt.Sprintf("%s %s", ass.Operator.UserName, time.Unix(ass.DutyDate, 0)))
		b.WriteRune('\n')
	}

	reply := tgbot.NewMessage(msg.Chat.ID, b.String())
	_, err = bot.Send(reply)
	if err != nil {
		log.Print(err)
	}
}
