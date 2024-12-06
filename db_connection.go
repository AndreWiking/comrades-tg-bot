package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

const testID = 863117815

const (
	host   = "localhost"
	port   = 5432
	user   = "super_admin"
	dbname = "postgres"
)

type DbConnection struct {
	db *sql.DB
}

func (connection *DbConnection) Connect() error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)

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

func (connection *DbConnection) IsUserAdded(id int64) (bool, error) {
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

func (connection *DbConnection) AddUser(id int64, username string, first_name string, last_name string) error {

	added, err := connection.IsUserAdded(id)

	if !added && err == nil {
		stmt, err := connection.db.Prepare(
			"INSERT INTO tg_users(id, username, first_name, last_name, state) VALUES( $1, $2, $3, $4, $5)")
		if err != nil {
			return err
		}

		defer stmt.Close()

		if _, err := stmt.Exec(id, username, first_name, last_name, StateMain); err != nil {
			return err
		}

		fmt.Printf("User \"%s\" added\n", username)
	}
	if err != nil {
		return nil
	}
	return nil
}

func (connection *DbConnection) SetUserState(user_id int64, state UserState) error {
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

	return nil
}

func (connection *DbConnection) GetUserState(user_id int64) (UserState, error) {
	stmt, err := connection.db.Prepare("SELECT state FROM tg_users WHERE id = $1")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	var state UserState
	err = stmt.QueryRow(user_id).Scan(&state)
	if err != nil {
		return 0, err
	} else {
		return state, nil
	}
}

func (connection *DbConnection) SetUserMatchPos(user_id int64, match_pos int) error {
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

func (connection *DbConnection) GetUserMatchPos(user_id int64) (int, error) {
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

func (connection *DbConnection) IsFormAdded(user_id int64) (bool, error) {
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

func (connection *DbConnection) AddForm(user_id int64) error {

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
	}
	return nil
}

func (connection *DbConnection) SetFormValue(user_id int64, column string, value string) error {
	stmt, err := connection.db.Prepare(
		"UPDATE tg_form SET date = CURRENT_TIMESTAMP, " + column + " = $1 WHERE user_id = $2")
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

type FormTgUser struct {
	first_name          string
	last_name           string
	sex                 SexType
	age                 int
	roommate_sex        SexType
	apartments_budget   int
	apartments_location string
	about_user          string
	about_roommate      string
}

func (connection *DbConnection) GetForm(user_id int64) (FormTgUser, error) {
	stmt, err := connection.db.Prepare(
		"SELECT first_name, last_name, sex, age, roommate_sex, apartments_budget, apartments_location, about_user, about_roommate FROM tg_form WHERE user_id = $1")
	var form FormTgUser
	if err != nil {
		return form, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(user_id).Scan(&form.first_name, &form.last_name, &form.sex, &form.age, &form.roommate_sex,
		&form.apartments_budget, &form.apartments_location, &form.about_user, &form.about_roommate)

	return form, err
}

func (connection *DbConnection) GetFormText(user_id int64) (string, error) {

	form, err := connection.GetForm(user_id)
	if err != nil {
		return "", err
	} else {
		return fmt.Sprintf(MyFormPatternText, form.first_name, form.last_name, SexTypeName[form.sex], form.age,
			SexTypeName[form.roommate_sex], form.apartments_budget, form.apartments_location, form.about_user,
			form.about_roommate), err
	}
}

type UserVK struct {
	vk_id             int
	fist_name         string
	last_name         string
	sex               SexType
	age               int
	photo_link        string
	post_link         string
	profile_link      string
	apartments_budget int
}

func PrintVkUserForm(user UserVK) string {
	return fmt.Sprintf(VkFormPatternText, user.fist_name, user.last_name, user.age, user.apartments_budget)
}

func (connection *DbConnection) GetLinkVkUserPost(vk_user_id int) (string, error) {
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

func (connection *DbConnection) GetMatchVkUser(user_id int64, matchPos int) (UserVK, error) {

	var userVK UserVK

	user_id = testID

	stmt, err := connection.db.Prepare(`
SELECT user_id, first_name, last_name, sex, age, photo_link, profile_link, apartments_budget
FROM VK_Match
         INNER JOIN VK_Users ON VK_Match.user2_id = VK_Users.id
         INNER JOIN VK_Post ON VK_Match.user2_id = VK_Post.user_id
WHERE user1_id = $1
`)

	if err != nil {
		return userVK, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(user_id)
	if err != nil {
		return userVK, err
	}
	for i := 0; i <= matchPos && rows.Next(); i++ {
		if err := rows.Scan(&userVK.vk_id, &userVK.fist_name, &userVK.last_name, &userVK.sex, &userVK.age, &userVK.photo_link,
			&userVK.profile_link, &userVK.apartments_budget); err != nil {
			return userVK, err
		}
	}

	userVK.post_link, err = connection.GetLinkVkUserPost(userVK.vk_id)

	return userVK, err
}

type PostVK struct {
	user_id               int
	apartments_budget     int
	apartments_location_s float32
	apartments_location_w float32
	roommate_sex          SexType
}

func (connection *DbConnection) GetAllVkPosts() ([]PostVK, error) {

	var posts []PostVK

	rows, err := connection.db.Query(
		`SELECT user_id, apartments_budget, apartments_location_s, apartments_location_w, roommate_sex FROM VK_Post`)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var post PostVK
		if err := rows.Scan(&post.user_id, &post.apartments_budget, &post.apartments_location_s,
			&post.apartments_location_w, &post.roommate_sex); err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}
