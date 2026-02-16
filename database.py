import sqlite3
import init
import datetime

def initialize():
    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()

    cursor.execute("""
    CREATE TABLE IF NOT EXISTS user (
        id INTEGER PRIMARY KEY,
        pfp TEXT,
        username TEXT,
        userID TEXT UNIQUE,
        password TEXT,
        joindate DATE
    )
    """)

    connection.commit()
    connection.close()

def queryTableValue(returnColumns, tableName, columnName,inputValue):
    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()
    columns = ", ".join(returnColumns)
    cursor.execute(
     f"SELECT {columns} FROM {tableName} WHERE {columnName} = ?",
    (inputValue,)
    )
    result = cursor.fetchone()
    connection.commit()
    connection.close()
    return result

def createUser(username,password,userID,pfp):
    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()
    cursor.execute(""" \
        INSERT INTO user (username,password,userID,pfp,joindate) 
        VALUES (?, ?, ?, ?, ?) """, (username,password,userID,pfp,datetime.date.today().strftime("%Y-%m-%d")))
    connection.commit()
    connection.close()

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