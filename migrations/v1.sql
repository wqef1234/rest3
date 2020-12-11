
use restdb;

CREATE TABLE IF NOT EXISTS books
(
    uid INT(10) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    Title VARCHAR(250),
    Author VARCHAR(200),
    Rating INT
);
COMMIT;


CREATE TABLE IF NOT EXISTS journals
(
    uid INT(10) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    Title VARCHAR(250),
    Editor  VARCHAR(100),
    PageAmount  INT
);
COMMIT;

INSERT INTO books (Title,Author,Rating) VALUES ('BOOK 1','Author 1',255);
INSERT INTO books (Title,Author,Rating) VALUES ('some book 2','new Author 2',155);
INSERT INTO books (Title,Author,Rating) VALUES ('another BOOK 3','illegal Author 3',455);

INSERT INTO journals (Title,Editor,PageAmount) VALUES ('JOURNAL 1','Editor 1',87);
INSERT INTO journals (Title,Editor,PageAmount) VALUES ('beauty JOURNAL 2','new Editor 2',817);
INSERT INTO journals (Title,Editor,PageAmount) VALUES ('most value JOURNAL 3','Editor 3',313);





