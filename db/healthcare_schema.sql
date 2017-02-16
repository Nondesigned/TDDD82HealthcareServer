
DROP TABLE IF EXISTS marking;

DROP TABLE IF EXISTS groupmember;

DROP TABLE IF EXISTS token;

DROP TABLE IF EXISTS user;

DROP TABLE IF EXISTS usergroup;

CREATE TABLE user(
	NFC_id BIGINT,
	name VARCHAR(100),
	password VARCHAR(64),
	salt VARCHAR(20),
	phonenumber INT,

	CONSTRAINT pk_user
		PRIMARY KEY(NFC_id)
)ENGINE = InnoDB;

CREATE TABLE usergroup(
	id INT AUTO_INCREMENT,
	name VARCHAR(30),

	CONSTRAINT pk_group
		PRIMARY KEY(id)
)ENGINE = InnoDB;

CREATE TABLE token(
	owner_id BIGINT,
	valid_to DATETIME,
	data VARCHAR(100),

	CONSTRAINT pk_token
		PRIMARY KEY(owner_id),

	CONSTRAINT fk_token_user
		FOREIGN KEY(owner_id) REFERENCES user(NFC_id) ON DELETE CASCADE
)ENGINE = InnoDB;

CREATE TABLE groupmember(
	user_id BIGINT,
	group_id INT,

	CONSTRAINT pk_groupmember
		PRIMARY KEY(user_id, group_id),

	CONSTRAINT fk_groupmember_user
		FOREIGN KEY(user_id) REFERENCES user(NFC_id) ON DELETE CASCADE,

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