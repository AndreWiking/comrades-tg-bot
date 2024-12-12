package main

import (
	"ComradesTG/db"
	"ComradesTG/gpt"
	"ComradesTG/matching"
	"ComradesTG/settings"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

type SendType int

const (
	TypeMessage SendType = iota
	TypePhoto
)

func ChangeKeyboard(bot *tgbotapi.BotAPI, chat_id int64, markup tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chat_id, settings.FindRoommateMText)
	msg.ReplyMarkup = markup

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func main() {
	//os.Exit(0)

	SetLogger()
	gpt.NewClient()
	var dbConnection db.Connection
	if err := dbConnection.Connect(); err != nil {
		log.Panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(settings.TgApiKey)
	if err != nil {
		log.Println(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	newUpdate := tgbotapi.NewUpdate(0)
	newUpdate.Timeout = 60

	updates := bot.GetUpdatesChan(newUpdate)

	for update := range updates {
		if update.Message == nil { // ignore non-Message updates
			continue
		}

		sendType := TypeMessage
		chat_id := update.Message.Chat.ID

		msg := tgbotapi.NewMessage(chat_id, "")
		msg.ParseMode = tgbotapi.ModeHTML

		var photo tgbotapi.PhotoConfig

		user := update.Message.From
		message := update.Message.Text

		switch message {
		case "/start":
			msg.Text = settings.StartMessage
			msg.ReplyMarkup = settings.MainKeyKeyboard
			if err := dbConnection.AddUser(user.ID, user.UserName, user.FirstName, user.LastName); err != nil {
				log.Println(err)
			}
			if err := dbConnection.SetUserState(user.ID, settings.StateMain); err != nil {
				log.Println(err)
			}
		case settings.FillFormBText:
			msg.Text = settings.EnterNameMText
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			if err := dbConnection.AddForm(user.ID); err != nil {
				log.Println(err)
			}
			if err := dbConnection.SetUserState(user.ID, settings.StateFormFirstName); err != nil {
				log.Println(err)
			}
		case settings.AboutUsBText:
			msg.Text = settings.AboutUsAnswerText
		case settings.MyFormBText:
			added, err := dbConnection.IsFormAdded(user.ID)
			if err != nil {
				log.Println(err)
			}
			if added {
				text, err := dbConnection.GetFormText(user.ID)
				if err != nil {
					log.Println(err)
				}
				msg.Text = text
				msg.ReplyMarkup = settings.MatchParamsKeyKeyboard
			} else {
				msg.Text = settings.NoFormFilledMText
			}

		case settings.FindRoommateBText:
			added, err := dbConnection.IsFormAdded(user.ID)
			if err != nil {
				log.Println(err)
			}
			if added {
				if err := matching.MatchGreedy(dbConnection, user.ID); err != nil {
					log.Println(err)
				}

				if err := dbConnection.SetUserMatchPos(user.ID, 0); err != nil {
					log.Println(err)
				}

				userVK, err := dbConnection.GetMatchVkUser(user.ID, 0)
				if err != nil {
					log.Println(err)
				}
				var emptyVK db.UserVK // todo: fix
				if userVK == emptyVK {
					msg.Text = settings.NoMatchFound
				} else {
					sendType = TypePhoto
					url := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(userVK.Photo_link))
					photo = tgbotapi.NewPhoto(update.Message.From.ID, url.Media)
					photo.Caption = db.PrintVkUserForm(userVK)
					photo.ReplyMarkup = settings.GetMatchInlineKeyboard(userVK.Profile_link, userVK.Post_link)
					ChangeKeyboard(bot, chat_id, settings.MatchKeyKeyboard)
				}

			} else {
				msg.Text = settings.NoFormFilledMText
			}

		case settings.NextBText:
			matchPos, err := dbConnection.GetUserMatchPos(user.ID)
			if err != nil {
				log.Println(err)
			}
			matchPos++
			if err := dbConnection.SetUserMatchPos(user.ID, matchPos); err != nil {
				log.Println(err)
			}

			userVK, err := dbConnection.GetMatchVkUser(user.ID, matchPos)
			if err != nil {
				log.Println(err)
			}

			sendType = TypePhoto
			url := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(userVK.Photo_link))
			photo = tgbotapi.NewPhoto(update.Message.From.ID, url.Media)
			photo.Caption = db.PrintVkUserForm(userVK)
			photo.ReplyMarkup = settings.GetMatchInlineKeyboard(userVK.Profile_link, userVK.Post_link)
			//ChangeKeyboard(bot, chat_id, matchKeyKeyboard)

		case settings.PrevBText:
			matchPos, err := dbConnection.GetUserMatchPos(user.ID)
			if err != nil {
				log.Println(err)
			}
			matchPos--
			if matchPos < 0 {
				matchPos = 0
			}
			if err := dbConnection.SetUserMatchPos(user.ID, matchPos); err != nil {
				log.Println(err)
			}

			userVK, err := dbConnection.GetMatchVkUser(user.ID, matchPos)
			if err != nil {
				log.Println(err)
			}

			sendType = TypePhoto
			url := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(userVK.Photo_link))
			photo = tgbotapi.NewPhoto(update.Message.From.ID, url.Media)
			photo.Caption = db.PrintVkUserForm(userVK)
			photo.ReplyMarkup = settings.GetMatchInlineKeyboard(userVK.Profile_link, userVK.Post_link)
			//ChangeKeyboard(bot, chat_id, matchKeyKeyboard)

		case settings.MenuBText:

			msg.Text = settings.StartMessage
			msg.ReplyMarkup = settings.MainKeyKeyboard
			if err := dbConnection.SetUserState(user.ID, settings.StateMain); err != nil {
				log.Println(err)
			}
		case settings.MatchBudgetBText:

			form, err := dbConnection.GetForm(user.ID)
			if err != nil {
				log.Println(err)
			}
			if err := dbConnection.SetUserState(user.ID, settings.StateMatchBudget); err != nil {
				log.Println(err)
			}
			msg.Text = fmt.Sprintf(settings.EnterMatchBudgetBText, form.Apartments_budget)
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

		case settings.MatchDistanceBText:

			form, err := dbConnection.GetForm(user.ID)
			if err != nil {
				log.Println(err)
			}
			if err := dbConnection.SetUserState(user.ID, settings.StateMatchDistance); err != nil {
				log.Println(err)
			}
			msg.Text = fmt.Sprintf(settings.EnterMatchDistanceBText, form.Apartments_location)
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

		default:
			//if state := settings.DetectUserState(message); state != settings.StateFormUnknown {
			//	if err := dbConnection.SetUserState(user.ID, state-1); err != nil {
			//		log.Println(err)
			//	}
			//}

			state, err := dbConnection.GetUserState(user.ID)
			if err != nil {
				log.Println(err)
			}

			fmt.Println(state)

			switch state {
			case settings.StateMain:
			case settings.StateFormFirstName:
				if err := dbConnection.SetFormValue(user.ID, "first_name", message); err != nil {
					log.Println(err)
				}
				if err := dbConnection.SetUserState(user.ID, settings.StateFormLastName); err != nil {
					log.Println(err)
				}
				msg.Text = settings.EnterLastnameMText

			case settings.StateFormLastName:
				if err := dbConnection.SetFormValue(user.ID, "last_name", message); err != nil {
					log.Println(err)
				}
				if err := dbConnection.SetUserState(user.ID, settings.StateFormSex); err != nil {
					log.Println(err)
				}
				msg.Text = settings.EnterSexMText
				msg.ReplyMarkup = settings.SexKeyKeyboard

			case settings.StateFormSex:
				sexType := settings.DetectSex(message)
				if sexType == settings.SexUnknown {
					msg.Text = settings.EnterIncorrectFormatText

				} else {
					if err := dbConnection.SetFormValue(user.ID, "sex", strconv.Itoa(int(sexType))); err != nil {
						log.Println(err)
					}
					if err := dbConnection.SetUserState(user.ID, settings.StateFormAge); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterAgeMText
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				}

			case settings.StateFormAge:
				age, err := strconv.Atoi(message)
				if err != nil {
					msg.Text = settings.EnterIncorrectFormatText
				} else {

					if err := dbConnection.SetFormValue(user.ID, "age", strconv.Itoa(age)); err != nil {
						log.Println(err)
					}
					if err := dbConnection.SetUserState(user.ID, settings.StateFormRoommateSex); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterRoommateSexMText
					msg.ReplyMarkup = settings.SexKeyKeyboard
				}

			case settings.StateFormRoommateSex:
				sexType := settings.DetectSex(message)
				if sexType == settings.SexUnknown {
					msg.Text = settings.EnterIncorrectFormatText

				} else {
					if err := dbConnection.SetFormValue(user.ID, "roommate_sex", strconv.Itoa(int(sexType))); err != nil {
						log.Println(err)
					}
					if err := dbConnection.SetUserState(user.ID, settings.StateFormApartmentsBudget); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterApartmentBudgetMText
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				}

			case settings.StateFormApartmentsBudget:
				budget, err := strconv.Atoi(message)
				if err != nil {
					msg.Text = settings.EnterIncorrectFormatText

				} else {

					if err := dbConnection.SetFormValue(user.ID, "apartments_budget", strconv.Itoa(budget)); err != nil {
						log.Println(err)
					}
					if err := dbConnection.SetUserState(user.ID, settings.StateFormApartmentsLocation); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterApartmentLocationMText
				}

			case settings.StateFormApartmentsLocation:

				locS, locW, err := gpt.TransformLocation(message)
				if err != nil {
					msg.Text = settings.EnterIncorrectFormatText
				} else {

					if err := dbConnection.SetFormValue(user.ID, "apartments_location_s", fmt.Sprintf("%f", locS)); err != nil {
						log.Println(err)
					}

					if err := dbConnection.SetFormValue(user.ID, "apartments_location_w", fmt.Sprintf("%f", locW)); err != nil {
						log.Println(err)
					}

					if err := dbConnection.SetFormValue(user.ID, "apartments_location", message); err != nil {
						log.Println(err)
					}

					if err := dbConnection.SetFormValue(user.ID, "apartments_location", message); err != nil {
						log.Println(err)
					}
					if err := dbConnection.SetUserState(user.ID, settings.StateFormAboutUser); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterAboutYouMText
				}

			case settings.StateFormAboutUser:
				if err := dbConnection.SetFormValue(user.ID, "about_user", message); err != nil {
					log.Println(err)
				}
				if err := dbConnection.SetUserState(user.ID, settings.StateFormAboutRoommate); err != nil {
					log.Println(err)
				}
				msg.Text = settings.EnterAboutRoommateMText

			case settings.StateFormAboutRoommate:
				if err := dbConnection.SetFormValue(user.ID, "about_roommate", message); err != nil {
					log.Println(err)
				}
				if err := dbConnection.SetUserState(user.ID, settings.StateMain); err != nil {
					log.Println(err)
				}
				msg.Text = settings.EnterSuccessFilledMText
				msg.ReplyMarkup = settings.MainKeyKeyboard

			case settings.StateMatchDistance:

				dist, err := strconv.ParseFloat(message, 64)
				if err != nil && dist >= 0 {
					msg.Text = settings.EnterIncorrectFormatText
				} else {
					if err := dbConnection.SetFormValue(user.ID, "match_distance", fmt.Sprintf("%f", dist)); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterMatchDistanceSuccessBText
					msg.ReplyMarkup = settings.MainKeyKeyboard
				}

			case settings.StateMatchBudget:

				budget, err := strconv.Atoi(message)
				if err != nil && budget >= 0 {
					msg.Text = settings.EnterIncorrectFormatText
				} else {
					if err := dbConnection.SetFormValue(user.ID, "match_budget", strconv.Itoa(budget)); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterMatchBudgetSuccessBText
					msg.ReplyMarkup = settings.MainKeyKeyboard
				}
			}
		}

		switch sendType {
		case TypeMessage:
			if msg.Text != "" {
				_, err = bot.Send(msg)
			}
		case TypePhoto:
			photo.ParseMode = tgbotapi.ModeHTML
			_, err = bot.Send(photo)
		}
		if err != nil {
			log.Println(err)
		}
	}
}

/*

systemctl status comrades-tg-bot.service
systemctl restart comrades-tg-bot.service
comrades-tg-bot
sudo systemctl restart nginx

7541929739:AAFylnUcAeDvSueJGIGQ5kAfow4nEw7P-Oc

*/
