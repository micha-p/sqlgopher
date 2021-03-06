# sqlgopher

A small web-based tool for database administration.

- simple user-interface (links are below values!)
- free of any language, just using well-known symbols, handlers and links
- stepping through tables with primary key
- direct access via query with credentials (use wisely)
- credentials stored in secure cookies
- fast dumping of database content
- inserting, querying and updating data
- templates for html
- changing database driver in future releases (TODO)

### Installation

    export GOPATH=$PWD
    go get github.com/go-sql-driver/mysql
    go get github.com/gorilla/securecookie
    go get -u github.com/micha-p/sqlgopher

### Usage

	cd $GOPATH/src/sqlgopher; $GOPATH/bin/sqlgopher -d -c="html/table.css"

	-c  supply customized style in CSS file
	-d  DEBUG: dynamically load html templates and css
	-h  server name
	-i  include INFORMATION_SCHEMA in overview
	-p  server port
	-r  READONLY access
	-s  https Connection TLS
	-m  modify database schema: create, alter, drop tables (Release 1.1)
	-x  expert mode to access privileges, routines, triggers, views (Release 1.2)


   [http://localhost:8080](http://localhost:8080)

   [http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306](http://localhost:8080/user=galagopher&pass=mypassword&host=localhost&port=3306)

    w3m 'http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306'
    lynx -accept_all_cookies 'http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306'
    curl -s 'http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306&db=galadb&t=posts' | html2text

### User interface


         ___________ home: lift restrictions from right lo left
        /   ___________ column: ascending and descending order
       /   /   ___________ indicator for primary key
      /   /   /
     #	 c1(ID)   c2	c3
     1	   -       -     -
     2	   - \     - \____ value: select group (browse by group)
      \    -  \______ key value: show record (browse by key values)
       \___________ row number: show record (browse by number or range)


    Symbol  |	Action
    --------|-------------
    +       |	Add
    ~       |	Change
    -       |	Remove
    i       |	Info
    <       |   Left (towards start)
    >       |   Right (towards end)


### Security

- access based on mysql grants
- no encrypted connection to mysql server
- use only in trusted environments
- passwords might be supplied or bookmarked via URL
- TLS-encryption possible


##### Javascript

The provided html-templates do not require javascript. However, when toggling null values or entering nonempty strings, two small scripts are used.

##### SQL-injection via Request values

To prevent SQL-injection, all supplied identifiers are backqoted and to prevent escaping, all backquotes are escaped by doubling them.
All values are doublequoted, supplied double quotes are escaped the same way.
Where-clauses are especially difficult to check, as this would require full parsing of SQL-expressions.
Therefore they are avoided, and identifiers and values are transmitted in separate query fields and quoted after importing.
Numbers in limits are not quoted and therefore filtered by a strict regular expression.

##### SQL-injection via Names of input fields

Query keys as taken from input forms might be altered as well, but as these values are looked up taking column names, they are just ignored.


##### Javascript-Injection via Identifiers and Values

If identifiers for tables or fields contain quotes or doublequotes, control might escape from these strings.
Therefore they are protected by escaping html in templates and manually.


##### Login-attack via credentials

Establishing connections to databases is done by the standard library-functions.
Credentials taken from a simple html-form are directly submitted to the library without any further processing.



# Limitations

- insert and query limited by request length
- some data types cause problems at driver level

# License

MIT License
