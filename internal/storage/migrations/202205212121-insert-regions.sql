-- noinspection SqlNoDataSourceInspectionForFile


-- +migrate Up

INSERT INTO regions (id, name, description) VALUES (1, 'Центральный', '');
INSERT INTO regions (id, name, description) VALUES (2, 'Северный', '');
INSERT INTO regions (id, name, description) VALUES (3, 'Северо-Восточный', '');
INSERT INTO regions (id, name, description) VALUES (4, 'Восточный', '');
INSERT INTO regions (id, name, description) VALUES (5, 'Юго-Восточный', '');
INSERT INTO regions (id, name, description) VALUES (6, 'Южный', '');
INSERT INTO regions (id, name, description) VALUES (7, 'Юго-Западный', '');
INSERT INTO regions (id, name, description) VALUES (8, 'Западный', '');
INSERT INTO regions (id, name, description) VALUES (9, 'Северо-Западный', '');
INSERT INTO regions (id, name, description) VALUES (10, 'Зеленоградский', '');
INSERT INTO regions (id, name, description) VALUES (11, 'Троицкий', '');
INSERT INTO regions (id, name, description) VALUES (12, 'Новомосковский', '');

-- +migrate Down

TRUNCATE TABLE regions CASCADE;