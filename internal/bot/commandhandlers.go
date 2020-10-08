package bot

import (
	"dutybot/internal/calendar"
	"dutybot/internal/database"
	db "dutybot/internal/database"
	"fmt"
	"log"
	"regexp"
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
	as, err := db.GetTodaysAssignment(msg.Chat.ID)
	if err != nil {
		log.Print(err)

		reply := tgbot.NewMessage(msg.Chat.ID, "Couldn't fetch today's duty.")
		_, err := bot.Send(reply)
		if err != nil {
			log.Print(err)
		}
		return
	}

	operator := fmt.Sprintf("@%s", as.Operator.UserName)
	reply := tgbot.NewMessage(msg.Chat.ID, operator)
	_, err = bot.Send(reply)
	if err != nil {
		log.Print(err)
	}
}

/// This function assumes that date
/// is ordered like DD MM YYYY
func parseTime(probablyTime string) (t time.Time, err error) {
	r := regexp.MustCompile("([0-9]{1,2}).*?([0-9]{1,2}).*?([0-9]{4})")
	if !r.MatchString(probablyTime) {
		err = fmt.Errorf("'%s' is not look like DD MM YYYY", probablyTime)
		return
	}
	parts := r.FindAllStringSubmatch(probablyTime, 1)[0]
	date := fmt.Sprintf("%02s", parts[1])
	month := fmt.Sprintf("%02s", parts[2])
	year := parts[3]
	t, err = time.Parse("02 01 2006", fmt.Sprintf("%s %s %s", date, month, year))
	return
}

func assign(bot *tgbot.BotAPI, msg *tgbot.Message) {
	dutydate, err := parseTime(msg.CommandArguments())
	if err != nil {
		log.Print(err)
		reply := tgbot.NewMessage(msg.Chat.ID, "Something wrong with date.")
		_, err := bot.Send(reply)
		if err != nil {
			log.Print(err)
		}
		return
	}

	if time.Now().After(dutydate) {
		reply := tgbot.NewMessage(
			msg.Chat.ID,
			fmt.Sprintf(
				"Assignment is possible only for a future date",
				dutydate.Format("02-01-2006"),
			),
		)
		_, err := bot.Send(reply)
		if err != nil {
			log.Print(err)
		}
		return
	}

	as, err := database.GetAssignmentByDate(msg.Chat.ID, dutydate)
	if err != nil {
		log.Print(err)
	}
	if as != nil {
		log.Printf("%+v %+v", as, as.Operator)
		reply := tgbot.NewMessage(
			msg.Chat.ID,
			fmt.Sprintf("This day already occupied by `%s`", as.Operator.UserName),
		)
		reply.ParseMode = "Markdown"
		_, err := bot.Send(reply)
		if err != nil {
			log.Print(err)
		}
		return
	}

	if calendar.IsHoliday(dutydate) {
		answer := fmt.Sprintf(
			"%s is a holiday. No duty on holidays",
			dutydate.Format(time.RFC1123),
		)
		reply := tgbot.NewMessage(msg.Chat.ID, answer)
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
	log.Printf("%+v", a)
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

	assignments, err := db.GetAssignmentSchedule(weeks, msg.Chat.ID)
	if err != nil {
		reply := tgbot.NewMessage(msg.Chat.ID, "Couldn't get assignments.")
		_, err := bot.Send(reply)
		if err != nil {
			log.Print(err)
		}
		return
	}

	var b strings.Builder
	for _, ass := range assignments {
		dutyDate := time.Unix(ass.DutyDate, 0).Format("Mon\tJan 02 2006")
		b.WriteString(fmt.Sprintf("`%s\t\t%s`", ass.Operator.UserName, dutyDate))
		b.WriteRune('\n')
	}

	reply := tgbot.NewMessage(msg.Chat.ID, b.String())
	reply.ParseMode = "Markdown"
	_, err = bot.Send(reply)
	if err != nil {
		log.Print(err)
	}
}
