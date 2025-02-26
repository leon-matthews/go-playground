
# Ranking data with SQL window functions

https://antonz.org/sql-window-functions-ranking/

## Setup

Create sample data set, given files *schema.sql* and *data.csv*:

    $ sqlite3 data.db < schema.sql
    sqlite> .separator ,
    sqlite> .import data.csv Employees
