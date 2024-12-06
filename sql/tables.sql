DROP TABLE IF EXISTS TG_Users CASCADE;
CREATE TABLE TG_Users
(
    id         BIGINT PRIMARY KEY,
    username   VARCHAR(128),
    first_name VARCHAR(128),
    last_name  VARCHAR(128),
    state      INT,
    match_pos  INT
);

DROP TABLE IF EXISTS TG_Form CASCADE;
CREATE TABLE TG_Form
(
    id                  SERIAL PRIMARY KEY,
    user_id             INT REFERENCES TG_Users (id) UNIQUE,
    first_name          VARCHAR(128),
    last_name           VARCHAR(128),
    sex                 INT,
    age                 INT,
    roommate_sex        INT,
    apartments_budget   INTEGER,
    apartments_location VARCHAR(1024),
    about_user          TEXT,
    about_roommate      TEXT,
    date                TIMESTAMP
);

DROP TABLE IF EXISTS TG_Match CASCADE;
CREATE TABLE TG_Match
(
    id         SERIAL PRIMARY KEY,
    tg_user_id BIGINT REFERENCES TG_Users (id),
    vk_user_id INT REFERENCES VK_Users (id)
);

select first_name, last_name, sex, age, photo_link, profile_link, apartments_budget
FROM vk_match
         INNER JOIN VK_Users ON vk_match.user2_id = VK_Users.id
         INNER JOIN vk_post ON vk_match.user2_id = vk_post.user_id
where user1_id = 384436474;