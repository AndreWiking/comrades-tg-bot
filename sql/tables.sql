DROP TABLE IF EXISTS VK_Users CASCADE;
CREATE TABLE VK_Users
(
    id             INTEGER PRIMARY KEY,
    first_name     VARCHAR(128),
    last_name      VARCHAR(128),
    sex            INT,
    age            INT,
    is_open_direct BOOL,
    photo_link     VARCHAR(1024),
    profile_link   VARCHAR(1024)
);

DROP TABLE IF EXISTS VK_Post CASCADE;
CREATE TABLE VK_Post
(
    id                    INTEGER PRIMARY KEY,
    user_id               INTEGER REFERENCES VK_Users (id),
    text                  TEXT,
    link                  VARCHAR(1024),
    date                  TIMESTAMP,
    apartments_budget     INTEGER,
    apartments_location_s FLOAT,
    apartments_location_w FLOAT,
    roommate_sex          INTEGER,
    type                  INT default 0
);

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
