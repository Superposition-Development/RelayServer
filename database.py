import sqlite3
import init

def initialize():

    readSqlFile = open('init.sql', 'r')
    sqlFile = readSqlFile.read()
    readSqlFile.close()

    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()

    connection.execute("PRAGMA foreign_keys = ON;")
    
    cursor.executescript(sqlFile)

    connection.commit()
    connection.close()

"""
get data from a table titled tableName
columnTitles: the column titles to query (they will be returned)
columnName: the column name to query by (best used with a UNIQUE or PRIMARY KEY value)
inputValue: the value to search for inside of the column titled {columnName}
"""
def queryTableValue(columnTitles, tableName, columnName, inputValue, duplicateResults=False):
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

    resultMap = {}
    
    result = cursor.fetchall() if duplicateResults else cursor.fetchone()

    if not result:
        connection.close()
        return None

    if duplicateResults:
        connection.close()
        return result

    resultMap = dict(zip(columnTitles, result))

    connection.close()
    return resultMap


def addRowAndReturnRowID(columnValueMap, tableName):
    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()

    columns = ", ".join(columnValueMap.keys())
    placeholders = ", ".join(["?"] * len(columnValueMap))
    values = tuple(columnValueMap.values())
    # print(values)

    cursor.execute(f""" \
        INSERT INTO {tableName} ({columns}) 
        VALUES ({placeholders});
         """, values)
    
    cursor.execute("SELECT last_insert_rowid()")
    result = cursor.fetchone()
    connection.commit()
    connection.close()
    return result[0]

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