# Motivation

Read from multiple slaves and write to single master while working with sql database cluster.
Inspired by http://github.com/tsenart/nap

## Install
```shell
$ go get github.com/datalinkE/sqlxentrypoint
```

## Usage
```go
package main

import (
  "log"

  "github.com/datalinkE/sqlx-entrypoint"
  _ "github.com/go-sql-driver/mysql" // Any sql.DB works
)

func main() {
  // The first DSN is assumed to be the master and all
  // other to be slaves
  dsns := "tcp://user:password@master/dbname;"
  dsns += "tcp://user:password@slave01/dbname;"
  dsns += "tcp://user:password@slave02/dbname"

  db, err := sqlx-entrypoint.Open("mysql", dsns)
  if err != nil {
    log.Fatal(err)
  }

  if err := db.Ping(); err != nil {
    log.Fatalf("Some physical database is unreachable: %s", err)
  }

  // Read queries are directed to slaves with Query and QueryRow.
  // Load distribution is round-robin.
  var count int
  err = db.Slave().QueryRow("SELECT COUNT(*) FROM sometable").Scan(&count)
  if err != nil {
    log.Fatal(err)
  }

  // Write queries are directed to the master with Exec.
  err = db.Master().Exec("UPDATE sometable SET something = 1")
  if err != nil {
    log.Fatal(err)
  }

  // Transactions always use the master
  tx, err := db.Master().Begin()
  if err != nil {
    log.Fatal(err)
  }
  // Do something transactional ...
  if err = tx.Commit(); err != nil {
    log.Fatal(err)
  }

}
```

## Todo
* Support other slave load balancing algorithms.

## License
```
The MIT License (MIT)

Copyright (c) 2013 Tomás Senart http://github.com/tsenart/nap

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```

