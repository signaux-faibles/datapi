# datAPI - simple API for data distribution with security

The main purpose of datAPI is to provide a simple mechanism to distribute datas through an API with security concerns.

## What does datAPI do ?
datAPI offers a datastore 

## How does it work ?
datAPI is a go program that works uppon PostgreSQL DBMS. Minimal configuration is needed server.  
Objects are stored with a key and a security tags and can be constructed by accumulation of data.  
A same key can be viewed as total different data by different users.

### Scope
A scope is a collection of tags. Users and objects have scopes.  
A user can only get access to data if his scope contains every tags of the object scopes.  

### Buckets and policies


### Key
Keys are 

## Installation
- Install postgresql and postgresql-commons.
- Create database, roles etc. The role you create needs to be dbowner.
- Configure config.toml
`go get -u github.com/signaux-faibles/datapi`

## Usage
### datAPI daemon
not written yet
### datAPI CLI
not written yet

