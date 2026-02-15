import sqlite3
import init

def initialize():
    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()

    cursor.execute("""
    CREATE TABLE IF NOT EXISTS user (
        id INTEGER PRIMARY KEY,
        pfp TEXT,
        username TEXT,
        userID TEXT UNIQUE,
        password TEXT
    )
    """)

    connection.commit()
    connection.close()

def testUser():
    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()
    cursor.execute(""" \
        INSERT INTO user (pfp, username, password, userID) 
        VALUES (?, ?, ?, ?) """, ("skibi", "skib", "skib","skibidi"))
    connection.commit()
    connection.close()

def queryUser(user):
    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()
    cursor.execute(
     "SELECT id, pfp, username, password FROM user WHERE userID = ?",
    (user,)
    )
    a = cursor.fetchone()
    connection.commit()
    connection.close()
    return a

# This will wipe everything, only use when "the trucks are here"
def TheTrucksAreHere():
    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()
    cursor.executescript("""
            PRAGMA writable_schema = 1;
            DELETE FROM sqlite_master;
            PRAGMA writable_schema = 0;
            VACUUM;
            PRAGMA integrity_check;
                         """)
    connection.commit()
    connection.close()