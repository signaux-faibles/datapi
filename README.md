# datAPI - simple API for data distribution with security

/!\ datAPI is in its early development stages, use with caution (and I would be please if you report me issues)
/!\ outdated readme

## What does datAPI do ?
The main purpose of datAPI is to provide a simple mechanism to distribute data through an API with security concerns.  

datAPI offers a REST datastore with an embedded authentication mechanism and rules to distribute/recieve data to/from authorized users.  

datAPI offers the possibility to execute queries in the past, publish data in the future and browse evolutions of data in time while continuing to apply all the rules that evolved in this time.

## How does it work ?
datAPI is a Go program that works uppon PostgreSQL RDBMS. Objects are constructed around a key and a scope and can are constructed by accumulation of json data. All data and rules are timestamped making it possible to evaluate all rules at any time.

### Keys and Subkeys
Keys are `map[string]string` variables stored as a postgresql `hstore`.  
A subkey is part of a more complete key and is used in queries to select more than one object in a single query.  

#### Example
Given 3 objects' keys:

`{'type': 'invoice', 'id': 'I201901041'}`  
`{'type': 'invoice', 'id': 'I201902091'}`  
`{'type': 'command', 'id': 'I201902091'}`

`{'type': 'invoice'}` subkey can query the two invoices with one single query.
`{'id': 'I201902091'}` subkey can query `I201902091` command and invoice at one time .

`{}` is the mother of all subkeys and query everything at one time.

### Scope
A scope is a collection of tags. Users and objects have scopes.  
A user can only get access to data if his scope contains every tags of the object scopes.  

#### Example
John Doe has a scope `['ohio', 'california', 'florida']`  
Jane Doe has another scope `['ohio', 'arizona', 'montana']`  

Both John and Jane can get data that have only `['ohio']` scope.  
John can get data that have `['california', 'florida']` scope but Jane can't.  
Neither John or Jane can get data that have `['ohio', 'utah']` scope.

### Buckets and policies
Every objects are associated with a bucket and a key. Buckets are virtual entities that carry security policies. Theses policies are associated to one or more buckets (given a posix regexp), a subkey, and apply to read and write operations.
- read: policy scope is added to object scopes, like a minimal access scope.
- write: additional scope tags are needed in user scope in order to write data.

In order to store administrative objects like users and policies, datAPI uses the auth bucket. A default policy is configured by default to protect auth bucket from external reads, but nothing will prevent you to break it or create a user that has this privilege... :)


## Installation
### Dependencies
At the time, I'm using go 1.10.4 and postgresql 10. No other versions have been tested.

- [postgresql](https://www.postgresql.org/)
- [gin-gonic](https://github.com/gin-gonic/gin)
- [gin-jwt](https://github.com/appleboy/gin-jwt)
- [viper](https://github.com/spf13/viper)
- [bcrypt](https://golang.org/x/crypto/bcrypt)

### Daemon
- Install postgresql and postgresql-commons.
- Create database, roles etc. The role you create needs to be dbowner.
- `go get github.com/signaux-faibles/datapi`
- Create a working directory to store config.toml
- Copy config.toml.example to config.toml from source and edit values
- run `$GOPATH/bin/datapi`

### Client
A minimal UI designed essentially to test auth, get and puts can be found in sources in the client subdirectory.  
To run it, you'll need to `yarn install` and `yarn serve`

