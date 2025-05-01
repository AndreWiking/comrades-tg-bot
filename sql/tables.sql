DROP TABLE IF EXISTS TG_Users CASCADE;
CREATE TABLE TG_Users
(
    id         BIGINT PRIMARY KEY,
    username   VARCHAR(128),
    first_name VARCHAR(128),
    last_name  VARCHAR(128),
    state      INT,
    match_pos  INT,
    utm        VARCHAR(128)
);

-- ALTER TABLE TG_Users ADD utm VARCHAR(128);

DROP TABLE IF EXISTS TG_Form CASCADE;
CREATE TABLE TG_Form
(
    id                    SERIAL PRIMARY KEY,
    user_id               INT REFERENCES TG_Users (id) UNIQUE,
    first_name            VARCHAR(128),
    last_name             VARCHAR(128),
    sex                   INT,
    age                   INT,
    roommate_sex          INT,
    apartments_budget     INTEGER,
    apartments_location   VARCHAR(1024),
    apartments_location_s FLOAT,
    apartments_location_w FLOAT,
    about_user            TEXT      DEFAULT '',
    about_roommate        TEXT      DEFAULT '',
    date                  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    match_budget          FLOAT     DEFAULT 5000,
    match_distance        FLOAT     DEFAULT 2.0
);

DROP TABLE IF EXISTS TG_History CASCADE;
CREATE TABLE TG_History
(
    id          SERIAL PRIMARY KEY,
    user_id     INT REFERENCES TG_Users (id),
    state       INT       DEFAULT -1,
    description VARCHAR(128),
    date        TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS TG_Match CASCADE;
CREATE TABLE TG_Match
(
    id         SERIAL,
    tg_user_id BIGINT REFERENCES TG_Users (id),
    vk_user_id INT REFERENCES VK_Users (id),
    PRIMARY KEY (tg_user_id, vk_user_id)
);

select first_name, last_name, sex, age, photo_link, profile_link, apartments_budget
FROM vk_match
         INNER JOIN VK_Users ON vk_match.user2_id = VK_Users.id
         INNER JOIN vk_post ON vk_match.user2_id = vk_post.user_id
where user1_id = 384436474;

UPDATE TG_Form
SET match_budget = 7000
WHERE user_id = 681591950;

-- sudo -u postgres psql


SELECT tg.username, tg.first_name, tg.last_name, h.description, h.state, h.date
FROM tg_history as h
         INNER JOIN tg_users as tg ON h.user_id = tg.id
ORDER BY h.date DESC