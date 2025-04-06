package db

import (
	"ComradesTG/settings"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"
)

const testID = 863117815

const (
	//host     = "localhost"
	host     = "46.17.41.227"
	port     = 5432
	user     = "super_admin"
	dbname   = "postgres"
	password = "gt53_gky94.rtG&xx-rp-ovD"
)

// game12345678
// todo: fix to pointer
type Connection struct {
	db *sql.DB
}

func (connection *Connection) Connect() error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	var err error
	connection.db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	//defer connection.db.Close() todo: close database

	err = connection.db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Database successfully connected!")

	return err
}

func (connection *Connection) IsUserAdded(id int64) (bool, error) {
	stmt, err := connection.db.Prepare("SELECT id FROM tg_users WHERE id = $1")
	if err != nil {
		//log.Println(err)
		return false, err
	}

	defer stmt.Close()

	var idTmp int64
	err = stmt.QueryRow(id).Scan(&idTmp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			//log.Println(err)
			return false, err
		}
	} else {
		return true, nil
	}
}

func (connection *Connection) AddUser(id int64, username string, first_name string, last_name string, utm settings.UserUtm) error {

	added, err := connection.IsUserAdded(id)

	if !added && err == nil {
		stmt, err := connection.db.Prepare(
			"INSERT INTO tg_users(id, username, first_name, last_name, state, utm) VALUES( $1, $2, $3, $4, $5, $6)")
		if err != nil {
			return err
		}

		defer stmt.Close()

		if _, err := stmt.Exec(id, username, first_name, last_name, settings.StateMain, utm.String()); err != nil {
			return err
		}

		fmt.Printf("User \"%s\" added\n", username)
	}
	if err != nil {
		return nil
	}
	return nil
}

func (connection *Connection) AddToHistory(user_id int64, state settings.UserState) error {
	stmt, err := connection.db.Prepare(
		"INSERT INTO tg_history(user_id, state, description) VALUES( $1, $2, $3 )")
	if err != nil {
		return err
	}

	defer stmt.Close()

	if _, err := stmt.Exec(user_id, state, state.Description()); err != nil {
		return err
	}
	return nil
}

func (connection *Connection) SetUserState(user_id int64, state settings.UserState) error {
	stmt, err := connection.db.Prepare(
		"UPDATE tg_users SET state = $1 WHERE id = $2")
	if err != nil {
		return err
	}

	defer stmt.Close()

	if _, err := stmt.Exec(state, user_id); err != nil {
		return err
	}

	fmt.Printf("User %d set state %d\n", user_id, state)

	return connection.AddToHistory(user_id, state)
}

func (connection *Connection) GetUserState(user_id int64) (settings.UserState, error) {
	stmt, err := connection.db.Prepare("SELECT state FROM tg_users WHERE id = $1")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	var state settings.UserState
	err = stmt.QueryRow(user_id).Scan(&state)
	if err != nil {
		return 0, err
	} else {
		return state, nil
	}
}

func (connection *Connection) SetUserMatchPos(user_id int64, match_pos int) error {
	stmt, err := connection.db.Prepare(
		"UPDATE tg_users SET match_pos = $1 WHERE id = $2")
	if err != nil {
		return err
	}

	defer stmt.Close()

	if _, err := stmt.Exec(match_pos, user_id); err != nil {
		return err
	}

	fmt.Printf("User %d set match pos %d\n", user_id, match_pos)

	return nil
}

func (connection *Connection) GetUserMatchPos(user_id int64) (int, error) {
	stmt, err := connection.db.Prepare("SELECT match_pos FROM tg_users WHERE id = $1")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	var matchPos int
	err = stmt.QueryRow(user_id).Scan(&matchPos)
	if err != nil {
		return 0, err
	} else {
		return matchPos, nil
	}
}

func (connection *Connection) IsFormAdded(user_id int64) (bool, error) {
	stmt, err := connection.db.Prepare("SELECT user_id FROM tg_form WHERE user_id = $1")
	if err != nil {
		return false, err
	}

	defer stmt.Close()

	var idTmp int64
	err = stmt.QueryRow(user_id).Scan(&idTmp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

func (connection *Connection) SetFormValue(user_id int64, column string, value string) error {
	stmt, err := connection.db.Prepare(
		"UPDATE tg_form SET Date = CURRENT_TIMESTAMP, " + column + " = $1 WHERE user_id = $2")
	if err != nil {
		return err
	}

	defer stmt.Close()

	if _, err := stmt.Exec(value, user_id); err != nil {
		return err
	}

	fmt.Printf("For user %d set form value %s = %s\n", user_id, column, value)

	return nil
}

func (connection *Connection) AddForm(user_id int64) error {

	added, err := connection.IsFormAdded(user_id)

	if !added && err == nil {
		stmt, err := connection.db.Prepare(
			"INSERT INTO tg_form(user_id) VALUES( $1 )")
		if err != nil {
			return err
		}

		defer stmt.Close()

		if _, err := stmt.Exec(user_id); err != nil {
			return err
		}

		fmt.Printf("Form for \"%d\" added\n", user_id)
	}
	if err != nil {
		return nil
	} //todo: understand why
	return nil
}

func (connection *Connection) AddFormFromVk(user_id int64, user UserVK, post PostVK) error {

	added, err := connection.IsFormAdded(user_id)

	if !added && err == nil {
		stmt, err := connection.db.Prepare(
			`INSERT INTO tg_form(user_id, first_name, last_name, sex, age, roommate_sex, apartments_budget, 
                    apartments_location, apartments_location_s, apartments_location_w, about_user, about_roommate)
VALUES( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`)

		if err != nil {
			return err
		}

		defer stmt.Close()

		if _, err := stmt.Exec(user_id, user.fist_name, user.last_name, user.Sex, user.age, post.Roommate_sex,
			post.Apartments_budget, "", post.Apartments_location_s, post.Apartments_location_w, "", ""); err != nil {
			return err
		} // todo: add apartments_location

		fmt.Printf("(from VK) Form for \"%d\" added\n", user_id)
	}
	if err != nil {
		return nil
	}
	return nil
}

type FormTgUser struct {
	first_name            string
	last_name             string
	Sex                   settings.SexType
	age                   int
	Roommate_sex          settings.SexType
	Apartments_budget     int
	Apartments_location   string
	Apartments_location_s float64
	Apartments_location_w float64
	about_user            string
	about_roommate        string
	Match_budget          int
	Match_distance        float64
}

func (connection *Connection) GetForm(user_id int64) (FormTgUser, error) {
	stmt, err := connection.db.Prepare(
		`SELECT first_name, last_name, sex, age, roommate_sex, apartments_budget, apartments_location,
       apartments_location_s, apartments_location_w, about_user, about_roommate, match_budget, match_distance FROM tg_form WHERE user_id = $1`)
	var form FormTgUser
	if err != nil {
		return form, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(user_id).Scan(&form.first_name, &form.last_name, &form.Sex, &form.age, &form.Roommate_sex,
		&form.Apartments_budget, &form.Apartments_location, &form.Apartments_location_s, &form.Apartments_location_w,
		&form.about_user, &form.about_roommate, &form.Match_budget, &form.Match_distance)

	return form, err
}

func (connection *Connection) GetFormText(user_id int64) (string, error) {

	form, err := connection.GetForm(user_id)
	if err != nil {
		return "", err
	} else {
		return fmt.Sprintf(settings.MyFormPatternText, form.first_name, form.last_name, settings.SexTypeName[form.Sex],
			form.age, settings.SexTypeName[form.Roommate_sex], form.Apartments_budget, form.Match_budget,
			form.Apartments_location, form.Match_distance, form.about_user,
			form.about_roommate), err
	}
}

type UserVK struct {
	Vk_id             int
	fist_name         string
	last_name         string
	Sex               settings.SexType
	age               int
	Photo_link        string
	Post_link         string
	Profile_link      string
	apartments_budget int
}

func PrintVkUserForm(user UserVK) string {
	var age string
	if user.age == 0 {
		age = "—"
	} else {
		age = strconv.Itoa(user.age)
	}
	return fmt.Sprintf(settings.VkFormPatternText, user.fist_name, user.last_name, age, user.apartments_budget)
}

func (connection *Connection) GetLinkVkUserPost(vk_user_id int) (string, error) {
	stmt, err := connection.db.Prepare("SELECT link FROM vk_post WHERE user_id = $1")
	if err != nil {
		return "", err
	}

	defer stmt.Close()
	var link string
	err = stmt.QueryRow(vk_user_id).Scan(&link)
	if err != nil {
		return "", err
	} else {
		return link, nil
	}
}

func (connection *Connection) GetMatchVkUser(tgUserId int64, matchPos int) (UserVK, bool, error) {

	var userVK UserVK
	stmt, err := connection.db.Prepare(`select vk_user_id from tg_match where tg_user_id = $1`)

	if err != nil {
		return userVK, false, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(tgUserId)
	if err != nil {
		return userVK, false, err
	}
	var vkUserId int
	for i := 0; i <= matchPos && rows.Next(); i++ {
		if err := rows.Scan(&vkUserId); err != nil {
			return userVK, false, err
		}
	}
	haveNext := rows.Next()

	userVK, err = connection.GetVkUser(vkUserId)
	if err != nil {
		return userVK, haveNext, err
	}
	post, err := connection.GetVkUserPost(vkUserId)
	if err != nil {
		return userVK, haveNext, err
	}
	userVK.apartments_budget = post.Apartments_budget
	userVK.Post_link = post.Link

	return userVK, haveNext, nil
}

type PostVK struct {
	User_id               int
	Apartments_budget     int
	Apartments_location_s float64
	Apartments_location_w float64
	Roommate_sex          settings.SexType
	Link                  string
	Sex                   settings.SexType
	Date                  string
}

func (connection *Connection) GetVkUserPost(user_id int) (PostVK, error) {

	var post PostVK

	stmt, err := connection.db.Prepare(`
SELECT user_id, apartments_budget, apartments_location_s, apartments_location_w, roommate_sex, link, date FROM VK_Post WHERE user_id = $1`)

	if err != nil {
		return post, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(user_id).Scan(&post.User_id, &post.Apartments_budget, &post.Apartments_location_s,
		&post.Apartments_location_w, &post.Roommate_sex, &post.Link, &post.Date)

	return post, err
}

//select link, vk_users.sex
//FROM vk_post
//INNER JOIN VK_Users ON vk_post.user_id = VK_Users.id;

func (connection *Connection) GetAllVkPosts() ([]PostVK, error) {

	var posts []PostVK

	rows, err := connection.db.Query(
		`SELECT user_id, apartments_budget, apartments_location_s, apartments_location_w, roommate_sex, link, vk_users.sex, date 
FROM VK_Post INNER JOIN VK_Users ON vk_post.user_id = VK_Users.id`)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var post PostVK
		if err := rows.Scan(&post.User_id, &post.Apartments_budget, &post.Apartments_location_s,
			&post.Apartments_location_w, &post.Roommate_sex, &post.Link, &post.Sex, &post.Date); err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (connection *Connection) GetVkUser(user_id int) (UserVK, error) {

	var userVK UserVK

	stmt, err := connection.db.Prepare(`
SELECT first_name, last_name, Sex, age, photo_link, profile_link FROM vk_users WHERE id = $1`)

	if err != nil {
		return userVK, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(user_id).Scan(&userVK.fist_name, &userVK.last_name, &userVK.Sex,
		&userVK.age, &userVK.Photo_link, &userVK.Profile_link)

	userVK.Vk_id = user_id

	return userVK, err
}

func (connection *Connection) AddTgMatch(tg_user_id int64, vk_user_id int) error {

	stmt, err := connection.db.Prepare(
		"INSERT INTO tg_match(tg_user_id, vk_user_id) VALUES($1, $2)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	if _, err := stmt.Exec(tg_user_id, vk_user_id); err != nil {
		return err
	}

	fmt.Printf("Tg match %d - %d added\n", tg_user_id, vk_user_id)

	return nil
}

func (connection *Connection) DeleteTgMatch(tg_user_id int64) error {

	stmt, err := connection.db.Prepare(
		"DELETE FROM tg_match WHERE tg_user_id = $1")
	if err != nil {
		return err
	}

	defer stmt.Close()

	if _, err := stmt.Exec(tg_user_id); err != nil {
		return err
	}

	return nil
}
