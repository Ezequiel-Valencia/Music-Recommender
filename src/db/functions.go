package db

const hasUserSubmitCountHitLimit = `
CREATE OR REPLACE FUNCTION reached_submit_limit(target_user_id INT, submission_limit INT) 
RETURNS BOOLEAN AS 
$reached_submit_limit$

DECLARE
    result BOOLEAN;
	num_submissions INT;
BEGIN
	SELECT song_sets_submitted INTO num_submissions FROM users WHERE user_id = target_user_id;
    IF num_submissions >= submission_limit THEN
        result := true;
    ELSIF num_submissions < submission_limit THEN
        result := false;
		UPDATE users SET song_sets_submitted = song_sets_submitted + 1 WHERE user_id = target_user_id;
    ELSE
        result := true;
    END IF;
    RETURN result;
END;

$reached_submit_limit$ LANGUAGE plpgsql;
`

