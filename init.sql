CREATE TABLE IF NOT EXISTS user (
        id INTEGER PRIMARY KEY,
        pfp TEXT,
        username TEXT,
        userID TEXT UNIQUE,
        password TEXT,
        timestamp INTEGER
    );

CREATE TABLE IF NOT EXISTS message (
        id INTEGER PRIMARY KEY,
        userID TEXT,
        channelID INTEGER,
        timestamp INTEGER,
        content TEXT,
        FOREIGN KEY (userID) REFERENCES user(userID),
        FOREIGN KEY (channelID) REFERENCES channel(id)
        );

CREATE TABLE IF NOT EXISTS channel (
        id INTEGER PRIMARY KEY,
        name TEXT,
        type TEXT NOT NULL,
        serverID INTEGER,
        FOREIGN KEY (serverID) REFERENCES server(id)
    );
  
CREATE TABLE IF NOT EXISTS server (
        id INTEGER PRIMARY KEY,
        pfp TEXT,
        name TEXT,
        timestamp INTEGER
    );

CREATE TABLE IF NOT EXISTS serverUser (
    id INTEGER PRIMARY KEY,
    serverID INTEGER,
    userID TEXT,
    timestamp INTEGER,
    FOREIGN KEY (serverID) REFERENCES server(id),
    FOREIGN KEY (userID) REFERENCES user(userID)
    );

CREATE TABLE IF NOT EXISTS dm (
    id INTEGER PRIMARY KEY,
    textID INTEGER,
    voiceID INTEGER,
    userA TEXT,
    userB TEXT,
    FOREIGN KEY (textID) REFERENCES channel(textID),
    FOREIGN KEY (voiceID) REFERENCES channel(voiceID),
    FOREIGN KEY (userA) REFERENCES user(userA),
    FOREIGN KEY (userB) REFERENCES user(userB)
);
