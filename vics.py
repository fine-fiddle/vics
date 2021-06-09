import sqlite3
import argparse


# Optional argument to use a listed database file. otherwise use vics.sqlite
# argparse with usage
# If no vics.sqlite3 then create it, and make the 'all' table.

parser = argparse.ArgumentParser()
parser.add_argument("-f", "--file", dest="sqlitedb", help="Optional. Path to an existing database file, if that file is not the default vics.sqlite")
args = parser.parse_args()
print(args.sqlitedb)