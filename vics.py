import sqlite3
import argparse

# Optional argument to use a listed database file. otherwise use vics.sqlite
# argparse with usage
# If no vics.sqlite3 then create it, and make the 'all' table.

parser = argparse.ArgumentParser()
parser.add_argument("-f", "--file", dest="sqlitedb", help="Optional. Path to vics database, if you do not want to use vics.sqlite")
args = parser.parse_args()
if args.sqlitedb:
    sqlitedb = args.sqlitedb
else:
    sqlitedb = "vics.sqlite"

con = sqlite3.connect(sqlitedb)
cur = con.cursor()

# TODO: check if table exists before creatign
cur.execute('''CREATE TABLE el_todo (date text, b64image text, sha text, tags text)''')
date = "2021-05-05"
b64image = "abcdefg1234"
sha = "123"
tags = "test baddata notanimage"
cur.execute("insert into el_todo values (?, ?, ?, ?)", (date, b64image, sha, tags))
con.commit()
con.close()
