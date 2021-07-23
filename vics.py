import sqlite3
import argparse
import logging

# Optional argument to use a listed database file. otherwise use vics.sqlite
# argparse with usage
# If no vics.sqlite3 then create it, and make the 'all' table.

parser = argparse.ArgumentParser()
parser.add_argument("-f", "--file", dest="db_file", help="Optional. Path to vics database, if you do not want to use vics.sqlite")
args = parser.parse_args()
if args.db_file:
    db_file = args.db_file
else:
    db_file = "vics.sqlite"
logging.info(f"database file: {db_file}")

table_creation_string = '''CREATE TABLE el_todo (date text, b64image text, sha text, tags text)'''

def create_new_database(sqlitedb_filename):
    con = sqlite3.connect(sqlitedb_filename)
    cur = con.cursor()
    cur.close()



def sqlite_table_schema(conn, name):
    """Return a string representing the table's CREATE. via https://techoverflow.net/2019/10/14/how-to-get-schema-of-sqlite3-table-in-python/"""
    con = sqlite3.connect(sqlitedb)
    cur = con.cursor()
    cursor = conn.execute("SELECT sql FROM sqlite_master WHERE name=?;", [name])
    sql = cursor.fetchone()[0]
    cursor.close()
    return sql

def old_stuff_from_first_session():
    try:
        el_todo_schema = sqlite_table_schema(con, 'el_todo')
        if table_creation_string != el_todo_schema:
            schema_mismatch_error = f"schema mismatch. \n\nExpected: {table_creation_string}\nFound: {el_todo_schema}\n"
            logging.critical(schema_mismatch_error)
            exit(schema_mismatch_error)
    except TypeError:
        logging.info("Table 'el_todo' not found, creating.")
        cur.execute(table_creation_string)
        con.commit()
    date = "2021-05-05"
    b64image = "abcdefg1234"
    sha = "123"
    tags = "test baddata notanimage"
    cur.execute("insert into el_todo values (?, ?, ?, ?)", (date, b64image, sha, tags))
    con.commit()
    con.close()
