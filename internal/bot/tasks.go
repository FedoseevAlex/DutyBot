package bot

import (
	"context"
	"fmt"

	"github.com/FedoseevAlex/DutyBot/internal/database/assignment"
	"github.com/FedoseevAlex/DutyBot/internal/logger"
	"github.com/FedoseevAlex/DutyBot/internal/utils"
)

func announceDutyTask() {
	msgFormat := "@%s is on duty today"
	logger.Log.Debug().Msg("Start duty announcing")
	assignments, err := assignment.AssignmentRepo.GetAssignmentScheduleAllChats(
		context.Background(),
		utils.GetToday(),
	)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("announceDutyTask job failed to get operators")
		return
	}

	for _, assignment := range assignments {
		logger.Log.Debug().Msgf("Sending %+v\n", assignment)
		sendMessage(
			assignment.ChatID,
			fmt.Sprintf(msgFormat, assignment.Operator),
			NoParseMode,
		)
	}
}

func warnAboutFreeSlots() {
	logger.Log.Debug().Msg("Start freeslots announcing")

	chats, err := assignment.AssignmentRepo.GetAllChats(context.Background())
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("warnAboutFreeSlots job failed to get all chat IDs")
		return
	}

	for _, chatID := range chats {
		outputSlots, err := getFreeSlotsTable(chatID, DefaultFreeSlotWeeks)
		if err != nil {
			logger.Log.Error().
				Err(err).
				Msg("warnAboutFreeSlots job failed to tabulate free slots")
			return
		}

		if outputSlots == "" {
			continue
		}

		sendMessage(
			chatID,
			fmt.Sprintf("Free slots still available!\n%s\n", outputSlots),
			NoParseMode,
		)
	}
}
