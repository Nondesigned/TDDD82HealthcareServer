DROP PROCEDURE IF EXISTS create_user;

DROP PROCEDURE IF EXISTS get_markings;

DELIMITER //

CREATE PROCEDURE create_user(IN user_nfc_id BIGINT, IN password VARCHAR(64), IN salt VARCHAR(20), OUT phonenumber INT)
	BEGIN
		DECLARE generated_number INT;

		SELECT FLOOR(RAND() * 2147482647 + 1000) INTO generated_number;

		INSERT INTO user(NFC_id, password, salt, phonenumber) VALUES(user_nfc_id, password, salt, generated_number);

		SET phonenumber = generated_number;
	END //

CREATE PROCEDURE get_markings(IN user_nfc_id BIGINT)
	BEGIN
		SELECT M.group_id, M.type, M.creation_time, M.latitude, M.longitude FROM marking M, user U, groupmember G 
		WHERE G.user_id = U.NFC_id AND G.group_id = M.group_id;
	END //
		
DELIMITER ;