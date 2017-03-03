DROP PROCEDURE IF EXISTS create_user;

DROP PROCEDURE IF EXISTS get_user_password;

DROP PROCEDURE IF EXISTS create_group;

DROP PROCEDURE IF EXISTS add_user_to_group;

DROP PROCEDURE IF EXISTS set_user_token;

DROP PROCEDURE IF EXISTS get_user_token;

DROP PROCEDURE IF EXISTS get_markings;


DELIMITER //

CREATE PROCEDURE create_user(IN name VARCHAR(100), IN user_nfc_id BIGINT, IN password VARCHAR(64), IN salt VARCHAR(20), OUT phonenumber INT)
	BEGIN
		DECLARE generated_number INT;

		SELECT FLOOR(RAND() * 2147482647 + 1000) INTO generated_number;

		INSERT INTO user(name, NFC_id, password, salt, phonenumber) VALUES(name, user_nfc_id, password, salt, generated_number);

		SET phonenumber = generated_number;
	END //

CREATE PROCEDURE get_user_password(IN user_nfc_id BIGINT)
	BEGIN
		SELECT password, salt FROM user WHERE NFC_id = user_nfc_id LIMIT 1;
	END //


CREATE PROCEDURE create_group(IN name VARCHAR(30), OUT group_id INT)
	BEGIN
		INSERT INTO usergroup(name) VALUES(name);
	END //

CREATE PROCEDURE add_user_to_group(IN user_nfc_id BIGINT, IN group_id INT)
	BEGIN
		INSERT INTO groupmember(user_id, group_id) VALUES(user_nfc_id, group_id);
	END //

CREATE PROCEDURE set_user_token(IN user_id BIGINT, IN valid_to DATETIME, IN data VARCHAR(200))
	BEGIN
		INSERT INTO token(owner_id, valid_to, data) VALUES(user_id, valid_to, data);
	END //

CREATE PROCEDURE get_user_token(IN user_id BIGINT)
	BEGIN
		SELECT data FROM token WHERE owner_id = user_id LIMIT 1;
	END //

CREATE PROCEDURE get_markings(IN user_nfc_id BIGINT)
	BEGIN
		SELECT M.group_id, M.type, M.creation_time, M.latitude, M.longitude FROM marking M, user U, groupmember G 
		WHERE G.user_id = U.NFC_id AND G.group_id = M.group_id;
	END //


		
DELIMITER ;