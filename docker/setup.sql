-- Password is password123
INSERT INTO users(username, email, password_hash, creation_source, 
		creation_date, user_role, user_privileges) 
VALUES('Ezequiel', 'fake@gmail.com','$2a$14$cHnCJY0.vz.N6M0XtKKrkOcNuBc0uVtTNR6aofL6e.ewaeWCmRr/2', 'Local', '2024-12-29 18:32:17.000', 'Unlimited', 'Owner');

-- Songs
INSERT INTO music(insert_date, name, artist, songURL, submitter_id) 
VALUES ('2024-12-29 18:32:17.000', 'Moon Child', 'Cibo Matto', 'https://youtu.be/3y3C6jAmC2w', 1);
INSERT INTO music(insert_date, name, artist, songURL, submitter_id) 
VALUES ('2024-12-29 18:32:17.000', 'Black Hole Sun', 'Sound Garden', 'https://youtu.be/3mbBbFH9fAg', 1);
INSERT INTO music(insert_date, name, artist, songURL, submitter_id) 
VALUES ('2024-12-29 18:32:17.000', 'Rebel Yell', 'Billy Idol', 'https://youtu.be/seHlHzYpWBU', 1);

-- Descriptions
INSERT INTO submissionDescriptions(description)
VALUES ('Ranking description');

-- Todays Ranking
INSERT INTO todaysRanking(song_id, curator_id, description_id, song_name, song_artist, song_path_resource, song_order)
VALUES (1, 1, 1, 'Moon Child', 'Cibo Matto', '3y3C6jAmC2w', 0);
INSERT INTO todaysRanking(song_id, curator_id, description_id, song_name, song_artist, song_path_resource, song_order)
VALUES (2, 1, 1, 'Black Hole Sun', 'Sound Garden', '3mbBbFH9fAg', 1);
INSERT INTO todaysRanking(song_id, curator_id, description_id, song_name, song_artist, song_path_resource, song_order)
VALUES (3, 1, 1, 'Rebel Yell', 'Billy Idol', 'seHlHzYpWBU', 2);
