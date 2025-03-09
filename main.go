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

func SendMessage(bot *tgbotapi.BotAPI, chat_id int64, text string) {
	msg := tgbotapi.NewMessage(chat_id, text)

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func ChangeKeyboard(bot *tgbotapi.BotAPI, chat_id int64, markup tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chat_id, "⬇️⬇️⬇️")
	msg.ReplyMarkup = markup

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func main() {
	//os.Exit(0)

	logFile := SetLogger()
	defer logFile.Close()
	defer log.Println("Session finished")

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

		//fmt.Println(update.Message.Invoice.StartParameter)
		fmt.Println(update.Message.Text)

		switch message {
		case settings.StartText + " " + settings.UtmYa1.String():
			msg.Text = settings.StartMessage
			msg.ReplyMarkup = settings.MainKeyKeyboard
			if err := dbConnection.AddUser(user.ID, user.UserName, user.FirstName, user.LastName, settings.UtmYa1); err != nil {
				log.Println(err)
			}

		case settings.StartText:
			msg.Text = settings.StartMessage
			msg.ReplyMarkup = settings.MainKeyKeyboard
			if err := dbConnection.AddUser(user.ID, user.UserName, user.FirstName, user.LastName, settings.UtmEmpty); err != nil {
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
				log.Println(err.Error())
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
				SendMessage(bot, chat_id, settings.FindRoommateMText)
				if err := matching.MatchGreedy(dbConnection, user.ID); err != nil {
					log.Println(err)
				}

				if err := dbConnection.SetUserMatchPos(user.ID, 0); err != nil {
					log.Println(err)
				}

				userVK, haveNext, err := dbConnection.GetMatchVkUser(user.ID, 0)
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

					if haveNext {
						ChangeKeyboard(bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNoPrev))
					} else {
						ChangeKeyboard(bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNoPrevNext))
					}
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

			userVK, haveNext, err := dbConnection.GetMatchVkUser(user.ID, matchPos)
			if err != nil {
				log.Println(err)
			}

			sendType = TypePhoto
			url := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(userVK.Photo_link))
			photo = tgbotapi.NewPhoto(update.Message.From.ID, url.Media)
			photo.Caption = db.PrintVkUserForm(userVK)
			photo.ReplyMarkup = settings.GetMatchInlineKeyboard(userVK.Profile_link, userVK.Post_link)

			if haveNext {
				ChangeKeyboard(bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNormal))
			} else {
				ChangeKeyboard(bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNoNext))
			}

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

			userVK, _, err := dbConnection.GetMatchVkUser(user.ID, matchPos)
			if err != nil {
				log.Println(err)
			}

			sendType = TypePhoto
			url := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(userVK.Photo_link))
			photo = tgbotapi.NewPhoto(update.Message.From.ID, url.Media)
			photo.Caption = db.PrintVkUserForm(userVK)
			photo.ReplyMarkup = settings.GetMatchInlineKeyboard(userVK.Profile_link, userVK.Post_link)

			if matchPos > 0 {
				ChangeKeyboard(bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNormal))
			} else {
				ChangeKeyboard(bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNoPrev))
			}

		case settings.MenuBText:

			msg.Text = settings.MenuMessage
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
				if len(message) <= settings.LimitEnterNamesLen {
					if err := dbConnection.SetFormValue(user.ID, "first_name", message); err != nil {
						log.Println(err)
					}
					if err := dbConnection.SetUserState(user.ID, settings.StateFormLastName); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterLastnameMText
				} else {
					msg.Text = settings.EnterIncorrectFormatText
				}

			case settings.StateFormLastName:
				if len(message) <= settings.LimitEnterNamesLen {
					if err := dbConnection.SetFormValue(user.ID, "last_name", message); err != nil {
						log.Println(err)
					}
					if err := dbConnection.SetUserState(user.ID, settings.StateFormSex); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterSexMText
					msg.ReplyMarkup = settings.SexKeyKeyboard
				} else {
					msg.Text = settings.EnterIncorrectFormatText
				}

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
				if err != nil || age < 0 || age > 99 {
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
				if err != nil || budget < 0 || budget > settings.LimitBudget {
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
				if err != nil || len(message) > settings.LimitEnterTextLen {
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
				if len(message) < settings.LimitEnterTextLen {
					if err := dbConnection.SetFormValue(user.ID, "about_user", message); err != nil {
						log.Println(err)
					}
					if err := dbConnection.SetUserState(user.ID, settings.StateFormAboutRoommate); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterAboutRoommateMText
				} else {
					msg.Text = settings.EnterIncorrectFormatText
				}

			case settings.StateFormAboutRoommate:
				if len(message) < settings.LimitEnterTextLen {
					if err := dbConnection.SetFormValue(user.ID, "about_roommate", message); err != nil {
						log.Println(err)
					}
					if err := dbConnection.SetUserState(user.ID, settings.StateMain); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterSuccessFilledMText
					msg.ReplyMarkup = settings.MainKeyKeyboard
				} else {
					msg.Text = settings.EnterIncorrectFormatText
				}

			case settings.StateMatchDistance:

				dist, err := strconv.ParseFloat(message, 64)
				if err != nil || dist < 0.0 || dist > settings.LimitMatchDist {
					msg.Text = settings.EnterIncorrectFormatText
				} else {
					if err := dbConnection.SetFormValue(user.ID, "match_distance", fmt.Sprintf("%f", dist)); err != nil {
						log.Println(err)
					}
					form, err := dbConnection.GetForm(user.ID)
					if err != nil {
						log.Println(err)
					}
					msg.Text = fmt.Sprintf(settings.EnterMatchDistanceSuccessBText, form.Apartments_location, dist)
					msg.ReplyMarkup = settings.MainKeyKeyboard
				}

			case settings.StateMatchBudget:

				budget, err := strconv.Atoi(message)
				if err != nil || budget < 0 || budget > settings.LimitBudget {
					msg.Text = settings.EnterIncorrectFormatText
				} else {
					if err := dbConnection.SetFormValue(user.ID, "match_budget", strconv.Itoa(budget)); err != nil {
						log.Println(err)
					}
					form, err := dbConnection.GetForm(user.ID)
					if err != nil {
						log.Println(err)
					}
					msg.Text = fmt.Sprintf(settings.EnterMatchBudgetSuccessBText, form.Apartments_budget, budget)
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

ssh root@46.17.41.227

systemctl status comrades-tg-bot.service
systemctl restart comrades-tg-bot.service
comrades-tg-bot
sudo systemctl restart nginx

7541929739:AAFylnUcAeDvSueJGIGQ5kAfow4nEw7P-Oc

scp -r /Users/andrewiking/GolandProjects/ComradesTG root@46.17.41.227:/root/
go build .
systemctl restart ComradesTG
systemctl status ComradesTG

psql -h <REMOTE HOST> -p <REMOTE PORT> -U <DB_USER> <DB_NAME>

psql -h 46.17.41.227 -U super_admin postgres

su - postgres
psql

systemctl start ComradesTG
systemctl status ComradesTG


systemctl status postgres


https://t.me/find_comrade_bot?start=ya1

*/
