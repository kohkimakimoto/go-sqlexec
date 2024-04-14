-- comment
CREATE TABLE IF NOT EXISTS post (
  id int NOT NULL,
  title text,
  body text,
  PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS comment (
  id int NOT NULL,
  body text,
  PRIMARY KEY(id)
);
