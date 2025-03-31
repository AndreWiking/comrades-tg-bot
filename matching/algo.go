package matching

import (
	"ComradesTG/db"
	"ComradesTG/settings"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
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

const (
	matchDistanceDefault = 2.01
	matchBudgetDefault   = 5000
)

type matchUser struct {
	ApartmentsBudget    int
	ApartmentsLocationS float64
	ApartmentsLocationW float64
	Sex                 settings.SexType
	RoommateSex         settings.SexType
}

func vkPostToMatchUser(post db.PostVK) (matchUser, error) {
	return matchUser{
		ApartmentsBudget:    post.Apartments_budget,
		ApartmentsLocationS: post.Apartments_location_w,
		ApartmentsLocationW: post.Apartments_location_w,
		Sex:                 post.Sex,
		RoommateSex:         post.Roommate_sex,
	}, nil
}

func distance(lat1, lon1, lat2, lon2 float64) float64 {
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

//func isMatch(connection *db.Connection, post db.PostVK, tgForm db.FormTgUser) (bool, error) {
//	if post.Apartments_location_s == 0 || post.Apartments_location_w == 0 || post.Apartments_budget == 0 {
//		return false, nil
//	}
//	dist := distance(post.Apartments_location_s, post.Apartments_location_w, tgForm.Apartments_location_s, tgForm.Apartments_location_w)
//
//	budget := math.Abs(float64(post.Apartments_budget-tgForm.Apartments_budget)) <= float64(tgForm.Match_budget)
//
//	vkUser, err := connection.GetVkUser(post.User_id)
//	if err != nil {
//		return false, err
//	}
//
//	sexMatch := (post.Roommate_sex == tgForm.Sex || post.Roommate_sex == settings.SexUnknown) &&
//		(tgForm.Roommate_sex == vkUser.Sex || tgForm.Roommate_sex == settings.SexUnknown)
//
//	return dist < tgForm.Match_distance && budget && sexMatch, nil
//}

func isMatch(user1 matchUser, user2 matchUser, matchDistance float64, matchBudget int) bool {
	if user1.ApartmentsLocationS == 0 || user1.ApartmentsLocationW == 0 || user1.ApartmentsBudget == 0 {
		return false
	}
	dist := distance(user1.ApartmentsLocationS, user1.ApartmentsLocationW, user2.ApartmentsLocationS, user2.ApartmentsLocationW)

	budget := math.Abs(float64(user1.ApartmentsBudget-user2.ApartmentsBudget)) <= float64(matchBudget)

	sexMatch := (user1.RoommateSex == user2.Sex || user1.RoommateSex == settings.SexUnknown) &&
		(user2.RoommateSex == user1.Sex || user2.RoommateSex == settings.SexUnknown)

	return dist < matchDistance && budget && sexMatch
}

func match(user1 matchUser, user2 matchUser, matchDistance float64, matchBudget int) (bool, float64) {
	if user1.ApartmentsLocationS == 0 || user1.ApartmentsLocationW == 0 || user1.ApartmentsBudget == 0 {
		return false, 0
	}
	dist := distance(user1.ApartmentsLocationS, user1.ApartmentsLocationW, user2.ApartmentsLocationS, user2.ApartmentsLocationW)

	budget := math.Abs(float64(user1.ApartmentsBudget-user2.ApartmentsBudget)) <= float64(matchBudget)

	sexMatch := (user1.RoommateSex == user2.Sex || user1.RoommateSex == settings.SexUnknown) &&
		(user2.RoommateSex == user1.Sex || user2.RoommateSex == settings.SexUnknown)

	if dist < matchDistance && budget && sexMatch {
		return true, dist
	} else {
		return false, 0
	}
}

func MatchGreedy(connection *db.Connection, tgUserId int64) error {

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

	matchedPairs := make([]pair, 0)
	for _, post := range posts {

		user1, err := vkPostToMatchUser(post)
		if err != nil {
			return err
		}

		user2 := matchUser{
			ApartmentsBudget:    tgForm.Apartments_budget,
			ApartmentsLocationS: tgForm.Apartments_location_w,
			ApartmentsLocationW: tgForm.Apartments_location_w,
			Sex:                 tgForm.Sex,
			RoommateSex:         tgForm.Roommate_sex,
		}

		if isMatch, dist := match(user1, user2, tgForm.Match_distance, tgForm.Match_budget); isMatch {
			matchedPairs = append(matchedPairs, pair{post, dist})
		}
	}

	for _, post := range sortPairs(matchedPairs) {
		if err := connection.AddTgMatch(tgUserId, post.User_id); err != nil {
			fmt.Println(err.Error()) // todo: fix
			//return err
		}
	}

	return nil
}

func comparePostLink(url1 string, url2 string) bool {
	pattern := "wall-"
	index1 := strings.Index(url1, pattern)
	index2 := strings.Index(url2, pattern)
	if index1 == -1 || index2 == -1 {
		return false
	}
	return url1[index1:] == url2[index2:]
}

type pair struct {
	post db.PostVK
	dist float64
}

func sortPairs(pairs []pair) []db.PostVK {
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].dist < pairs[j].dist
	})
	res := make([]db.PostVK, 0)
	for _, mPair := range pairs {
		res = append(res, mPair.post)
	}
	return res
}

func FindMatchVk(connection *db.Connection, vkPostLink string) ([]db.PostVK, int, error) {

	posts, err := connection.GetAllVkPosts()
	if err != nil {
		return nil, 0, err
	}

	selectedPost := (*db.PostVK)(nil)
	for _, post := range posts {
		if comparePostLink(post.Link, vkPostLink) {
			selectedPost = &post
			break
		}
	}
	if selectedPost == nil {
		return nil, 0, errors.New("no post link found")
	}

	matchedPairs := make([]pair, 0)
	for _, post := range posts {
		if post.User_id == selectedPost.User_id {
			continue
		}

		user1, err := vkPostToMatchUser(*selectedPost)
		if err != nil {
			return nil, 0, err
		}
		user2, err := vkPostToMatchUser(post)
		if err != nil {
			return nil, 0, err
		}

		if isMatch, dist := match(user1, user2, matchDistanceDefault, matchBudgetDefault); isMatch {
			matchedPairs = append(matchedPairs, pair{post, dist})
		}
	}

	return sortPairs(matchedPairs), selectedPost.User_id, nil
}

//func distance(post1 db.PostVK, post2 db.PostVK) float64 {
//	return math.Sqrt(math.Pow(float64(post1.Apartments_location_s-post2.Apartments_location_s), 2) +
//		math.Pow(float64(post1.Apartments_location_w-post2.Apartments_location_w), 2))
//}
//
//func isMatch(connection db.Connection, post1 db.PostVK, post2 db.PostVK) (bool, error) {
//	if post1.Apartments_location_s == 0 || post1.Apartments_location_w == 0 || post2.Apartments_location_s == 0 || post2.Apartments_location_w == 0 {
//		return false, nil
//	}
//	dist := distance(post1, post2)
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
