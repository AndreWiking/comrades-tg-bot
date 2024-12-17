package matching

import (
	"ComradesTG/db"
	"ComradesTG/settings"
	"fmt"
	"math"
	//"ComradesTG/settings"
	//"math"
	//"strconv"
	//"strings"
)

//const (
//	budgetDiff   = 0.15
//	moscowDist   = 0.34213203515748
//	locationDiff = moscowDist / 15.0
//)

// 37.849  55.908515 , 37.543571 : 55.573800 , 37.653483

func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	// Преобразование градусов в радианы
	lat1 = lat1 * math.Pi / 180
	lon1 = lon1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180
	lon2 = lon2 * math.Pi / 180

	// Радиус Земли (в километрах)
	const R = 6371

	// Вычисление разницы в широте и долготе
	dLat := lat2 - lat1
	dLon := lon2 - lon1

	// Вычисление расстояния
	a := math.Pow(math.Sin(dLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func isMatch(connection db.Connection, post db.PostVK, tgForm db.FormTgUser) (bool, error) {
	if post.Apartments_location_s == 0 || post.Apartments_location_w == 0 || post.Apartments_budget == 0 {
		return false, nil
	}
	dist := Distance(post.Apartments_location_s, post.Apartments_location_w, tgForm.Apartments_location_s, tgForm.Apartments_location_w)

	budget := math.Abs(float64(post.Apartments_budget-tgForm.Apartments_budget)) <= float64(tgForm.Match_budget)

	vkUser, err := connection.GetVkUser(post.User_id)
	if err != nil {
		return false, err
	}

	sexMatch := (post.Roommate_sex == tgForm.Sex || post.Roommate_sex == settings.SexUnknown) &&
		(tgForm.Roommate_sex == vkUser.Sex || tgForm.Roommate_sex == settings.SexUnknown)

	return dist < tgForm.Match_distance && budget && sexMatch, nil
}

func MatchGreedy(connection db.Connection, tgUserId int64) error {

	tgForm, err := connection.GetForm(tgUserId)
	if err != nil {
		return err
	}

	posts, err := connection.GetAllVkPosts()
	if err != nil {
		return err
	}

	if err := connection.DeleteTgMatch(tgUserId); err != nil {
		return err
	}

	for _, post := range posts {
		match, err := isMatch(connection, post, tgForm)
		if err != nil {
			return err
		}
		if match {
			if err := connection.AddTgMatch(tgUserId, post.User_id); err != nil {
				fmt.Println(err.Error()) // todo: fix
				//return err
			}
		}
	}

	return nil
}

//func Distance(post1 db.PostVK, post2 db.PostVK) float64 {
//	return math.Sqrt(math.Pow(float64(post1.Apartments_location_s-post2.Apartments_location_s), 2) +
//		math.Pow(float64(post1.Apartments_location_w-post2.Apartments_location_w), 2))
//}
//
//func isMatch(connection db.Connection, post1 db.PostVK, post2 db.PostVK) (bool, error) {
//	if post1.Apartments_location_s == 0 || post1.Apartments_location_w == 0 || post2.Apartments_location_s == 0 || post2.Apartments_location_w == 0 {
//		return false, nil
//	}
//	dist := Distance(post1, post2)
//
//	budget := math.Abs(float64(post1.Apartments_budget)-float64(post2.Apartments_budget)) /
//		(float64(post1.Apartments_budget) + float64(post2.Apartments_budget))
//
//	user1, err := connection.GetVkUser(post1.User_id)
//	if err != nil {
//		return false, err
//	}
//	user2, err := connection.GetVkUser(post1.User_id)
//	if err != nil {
//		return false, err
//	}
//
//	sexMatch := (post1.Roommate_sex == user1.Sex || post1.Roommate_sex == settings.SexUnknown) &&
//		(post2.Roommate_sex == user2.Sex || post2.Roommate_sex == settings.SexUnknown)
//
//	return dist < locationDiff && budget < budgetDiff && sexMatch, nil
//}
