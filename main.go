package main

import (
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
	msg := tgbotapi.NewMessage(chat_id, "Поиск ...")
	msg.ReplyMarkup = markup

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func main() {

	var dbConnection DbConnection
	if err := dbConnection.Connect(); err != nil {
		log.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(TgApiKey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

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
			msg.Text = StartMessage
			msg.ReplyMarkup = mainKeyKeyboard
			if err := dbConnection.AddUser(user.ID, user.UserName, user.FirstName, user.LastName); err != nil {
				log.Fatal(err)
			}
			if err := dbConnection.SetUserState(user.ID, StateMain); err != nil {
				log.Fatal(err)
			}
		case FillFormBText:
			msg.Text = EnterNameMText
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			if err := dbConnection.AddForm(user.ID); err != nil {
				log.Fatal(err)
			}
			if err := dbConnection.SetUserState(user.ID, StateFormFirstName); err != nil {
				log.Fatal(err)
			}
		case AboutUsBText:
			msg.Text = AboutUsAnswerText
		case MyFormBText:
			text, err := dbConnection.GetFormText(user.ID)
			if err != nil {
				log.Fatal(err)
			}
			msg.Text = text

		case FindRoommateBText:
			added, err := dbConnection.IsFormAdded(user.ID)
			if err != nil {
				log.Fatal(err)
			}
			if added {
				if err := dbConnection.SetUserMatchPos(user.ID, 0); err != nil {
					log.Fatal(err)
				}

				userVK, err := dbConnection.GetMatchVkUser(user.ID, 0)
				if err != nil {
					log.Fatal(err)
				}

				sendType = TypePhoto
				url := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(userVK.photo_link))
				photo = tgbotapi.NewPhoto(update.Message.From.ID, url.Media)
				photo.Caption = PrintVkUserForm(userVK)
				photo.ReplyMarkup = GetMatchInlineKeyboard(userVK.profile_link, userVK.post_link)
				ChangeKeyboard(bot, chat_id, matchKeyKeyboard)

			} else {
				msg.Text = NoFormFilledMText
			}

		case NextBText:
			matchPos, err := dbConnection.GetUserMatchPos(user.ID)
			if err != nil {
				log.Fatal(err)
			}
			matchPos++
			if err := dbConnection.SetUserMatchPos(user.ID, matchPos); err != nil {
				log.Fatal(err)
			}

			userVK, err := dbConnection.GetMatchVkUser(user.ID, matchPos)
			if err != nil {
				log.Fatal(err)
			}

			sendType = TypePhoto
			url := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(userVK.photo_link))
			photo = tgbotapi.NewPhoto(update.Message.From.ID, url.Media)
			photo.Caption = PrintVkUserForm(userVK)
			photo.ReplyMarkup = GetMatchInlineKeyboard(userVK.profile_link, userVK.post_link)
			//ChangeKeyboard(bot, chat_id, matchKeyKeyboard)

		case PrevBText:
			matchPos, err := dbConnection.GetUserMatchPos(user.ID)
			if err != nil {
				log.Fatal(err)
			}
			matchPos--
			if matchPos < 0 {
				matchPos = 0
			}
			if err := dbConnection.SetUserMatchPos(user.ID, matchPos); err != nil {
				log.Fatal(err)
			}

			userVK, err := dbConnection.GetMatchVkUser(user.ID, matchPos)
			if err != nil {
				log.Fatal(err)
			}

			sendType = TypePhoto
			url := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(userVK.photo_link))
			photo = tgbotapi.NewPhoto(update.Message.From.ID, url.Media)
			photo.Caption = PrintVkUserForm(userVK)
			photo.ReplyMarkup = GetMatchInlineKeyboard(userVK.profile_link, userVK.post_link)
			//ChangeKeyboard(bot, chat_id, matchKeyKeyboard)

		case MenuBText:

			msg.Text = StartMessage
			msg.ReplyMarkup = mainKeyKeyboard
			if err := dbConnection.SetUserState(user.ID, StateMain); err != nil {
				log.Fatal(err)
			}

		default:
			state, err := dbConnection.GetUserState(user.ID)
			if err != nil {
				log.Fatal(err)
			}

			switch state {
			case StateMain:
			case StateFormFirstName:
				if err := dbConnection.SetFormValue(user.ID, "first_name", message); err != nil {
					log.Fatal(err)
				}
				if err := dbConnection.SetUserState(user.ID, StateFormLastName); err != nil {
					log.Fatal(err)
				}
				msg.Text = EnterLastnameMText

			case StateFormLastName:
				if err := dbConnection.SetFormValue(user.ID, "last_name", message); err != nil {
					log.Fatal(err)
				}
				if err := dbConnection.SetUserState(user.ID, StateFormSex); err != nil {
					log.Fatal(err)
				}
				msg.Text = EnterSexMText
				msg.ReplyMarkup = sexKeyKeyboard

			case StateFormSex:
				sexType := DetectSex(message)
				if sexType == SexUnknown {
					msg.Text = EnterIncorrectFormatText

				} else {
					if err := dbConnection.SetFormValue(user.ID, "sex", strconv.Itoa(int(sexType))); err != nil {
						log.Fatal(err)
					}
					if err := dbConnection.SetUserState(user.ID, StateFormAge); err != nil {
						log.Fatal(err)
					}
					msg.Text = EnterAgeMText
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				}

			case StateFormAge:
				age, err := strconv.Atoi(message)
				if err != nil {
					msg.Text = EnterIncorrectFormatText
				} else {

					if err := dbConnection.SetFormValue(user.ID, "age", strconv.Itoa(age)); err != nil {
						log.Fatal(err)
					}
					if err := dbConnection.SetUserState(user.ID, StateFormRoommateSex); err != nil {
						log.Fatal(err)
					}
					msg.Text = EnterRoommateSexMText
					msg.ReplyMarkup = sexKeyKeyboard
				}

			case StateFormRoommateSex:
				sexType := DetectSex(message)
				if sexType == SexUnknown {
					msg.Text = EnterIncorrectFormatText

				} else {
					if err := dbConnection.SetFormValue(user.ID, "roommate_sex", strconv.Itoa(int(sexType))); err != nil {
						log.Fatal(err)
					}
					if err := dbConnection.SetUserState(user.ID, StateFormApartmentsBudget); err != nil {
						log.Fatal(err)
					}
					msg.Text = EnterApartmentBudgetMText
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				}

			case StateFormApartmentsBudget:
				budget, err := strconv.Atoi(message)
				if err != nil {
					msg.Text = EnterIncorrectFormatText

				} else {

					if err := dbConnection.SetFormValue(user.ID, "apartments_budget", strconv.Itoa(budget)); err != nil {
						log.Fatal(err)
					}
					if err := dbConnection.SetUserState(user.ID, StateFormApartmentsLocation); err != nil {
						log.Fatal(err)
					}
					msg.Text = EnterApartmentLocationMText
				}

			case StateFormApartmentsLocation:
				if err := dbConnection.SetFormValue(user.ID, "apartments_location", message); err != nil {
					log.Fatal(err)
				}
				if err := dbConnection.SetUserState(user.ID, StateFormAboutUser); err != nil {
					log.Fatal(err)
				}
				msg.Text = EnterAboutYouMText

			case StateFormAboutUser:
				if err := dbConnection.SetFormValue(user.ID, "about_user", message); err != nil {
					log.Fatal(err)
				}
				if err := dbConnection.SetUserState(user.ID, StateFormAboutRoommate); err != nil {
					log.Fatal(err)
				}
				msg.Text = EnterAboutRoommateMText

			case StateFormAboutRoommate:
				if err := dbConnection.SetFormValue(user.ID, "about_roommate", message); err != nil {
					log.Fatal(err)
				}
				if err := dbConnection.SetUserState(user.ID, StateMain); err != nil {
					log.Fatal(err)
				}
				msg.Text = EnterSuccessFilledMText
				msg.ReplyMarkup = mainKeyKeyboard
			}
		}

		switch sendType {
		case TypeMessage:
			_, err = bot.Send(msg)
		case TypePhoto:
			photo.ParseMode = tgbotapi.ModeHTML
			_, err = bot.Send(photo)
		}
		if err != nil {
			log.Panic(err)
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
