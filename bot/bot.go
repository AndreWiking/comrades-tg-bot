package bot

import (
	"ComradesTG/db"
	"ComradesTG/gpt"
	"ComradesTG/matching"
	"ComradesTG/settings"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

type Bot struct {
	bot          *tgbotapi.BotAPI
	dbConnection *db.Connection
	gpt          *gpt.Client
}

func NewBot() *Bot {

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

	return &Bot{bot: bot, dbConnection: &dbConnection, gpt: gpt.NewClient()}
}

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

func (b *Bot) addUserFromVk(vkId int, tgId int64) error {

	if vkUser, err := b.dbConnection.GetVkUser(vkId); err != nil {
		return err
	} else if post, err := b.dbConnection.GetVkUserPost(vkUser.Vk_id); err != nil {
		return err
	} else if err := b.dbConnection.AddFormFromVk(tgId, vkUser, post); err != nil {
		return err
	}
	return nil
}

func (b *Bot) RunUpdates() {

	newUpdate := tgbotapi.NewUpdate(0)
	newUpdate.Timeout = 60
	updates := b.bot.GetUpdatesChan(newUpdate)

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

		start := settings.StartText
		if len(message) > len(start)+1 && message[:len(start)] == start {

			if vkId, err := strconv.Atoi(message[len(start)+1:]); err != nil {
				log.Println(err)
			} else {
				msg.Text = settings.StartMessage
				msg.ReplyMarkup = settings.MainKeyKeyboard
				if err := b.dbConnection.AddUser(user.ID, user.UserName, user.FirstName, user.LastName, settings.UtmVKSpam); err != nil {
					log.Println(err)
				}

				if err := b.addUserFromVk(vkId, user.ID); err != nil {
					log.Println(err)
				}
			}
		}
		switch message {
		case settings.StartText + " " + settings.UtmYa1.String():
			msg.Text = settings.StartMessage
			msg.ReplyMarkup = settings.MainKeyKeyboard
			if err := b.dbConnection.AddUser(user.ID, user.UserName, user.FirstName, user.LastName, settings.UtmYa1); err != nil {
				log.Println(err)
			}

		case settings.StartText:
			msg.Text = settings.StartMessage
			msg.ReplyMarkup = settings.MainKeyKeyboard
			if err := b.dbConnection.AddUser(user.ID, user.UserName, user.FirstName, user.LastName, settings.UtmEmpty); err != nil {
				log.Println(err)
			}
		case settings.AdminPanelText:
			if settings.IsAdmin(user.ID) {
				msg.Text = settings.AdminPanelMessage
				msg.ReplyMarkup = settings.AdminKeyKeyboard
			}
		case settings.AdminFindMatchBText:
			msg.Text = settings.AdminFindMatchMessage
			msg.ReplyMarkup = settings.AdminBackKeyKeyboard
			if err := b.dbConnection.SetUserState(user.ID, settings.StateAdminVkUrlEnter); err != nil {
				log.Println(err)
			}

		case settings.FillFormBText:
			msg.Text = settings.EnterNameMText
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			if err := b.dbConnection.AddForm(user.ID); err != nil {
				log.Println(err)
			}
			if err := b.dbConnection.SetUserState(user.ID, settings.StateFormFirstName); err != nil {
				log.Println(err)
			}
		case settings.AboutUsBText:
			msg.Text = settings.AboutUsAnswerText
		case settings.MyFormBText:
			added, err := b.dbConnection.IsFormAdded(user.ID)
			if err != nil {
				log.Println(err)
			}
			if added {
				text, err := b.dbConnection.GetFormText(user.ID)
				if err != nil {
					log.Println(err)
				}
				msg.Text = text
				msg.ReplyMarkup = settings.MatchParamsKeyKeyboard

			} else {
				msg.Text = settings.NoFormFilledMText
			}
		case settings.EditFormBText:
			added, err := b.dbConnection.IsFormAdded(user.ID)
			if err != nil {
				log.Println(err)
			}
			if added {
				msg.Text = settings.EditFormMText
				msg.ReplyMarkup = settings.EditFromKeyKeyboard
				if err := b.dbConnection.SetUserState(user.ID, settings.StateFormFirstName); err != nil {
					log.Println(err)
				}
			} else {
				msg.Text = settings.NoFormFilledMText
			}

		case settings.FindRoommateBText:
			added, err := b.dbConnection.IsFormAdded(user.ID)
			if err != nil {
				log.Println(err)
			}
			if added {

				if err := b.dbConnection.AddToHistory(user.ID, settings.StateFindRoommate); err != nil {
					log.Println(err)
				}

				SendMessage(b.bot, chat_id, settings.FindRoommateMText)
				if err := matching.MatchGreedy(b.dbConnection, b.gpt, user.ID); err != nil {
					log.Println(err)
				}

				if err := b.dbConnection.SetUserMatchPos(user.ID, 0); err != nil {
					log.Println(err)
				}

				userVK, haveNext, err := b.dbConnection.GetMatchVkUser(user.ID, 0)
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
						ChangeKeyboard(b.bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNoPrev))
					} else {
						ChangeKeyboard(b.bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNoPrevNext))
					}
				}

			} else {
				msg.Text = settings.NoFormFilledMText
			}

		case settings.NextBText:

			if err := b.dbConnection.AddToHistory(user.ID, settings.StateNextRoommate); err != nil {
				log.Println(err)
			}

			matchPos, err := b.dbConnection.GetUserMatchPos(user.ID)
			if err != nil {
				log.Println(err)
			}
			matchPos++
			if err := b.dbConnection.SetUserMatchPos(user.ID, matchPos); err != nil {
				log.Println(err)
			}

			userVK, haveNext, err := b.dbConnection.GetMatchVkUser(user.ID, matchPos)
			if err != nil {
				log.Println(err)
			}

			sendType = TypePhoto
			url := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(userVK.Photo_link))
			photo = tgbotapi.NewPhoto(update.Message.From.ID, url.Media)
			photo.Caption = db.PrintVkUserForm(userVK)
			photo.ReplyMarkup = settings.GetMatchInlineKeyboard(userVK.Profile_link, userVK.Post_link)

			if haveNext {
				ChangeKeyboard(b.bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNormal))
			} else {
				ChangeKeyboard(b.bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNoNext))
			}

		case settings.PrevBText:

			if err := b.dbConnection.AddToHistory(user.ID, settings.StatePrevRoommate); err != nil {
				log.Println(err)
			}

			matchPos, err := b.dbConnection.GetUserMatchPos(user.ID)
			if err != nil {
				log.Println(err)
			}
			matchPos--
			if matchPos < 0 {
				matchPos = 0
			}
			if err := b.dbConnection.SetUserMatchPos(user.ID, matchPos); err != nil {
				log.Println(err)
			}

			userVK, _, err := b.dbConnection.GetMatchVkUser(user.ID, matchPos)
			if err != nil {
				log.Println(err)
			}

			sendType = TypePhoto
			url := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(userVK.Photo_link))
			photo = tgbotapi.NewPhoto(update.Message.From.ID, url.Media)
			photo.Caption = db.PrintVkUserForm(userVK)
			photo.ReplyMarkup = settings.GetMatchInlineKeyboard(userVK.Profile_link, userVK.Post_link)

			if matchPos > 0 {
				ChangeKeyboard(b.bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNormal))
			} else {
				ChangeKeyboard(b.bot, chat_id, settings.MatchKeyboard(settings.MatchKeyboardNoPrev))
			}

		case settings.MenuBText:

			msg.Text = settings.MenuMessage
			msg.ReplyMarkup = settings.MainKeyKeyboard
			if err := b.dbConnection.SetUserState(user.ID, settings.StateMain); err != nil {
				log.Println(err)
			}
		case settings.MatchBudgetBText:

			form, err := b.dbConnection.GetForm(user.ID)
			if err != nil {
				log.Println(err)
			}
			if err := b.dbConnection.SetUserState(user.ID, settings.StateMatchBudget); err != nil {

				log.Println(err)
			}
			msg.Text = fmt.Sprintf(settings.EnterMatchBudgetBText, form.Apartments_budget)
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

		case settings.MatchDistanceBText:

			form, err := b.dbConnection.GetForm(user.ID)
			if err != nil {
				log.Println(err)
			}
			if err := b.dbConnection.SetUserState(user.ID, settings.StateMatchDistance); err != nil {
				log.Println(err)
			}
			msg.Text = fmt.Sprintf(settings.EnterMatchDistanceBText, form.Apartments_location)
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

		default:

			state, err := b.dbConnection.GetUserState(user.ID)
			if err != nil {
				log.Println(err)
			}

			fmt.Println(state)

			//editForm := false
			if state == settings.StateFormEdit {
				state = settings.DetectUserState(message)
				state--
				//editForm = true
			}

			switch state {
			case settings.StateMain:
			case settings.StateFormFirstName:
				if len(message) <= settings.LimitEnterNamesLen {
					if err := b.dbConnection.SetFormValue(user.ID, "first_name", message); err != nil {
						log.Println(err)
					}
					if err := b.dbConnection.SetUserState(user.ID, settings.StateFormLastName); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterLastnameMText
				} else {
					msg.Text = settings.EnterIncorrectFormatText
				}

			case settings.StateFormLastName:
				if len(message) <= settings.LimitEnterNamesLen {
					if err := b.dbConnection.SetFormValue(user.ID, "last_name", message); err != nil {
						log.Println(err)
					}
					if err := b.dbConnection.SetUserState(user.ID, settings.StateFormSex); err != nil {
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
					if err := b.dbConnection.SetFormValue(user.ID, "sex", strconv.Itoa(int(sexType))); err != nil {
						log.Println(err)
					}
					if err := b.dbConnection.SetUserState(user.ID, settings.StateFormAge); err != nil {
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

					if err := b.dbConnection.SetFormValue(user.ID, "age", strconv.Itoa(age)); err != nil {
						log.Println(err)
					}
					if err := b.dbConnection.SetUserState(user.ID, settings.StateFormRoommateSex); err != nil {
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
					if err := b.dbConnection.SetFormValue(user.ID, "roommate_sex", strconv.Itoa(int(sexType))); err != nil {
						log.Println(err)
					}
					if err := b.dbConnection.SetUserState(user.ID, settings.StateFormApartmentsBudget); err != nil {
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

					if err := b.dbConnection.SetFormValue(user.ID, "apartments_budget", strconv.Itoa(budget)); err != nil {
						log.Println(err)
					}
					if err := b.dbConnection.SetUserState(user.ID, settings.StateFormApartmentsLocation); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterApartmentLocationMText
				}

			case settings.StateFormApartmentsLocation:

				locS, locW, err := b.gpt.TransformLocation(message)
				if err != nil || len(message) > settings.LimitEnterTextLen {
					msg.Text = settings.EnterIncorrectFormatText
				} else {

					if err := b.dbConnection.SetFormValue(user.ID, "apartments_location_s", fmt.Sprintf("%f", locS)); err != nil {
						log.Println(err)
					}

					if err := b.dbConnection.SetFormValue(user.ID, "apartments_location_w", fmt.Sprintf("%f", locW)); err != nil {
						log.Println(err)
					}

					if err := b.dbConnection.SetFormValue(user.ID, "apartments_location", message); err != nil {
						log.Println(err)
					}

					if err := b.dbConnection.SetFormValue(user.ID, "apartments_location", message); err != nil {
						log.Println(err)
					}
					if err := b.dbConnection.SetUserState(user.ID, settings.StateFormAboutUser); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterAboutYouMText
				}

			case settings.StateFormAboutUser:
				if len(message) < settings.LimitEnterTextLen {
					if err := b.dbConnection.SetFormValue(user.ID, "about_user", message); err != nil {
						log.Println(err)
					}
					if err := b.dbConnection.SetUserState(user.ID, settings.StateFormAboutRoommate); err != nil {
						log.Println(err)
					}
					msg.Text = settings.EnterAboutRoommateMText
				} else {
					msg.Text = settings.EnterIncorrectFormatText
				}

			case settings.StateFormAboutRoommate:
				if len(message) < settings.LimitEnterTextLen {
					if err := b.dbConnection.SetFormValue(user.ID, "about_roommate", message); err != nil {
						log.Println(err)
					}
					if err := b.dbConnection.SetUserState(user.ID, settings.StateMain); err != nil {
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
					if err := b.dbConnection.SetFormValue(user.ID, "match_distance", fmt.Sprintf("%f", dist)); err != nil {
						log.Println(err)
					}
					form, err := b.dbConnection.GetForm(user.ID)
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
					if err := b.dbConnection.SetFormValue(user.ID, "match_budget", strconv.Itoa(budget)); err != nil {
						log.Println(err)
					}
					form, err := b.dbConnection.GetForm(user.ID)
					if err != nil {
						log.Println(err)
					}
					msg.Text = fmt.Sprintf(settings.EnterMatchBudgetSuccessBText, form.Apartments_budget, budget)
					msg.ReplyMarkup = settings.MainKeyKeyboard
				}

			case settings.StateAdminVkUrlEnter:
				posts, vkId, err := matching.FindMatchVk(b.dbConnection, b.gpt, message)
				if err != nil {
					msg.Text = "Fail: " + err.Error()
				} else {
					var matches, spamMatches strings.Builder
					for i, post := range posts {
						matches.WriteString(post.Link)
						matches.WriteString("\n")
						if i < settings.MaxSpamMessageMatchesCount {
							spamMatches.WriteString(post.Link)
							spamMatches.WriteString("\n")
						}
					}
					msg.Text = "Result for " + message + " :\n" + matches.String()
					if len(msg.Text) > settings.MaxMessageSize {
						msg.Text = msg.Text[:settings.MaxMessageSize]
					}

					botUrl := fmt.Sprintf(settings.BotStartUrlPattern, vkId)
					for i := 0; i < settings.SpamMessageCount; i++ {
						if spamMessage, err := b.gpt.GenerateSpamMessage(spamMatches.String(), botUrl); err != nil {
							log.Println(err)
						} else {
							SendMessage(b.bot, chat_id, spamMessage)
						}
					}
				}
			}
		}
		var err error
		switch sendType {
		case TypeMessage:
			if msg.Text != "" {
				_, err = b.bot.Send(msg)
			}
		case TypePhoto:
			photo.ParseMode = tgbotapi.ModeHTML
			_, err = b.bot.Send(photo)
		}
		if err != nil {
			log.Println(err)
		}
	}
}
