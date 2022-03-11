[![GoDoc](https://godoc.org/github.com/golobby/orm?status.svg)](https://godoc.org/github.com/golobby/orm)

[//]: # ([![CI]&#40;https://github.com/golobby/orm/actions/workflows/ci.yml/badge.svg&#41;]&#40;https://github.com/golobby/orm/actions/workflows/ci.yml&#41;)

[//]: # ([![CodeQL]&#40;https://github.com/golobby/orm/workflows/CodeQL/badge.svg&#41;]&#40;https://github.com/golobby/orm/actions?query=workflow%3ACodeQL&#41;)

[![Go Report Card](https://goreportcard.com/badge/github.com/golobby/orm)](https://goreportcard.com/report/github.com/golobby/orm)
[![Coverage Status](https://coveralls.io/repos/github/golobby/orm/badge.svg)](https://coveralls.io/github/golobby/orm?branch=master)

# GolobbyORM

GoLobbyORM is a simple yet powerfull, fast, safe, customizable, type-safe database toolkit for Golang.

## Table Of Contents
- [GolobbyORM](#golobbyorm)
  * [Table Of Contents](#table-of-contents)
  * [Features](#features)
    + [Introduction](#introduction)
    + [Creating a new Entity](#creating-a-new-entity)
      - [Conventions](#conventions)
        * [Timestamps](#timestamps)
        * [Column names](#column-names)
        * [Primary Key](#primary-key)
    + [Initializing ORM](#initializing-orm)
    + [Fetching an entity from database](#fetching-an-entity-from-database)
    + [Saving entities or Insert/Update](#saving-entities-or-insert-update)
    + [Deleting entities](#deleting-entities)
    + [Relationships](#relationships)
      - [HasMany](#hasmany)
      - [HasOne](#hasone)
      - [BelongsTo](#belongsto)
      - [BelongsToMany](#belongstomany)
      - [Saving with relation](#saving-with-relation)
  * [License](#license)


## Features
- Simple, type safe, elegant API with help of `Generics`
- Minimum reflection usage, mostly at startup of application
- No code generation
- Query builder for various query types
- Binding query results to entities
- Supports relationship/Association types
    - HasMany
    - HasOne
    - BelongsTo
    - BelongsToMany (ManyToMany)

  
### Introduction
GolobbyORM an object-relational mapper (ORM) that makes it enjoyable to interact with your database. 
When using GolobbyORM, each database table has a corresponding "Entity" that is used to interact with that table.
In addition to retrieving records from the database table, GolobbyORM entities allow you to insert,
update, and delete records from the table as well.

### Creating a new Entity
Lets create a new `Entity` to represent `User` in our application.

```go
package main

import "github.com/golobby/orm"

type User struct {
  ID       int64
  Name     string
  LastName string
  Email    string
  orm.Timestamps
}

func (u User) ConfigureEntity(e *orm.EntityConfigurator) {
    e.
		Table("users").
		Connection("default") // You can omit connection name if you only have one.
	
}
```
as you see our user entity is nothing else than a simple struct and two methods.
Entities in GolobbyORM are implementations of `Entity` interface which defines two methods:
- ConfigureEntity: configures table and also relations to other entities.
#### Conventions
##### Timestamps
for having `created_at`, `updated_at`, `deleted_at` timestamps in your entities you can embed `orm.Timestamps` struct in your entity,
```go
type User struct {
  ID       int64
  Name     string
  LastName string
  Email    string
  orm.Timestamps
}

```
also if you want custom names for them you can do it like this.
```go
type User struct {
    ID       int64
    Name     string
    LastName string
    Email    string
    MyCreatedAt sql.NullTime `orm:"created_at=true"`
    MyUpdatedAt sql.NullTime `orm:"updated_at=true"`
    MyDeletedAt sql.NullTime `orm:"deleted_at=true"`
}
```
You can use `orm` struct tag and there is a key for each timestamp which you can use on any time compatible struct field.
##### Column names
GolobbyORM for each struct field(except slice, arrays, maps and other nested structs) assumes a respective column named using snake case syntax.
if you want to have a custom column name you should specify it in entity struct.
```go
package main
type User struct {
	Name string `orm:"column=username"` // now this field will be mapped to `username` column in sql database. 
}
```
##### Primary Key
GolobbyORM assumes that each entity has primary key named `id`, if you want to have a custom named primary key you need to specify it in entity struct.
```go
package main
type User struct {
	PK int64 `orm:"pk=true"`
}
```

### Initializing ORM
after creating our entities we need to initialize GolobbyORM.
```go
package main

import "github.com/golobby/orm"

func main() {
  orm.Initialize(orm.ConnectionConfig{
    //Name:             "default", You should specify connection name if you have multiple connections
    Driver:           "sqlite3",
    ConnectionString: ":memory:",
  })
}
```
After this step we can start using ORM.
### Fetching an entity from database
GolobbyORM makes it trivial to fetch entity from database using its primary key.
```go
user, err := orm.Find[User](1)
```
`orm.Find` is a generic function that takes a generic parameter that specifies the type of `Entity` we want to query and it's primary key value.
You can also use custom queries to get entities from database.
```go

user, err := orm.Query[User]().Where("id", 1).One()
user, err := orm.Query[User]().WherePK(1).One()
```
GolobbyORM contains a powerful query builder which you can use to build `Select`, `Update` and `Delete` queries, but if you want to write a raw sql query you can.
```go
users, err := orm.QueryRaw[User](`SELECT * FROM users`)
```

### Saving entities or Insert/Update
GolobbyORM makes it easy to persist an `Entity` to the database using `Save` method, it's an UPSERT method, if the primary key field is not zero inside entity
it will go for update query, otherwise it goes for insert.
```go
// this will insert entity into the table
err := orm.Save(&User{Name: "Amirreza"}) // INSERT INTO users (name) VALUES (?) , "Amirreza"
```
```go
//this will update entity with id = 1
orm.Save(&User{ID: 1, Name: "Amirreza2"}) // UPDATE users SET name=? WHERE id=?, "Amirreza2", 1
```
also you can do custom update queries using again query builder or raw sql as well.
```go
res, err := orm.Query[User]().Where("id", 1).Update(orm.KV{"name": "amirreza2"})
```

using raw sql

```go
_, affected, err := orm.ExecRaw[User](`UPDATE users SET name=? WHERE id=?`, "amirreza", 1)
```
### Deleting entities
It is also easy to delete entities from database.
```go
err := orm.Delete(user)
```
you can also use query builder or raw sql.
```go
_, affected, err := orm.Query[Post]().WherePK(1).Delete()

_, affected, err := orm.Query[Post]().Where("id", 1).Delete()

```
```go
_, affected, err := orm.ExecRaw[Post](`DELETE FROM posts WHERE id=?`, 1)
```
### Relationships
GolobbyORM makes it easy to have entities that have relationships with each other. Configuring relations is using `ConfigureEntity` method as you will see.
#### HasMany
```go
type Post struct {}

func (p Post) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Table("posts").HasMany(&Comment{}, orm.HasManyConfig{})
}
```
As you can see we are defining a `Post` entity which has a `HasMany` relation with `Comment`. You can configure how GolobbyORM queries `HasMany` relation with `orm.HasManyConfig` object, by default it will infer all fields for you.
now you can use this relationship anywhere in your code.
```go
comments, err := orm.HasMany[Comment](post).All()
```
`HasMany` and other relation functions in GolobbyORM return `QueryBuilder` and you can use them like other query builders and create even more
complex queries for relationships. for example you can create a query for getting all comments of a post that are created today.
```go
todayComments, err := orm.HasMany[Comment](post).Where("created_at", "CURDATE()").All()
```
#### HasOne
Configuring a `HasOne` relation is like `HasMany`.
```go
type Post struct {}

func (p Post) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Table("posts").HasOne(&HeaderPicture{}, orm.HasOneConfig{})
}
```
As you can see we are defining a `Post` entity which has a `HasOne` relation with `HeaderPicture`. You can configure how GolobbyORM queries `HasOne` relation with `orm.HasOneConfig` object, by default it will infer all fields for you.
now you can use this relationship anywhere in your code.
```go
picture, err := orm.HasOne[HeaderPicture](post)
```
`HasOne` also returns a query builder and you can create more complex queries for relations.
#### BelongsTo
```go
type Comment struct {}

func (c Comment) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Table("comments").BelongsTo(&Post{}, orm.BelongsToConfig{})
}
```
As you can see we are defining a `Comment` entity which has a `BelongsTo` relation with `Post` that we saw earlier. You can configure how GolobbyORM queries `BelongsTo` relation with `orm.BelongsToConfig` object, by default it will infer all fields for you.
now you can use this relationship anywhere in your code.
```go
post, err := orm.BelongsTo[Post](comment)
```
#### BelongsToMany
```go
type Post struct {}

func (p Post) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Table("posts").BelongsToMany(&Category{}, orm.BelongsToManyConfig{IntermediateTable: "post_categories"})
}

type Category struct{}
func(c Category) ConfigureEntity(r *orm.EntityConfigurator) {
    e.Table("categories").BelongsToMany(&Post{}, orm.BelongsToManyConfig{IntermediateTable: "post_categories"})
}

```
we are defining a `Post` entity and also a `Category` entity which have a many2many relationship, as you can see it's mandatory for us to configure IntermediateTable name which GolobbyORM cannot infer by itself now.
now you can use this relationship anywhere in your code.
```go
categories, err := orm.BelongsToMany[Category](post)
```
#### Saving with relation
You may need to save an entity which has some kind of relation with another entity, in that case you can use `Add` method.
```go
orm.Add(post, comments...) // inserts all comments passed in and also sets all post_id to the primary key of the given post.
```
## License

GoLobby ORM is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
