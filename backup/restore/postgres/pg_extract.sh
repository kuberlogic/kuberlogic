#!/bin/bash
# extract all postgres databases from a sql file created by pg_dumpall
# this script outputs one .sql file for each database in the original .sql file
# unless you pass the name of a database in the dump

if [ $# -lt 1 ]
then
    echo "Usage: $0 <postgresql sql dump> [dbname]" >&2
        exit 1
fi

DB_FILE=$1
DB_NAME=$2
FILEPATH=$3

if [ ! -f $DB_FILE -o ! -r $DB_FILE ]
then
        echo "error: $DB_FILE not found or not readable" >&2
        exit 2
fi

# this loops through all instances of "\connect databasename"
# and tells the line number. the $LINE variable will look like this:
# 3504:\connect databasename
egrep -n "\\connect\ $DB_NAME" $DB_FILE | while read LINE
do

    echo "Evaluating $DB_NAME..."

    # get "3504" from contains "3504:\connect databasename"
    STARTING_LINE_NUMBER=$(echo $LINE | cut -d: -f1)

    # the exported sql should not contain the first line that reads
    # "\connect databasename" otherwise you won't be able to rename the database
    # if we start after that line, you could do something like this:
    # psql new_databasename < databasename.sql
    STARTING_LINE_NUMBER=$(($STARTING_LINE_NUMBER+1))

    # use tail to print out all of the file after the STARTING_LINE_NUMBER
    TOTAL_LINES=$(tail -n +$STARTING_LINE_NUMBER $DB_FILE | \
        # search for the line at the end of the sql import for this database
        egrep -n -m 1 "PostgreSQL\ database\ dump\ complete" | \
        # make sure we only act on the first match
        head -n 1 | \
        # and get the line number where we found the match
        cut -d: -f1)
        # we should now know how long the sql import is for this database
        # specifically, we should know how many lines there are

    echo "$DB_NAME begins on line $STARTING_LINE_NUMBER and ends after $TOTAL_LINES lines"

    # use tail to pipe from the starting line number, and piping X amount of lines (where X is TOTAL_LINES)
    # this gets piped into a file named after the database: DB_NAME.sql
    tail -n +$STARTING_LINE_NUMBER $DB_FILE | head -n +$TOTAL_LINES > $FILEPATH

done