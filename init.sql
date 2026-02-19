CREATE TABLE IF NOT EXISTS user (
        id INTEGER PRIMARY KEY,
        pfp TEXT,
        username TEXT,
        userID TEXT UNIQUE,
        password TEXT,
        joindate DATE
    );

CREATE TABLE IF NOT EXISTS message (
        id INTEGER PRIMARY KEY,
        senderID TEXT,
        channelID INTEGER,
        date DATETIME,
        content TEXT,
        FOREIGN KEY (senderID) REFERENCES user(userID),
        FOREIGN KEY (channelID) REFERENCES channel(id)
        );

CREATE TABLE IF NOT EXISTS channel (
        id INTEGER PRIMARY KEY,
        name TEXT,
        password TEXT,
        joindate DATE,
        serverID INTEGER,
        FOREIGN KEY (serverID) REFERENCES server(id)
    );
  
CREATE TABLE IF NOT EXISTS server (
        id INTEGER PRIMARY KEY,
        pfp TEXT,
        name TEXT,
        password TEXT,
        joindate DATE
    );

CREATE TABLE IF NOT EXISTS serverUser (
    id INTEGER PRIMARY KEY,
    serverID INTEGER,
    userID TEXT,
    joindate DATE,
    FOREIGN KEY (serverID) REFERENCES server(id),
    FOREIGN KEY (userID) REFERENCES user(userID)
    );