
DROP TABLE IF EXISTS marking;

DROP TABLE IF EXISTS groupmember;

DROP TABLE IF EXISTS token;

DROP TABLE IF EXISTS user;

DROP TABLE IF EXISTS usergroup;

CREATE TABLE user(
	phonenumber INT,
	NFC_id BIGINT,
	name VARCHAR(100),
	password VARCHAR(64),
	salt VARCHAR(20),

	CONSTRAINT pk_user
		PRIMARY KEY(phonenumber)
)ENGINE = InnoDB;

CREATE TABLE usergroup(
	id INT AUTO_INCREMENT,
	name VARCHAR(30),

	CONSTRAINT pk_group
		PRIMARY KEY(id)
)ENGINE = InnoDB;

CREATE TABLE token(
	owner_number INT,
	data VARCHAR(200),

	CONSTRAINT pk_token
		PRIMARY KEY(owner_number),

	CONSTRAINT fk_token_user
		FOREIGN KEY(owner_number) REFERENCES user(phonenumber) ON DELETE CASCADE
)ENGINE = InnoDB;

CREATE TABLE groupmember(
	user_number INT,
	group_id INT,

	CONSTRAINT pk_groupmember
		PRIMARY KEY(user_number, group_id),

	CONSTRAINT fk_groupmember_user
		FOREIGN KEY(user_number) REFERENCES user(phonenumber) ON DELETE CASCADE,

	CONSTRAINT fk_groupmember_usergroup
		FOREIGN KEY(group_id) REFERENCES usergroup(id) ON DELETE CASCADE
)ENGINE = InnoDB;

CREATE TABLE marking(
	id INT AUTO_INCREMENT,
	group_id INT,
	type ENUM('wounded_guy', 'wounded_cat'),
	creation_time DATETIME,
	latitude DECIMAL(9,6),
	longitude DECIMAL(9,6),

	CONSTRAINT pk_marking
		PRIMARY KEY(id),

	CONSTRAINT fk_marking_usergroup
		FOREIGN KEY(group_id) REFERENCES usergroup(id)
)ENGINE = InnoDB;