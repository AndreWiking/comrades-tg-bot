package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const TgApiKey = "7541929739:AAFylnUcAeDvSueJGIGQ5kAfow4nEw7P-Oc"

const (
	StartMessage = "Привет! Добро пожаловать в сервис по поиску соседа для совместного съёма жилья.\nМы поможем вам найти идеального соседа для комфортного и дружного проживания."

	FindRoommateBText = "Найти соседа"
	FillFormBText     = "Заполнить анкету"
	MyFormBText       = "Моя анкета"
	AboutUsBText      = "О нас"

	NextBText = "Следующий"
	PrevBText = "Предыдущий"
	MenuBText = "Меню"

	WriteIBText = "Написать"
	MoreIBText  = "Подробнее"

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

	EnterIncorrectFormatText = "Неверный формат, попробуйте ещё раз:"

	AboutUsAnswerText = "Мы предоставляем удобную платформу для тех, кто ищет товарища по квартире, чтобы совместно снять жилье и снизить расходы.\nНаша цель — помочь вам найти идеального сожителя, который соответствует вашим предпочтениям и образу жизни."

	NoFormFilledMText = "Сначала необходимо заполнить"
	FindRoommateMText = "Поиск кандидатов"

	MyFormPatternText = `Имя: <b>%s</b>
Фамилия: <b>%s</b>
Пол: <b>%s</b>
Возраст: <b>%d</b>
Пол соседа: <b>%s</b>
Бюджет: <b>%d₽</b>
Локация квартиры: <b>%s</b>
О себе: <b>%s</b>
Пожелания по соседу: <b>%s</b>`

	VkFormPatternText = `<b>%s %s</b>
Возраст: <b>%d</b>
Бюджет на квартиру: <b>%d₽</b>
`
)

var mainKeyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(FindRoommateBText),
		tgbotapi.NewKeyboardButton(FillFormBText),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(MyFormBText),
		tgbotapi.NewKeyboardButton(AboutUsBText),
	),
)

var matchKeyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(PrevBText),
		tgbotapi.NewKeyboardButton(NextBText),
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
	SexUnknown SexType = 0
	SexFemale  SexType = 1
	SexMale    SexType = 2
)

var SexTypeName = map[SexType]string{
	SexFemale: "Женский",
	SexMale:   "Мужской",
}

var sexKeyKeyboard = tgbotapi.NewReplyKeyboard(
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
)
