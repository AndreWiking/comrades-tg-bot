package settings

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"slices"
)

const MaxMessageSize = 4096

const (
	StartMessage      = "Привет! Добро пожаловать в сервис по поиску соседа для совместного съёма жилья.\nМы поможем вам найти идеального соседа для комфортного и дружного проживания."
	MenuMessage       = "Мы поможем найти идеального соседа для комфортного и дружного проживания."
	AdminPanelMessage = "Admin panel"

	AdminPanelText    = "/admin"
	StartText         = "/start"
	FindRoommateBText = "Найти соседа"
	FillFormBText     = "Заполнить анкету"
	MyFormBText       = "Моя анкета"
	AboutUsBText      = "О нас"

	NextBText = "Следующий"
	PrevBText = "Предыдущий"
	MenuBText = "Меню"

	WriteIBText = "Написать"
	MoreIBText  = "Подробнее"

	AdminFindMatchBText   = "Найти соседа для VK"
	AdminFindMatchMessage = "Введите url поста пользователя:"

	MatchDistanceBText = "Расстояние поиска"
	MatchBudgetBText   = "Вилка бюджета"
	EditFormBText      = "Изменить анкету"

	EditFormMText               = "Выберите поле для изменения"
	EnterNameMText              = "Введите ваше имя:"
	EnterLastnameMText          = "Ваша фамилия:"
	EnterAgeMText               = "Сколько вам лет?"
	EnterSexMText               = "Ваш пол:"
	EnterRoommateSexMText       = "Пол желаемого соседа:"
	EnterApartmentLocationMText = "Где планируете снимать квартиру?"
	EnterApartmentBudgetMText   = "Какой у вас бюджет(в ₽)?"
	EnterAboutYouMText          = "Расскажите немного о себе, то что может быть важно для вашего потенциального соседа"
	EnterAboutRoommateMText     = "Какие у вас пожелания по соседу?"
	EnterSuccessFilledMText     = "Анкета успешно заполнена! Можете приступать к поиску соседа."

	EnterMatchDistanceBText = "Введите приемлемое для вас расстояние(в км)\n\n%s ± __км"
	EnterMatchBudgetBText   = "Введите приемлемую для вас разницу в бюджете(в ₽)\n\n%d₽ ± __₽"

	EnterMatchDistanceSuccessBText = "Расстояние успешно изменено: %s ± %.1fкм.\nМожете приступать к поиску соседа."
	EnterMatchBudgetSuccessBText   = "Разница в бюджете успешно изменена: %d₽ ± %d₽.\nМожете приступать к поиску соседа."

	EnterIncorrectFormatText = "Неверный формат, попробуйте ещё раз:"

	AboutUsAnswerText = "Мы предоставляем удобную платформу для тех, кто ищет товарища по квартире, чтобы совместно снять жилье и снизить расходы.\nНаша цель — помочь вам найти идеального сожителя, который соответствует вашим предпочтениям и образу жизни."

	NoFormFilledMText = "Сначала необходимо заполнить анкету"
	NoMatchFound      = "К сожалению, не нашлось ни одного кандидата.\nПоменяйте параметры поиска в анкете и попробуйте ещё раз."
	FindRoommateMText = "Поиск кандидатов ..."

	MyFormPatternText = `Имя: <b>%s</b>
Фамилия: <b>%s</b>
Пол: <b>%s</b>
Возраст: <b>%d</b>
Пол соседа: <b>%s</b>
Бюджет: <b>%d₽</b> ± %d₽
Локация квартиры: <b>%s</b> ± %.1fкм
О себе: <b>%s</b>
Пожелания по соседу: <b>%s</b>`

	VkFormPatternText = `<b>%s %s</b>
Возраст: <b>%s</b>
Бюджет на квартиру: <b>%d₽</b>
`
	BotStartUrlPattern = "https://t.me/find_comrade_bot?start=%d"

	SpamMessagePattern = `Здравствуйте! Мы заметили, что вы ищете соседа для съема квартиры. 
Мы подобрали вам возможных кандидатов, вот некоторые из них: %s 
Если хотите еще предложений, переходите в наш телеграмм бот: %s`

	MaxSpamMessageMatchesCount = 3
	SpamMessageCount           = 3
)

var adminIDList = []int64{681591950, 7291028590, 959853862}

func IsAdmin(id int64) bool {
	return slices.Contains(adminIDList, id)
}

var AdminKeyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(AdminFindMatchBText),
	),
)

var AdminBackKeyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(AdminPanelText),
		tgbotapi.NewKeyboardButton(MenuBText),
	),
)

var MainKeyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(FindRoommateBText),
		tgbotapi.NewKeyboardButton(FillFormBText),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(MyFormBText),
		tgbotapi.NewKeyboardButton(AboutUsBText),
	),
)

type MatchKeyKeyboardType int

const (
	MatchKeyboardNormal MatchKeyKeyboardType = iota
	MatchKeyboardNoPrev
	MatchKeyboardNoNext
	MatchKeyboardNoPrevNext
)

func MatchKeyboard(keyboardType MatchKeyKeyboardType) tgbotapi.ReplyKeyboardMarkup {

	switch keyboardType {
	case MatchKeyboardNormal:
		return tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(PrevBText),
				tgbotapi.NewKeyboardButton(NextBText),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(MenuBText),
			),
		)
	case MatchKeyboardNoPrev:
		return tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(NextBText),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(MenuBText),
			),
		)
	case MatchKeyboardNoNext:
		return tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(PrevBText),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(MenuBText),
			),
		)
	case MatchKeyboardNoPrevNext:
		return tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(MenuBText),
			),
		)
	default:
		return tgbotapi.NewReplyKeyboard()
	}
}

var MatchParamsKeyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(MatchBudgetBText),
		tgbotapi.NewKeyboardButton(MatchDistanceBText),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(EditFormBText),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(MenuBText),
	),
)

var EditFromKeyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(StateFormFirstName.String()),
		tgbotapi.NewKeyboardButton(StateFormLastName.String()),
		tgbotapi.NewKeyboardButton(StateFormAge.String()),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(StateFormSex.String()),
		tgbotapi.NewKeyboardButton(StateFormRoommateSex.String()),
		tgbotapi.NewKeyboardButton(StateFormApartmentsLocation.String()),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(StateFormApartmentsBudget.String()),
		tgbotapi.NewKeyboardButton(StateFormAboutUser.String()),
		tgbotapi.NewKeyboardButton(StateFormAboutRoommate.String()),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(MenuBText),
	),
)

func GetMatchInlineKeyboard(writeLink string, profileLink string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(WriteIBText, writeLink),
			tgbotapi.NewInlineKeyboardButtonURL(MoreIBText, profileLink),
		),
	)
}

type SexType int

const (
	SexFemale  SexType = 1
	SexMale    SexType = 2
	SexUnknown SexType = 0
)

var SexTypeName = map[SexType]string{
	SexFemale: "Женский",
	SexMale:   "Мужской",
}

var SexKeyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(SexTypeName[SexMale]),
		tgbotapi.NewKeyboardButton(SexTypeName[SexFemale]),
	),
)

func DetectSex(sex string) SexType {
	switch sex {
	case SexTypeName[SexFemale]:
		return SexFemale
	case SexTypeName[SexMale]:
		return SexMale
	default:
		return SexUnknown
	}
}

type UserState int

const (
	StateMain UserState = iota
	StateFormFirstName
	StateFormLastName
	StateFormSex
	StateFormAge
	StateFormRoommateSex
	StateFormApartmentsBudget
	StateFormApartmentsLocation
	StateFormAboutUser
	StateFormAboutRoommate
	StateMatchDistance
	StateMatchBudget
	StateFormEdit
	StateAdminVkUrlEnter
	StateFindRoommate
	StatePrevRoommate
	StateNextRoommate
	StateFormUnknown UserState = -1
)

var userStateName = map[UserState]string{
	StateFormFirstName:          "Имя",
	StateFormLastName:           "Фамилия",
	StateFormAge:                "Возраст",
	StateFormSex:                "Пол",
	StateFormRoommateSex:        "Пол соседа",
	StateFormApartmentsLocation: "Локация квартиры",
	StateFormApartmentsBudget:   "Бюджет",
	StateFormAboutUser:          "О себе",
	StateFormAboutRoommate:      "О соседе",
}

var userStateDesc = map[UserState]string{
	StateMain:                   "Main",
	StateFormFirstName:          "Form First Name",
	StateFormLastName:           "Form Last Name",
	StateFormSex:                "Form Sex",
	StateFormAge:                "Form Age",
	StateFormRoommateSex:        "Form Roommate Sex",
	StateFormApartmentsBudget:   "Form Apartments Budget",
	StateFormApartmentsLocation: "Form Apartments Location",
	StateFormAboutUser:          "Form About User",
	StateFormAboutRoommate:      "Form About Roommate",
	StateMatchDistance:          "Match Distance",
	StateMatchBudget:            "Match Budget",
	StateFormEdit:               "Form Edit",
	StateAdminVkUrlEnter:        "Admin Vk Url Enter",
	StateFindRoommate:           "Find Roommate",
	StatePrevRoommate:           "Prev Roommate",
	StateNextRoommate:           "Next Roommate",
	StateFormUnknown:            "Form Unknown",
}

func (us UserState) String() string {
	return userStateName[us]
}

func (us UserState) Description() string {
	return userStateDesc[us]
}

func DetectUserState(name string) UserState {
	for s, n := range userStateName {
		if n == name {
			return s
		}
	}
	return StateFormUnknown
}

type UserUtm int

const (
	UtmEmpty UserUtm = iota
	UtmYa1
	UtmVKSpam
	UtmVk1
)

var userUtmName = map[UserUtm]string{
	UtmEmpty:  "",
	UtmYa1:    "ya1",
	UtmVKSpam: "vk_spam",
	UtmVk1:    "vk1",
}

func (u UserUtm) String() string {
	return userUtmName[u]
}

const (
	LimitBudget        = 10000000
	LimitMatchDist     = 10000000.0
	LimitEnterNamesLen = 127
	LimitEnterTextLen  = 1027
)
