import sqlite3
import init

#set up all tables
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

"""
get data from a table titled tableName
columnTitles: the column titles to query (they will be returned)
columnName: the column name to query by (best used with a UNIQUE or PRIMARY KEY value)
inputValue: the value to search for inside of the column titled {columnName}
"""
def queryTableValue(columnTitles, tableName, columnName, inputValue):
    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()

    if isinstance(columnTitles, (list, tuple)):
        columns = ", ".join(columnTitles)
    else:
        columns = columnTitles  # single column

    cursor.execute(
        f"SELECT {columns} FROM {tableName} WHERE {columnName} = ?",
        (inputValue,)
    )

    result = cursor.fetchone()
    connection.close()
    return result

def addRowToTable(columnValueMap, tableName):
    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()

    columns = ", ".join(columnValueMap.keys())
    placeholders = ", ".join(["?"] * len(columnValueMap))
    values = tuple(columnValueMap.values())

    cursor.execute(f""" \
        INSERT INTO {tableName} ({columns}) 
        VALUES ({placeholders}) """, values)
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