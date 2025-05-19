# atlas-provider-gorm

Load [GORM](https://gorm.io/) schemas into an [Atlas](https://atlasgo.io) project.

### Use-cases
1. **Declarative migrations** - use a Terraform-like `atlas schema apply --env gorm` to apply your GORM schema to the database.
2. **Automatic migration planning** - use `atlas migrate diff --env gorm` to automatically plan a migration from  
  the current database version to the GORM schema.

### Installation

Install Atlas from macOS or Linux by running:
```bash
curl -sSf https://atlasgo.sh | sh
```
See [atlasgo.io](https://atlasgo.io/getting-started#installation) for more installation options.

Install the provider by running:
```bash
go get -u ariga.io/atlas-provider-gorm
``` 

#### Standalone 

If all of your GORM models and [views](#views) exist in a single package, and the models either embed `gorm.Model` or contain `gorm` struct tags, 
you can use the provider directly to load your GORM schema into Atlas.

In your project directory, create a new file named `atlas.hcl` with the following contents:

```hcl
data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./path/to/models",
    "--dialect", "mysql", // | postgres | sqlite | sqlserver
    "--build-tags", "" // this is optional in case some models are in tagged packages
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = "docker://mysql/8/dev"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
```

##### Pinning Go dependencies

Next, to prevent the Go Modules system from dropping this dependency from our `go.mod` file, let's
follow its [official recommendation](https://go.dev/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)
for tracking dependencies of tools and add a file named `tools.go` with the following contents:

```go title="tools.go"
//go:build tools
package main

import _ "ariga.io/atlas-provider-gorm/gormschema"
```
Alternatively, you can simply add a blank import to the `models.go` file we created
above.

Finally, to tidy things up, run:

```text
go mod tidy
```

#### As Go File

If you want to use the provider as a Go file, you can use the provider as follows:

Create a new program named `loader/main.go` with the following contents:

```go
package main

import (
  "io"
  "os"

  "ariga.io/atlas-provider-gorm/gormschema"
  _ "ariga.io/atlas-go-sdk/recordriver"
  "github.com/<yourorg>/<yourrepo>/path/to/models"
)

func main() {
  stmts, err := gormschema.New("mysql").Load(&models.User{}, &models.Pet{})
  if err != nil {
    fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
    os.Exit(1)
  }
  io.WriteString(os.Stdout, stmts)
}
```

In your project directory, create a new file named `atlas.hcl` with the following contents:

```hcl
data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "./loader",
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = "docker://mysql/8/dev"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
```

### Examples

- [Composite Types](https://atlasgo.io/guides/orms/gorm/composite-types)
- [Domain Types](https://atlasgo.io/guides/orms/gorm/domain-types)
- [Enum Types](https://atlasgo.io/guides/orms/gorm/enum-types)
- [Extensions](https://atlasgo.io/guides/orms/gorm/extensions)
- [Row-Level Security (RLS)](https://atlasgo.io/guides/orms/gorm/row-level-security)
- [Triggers](https://atlasgo.io/guides/orms/gorm/triggers)

### Extra Features

#### Views

> Note: Views are available for logged-in users, run `atlas login` if you haven't already. To learn more about logged-in features for Atlas, visit [Feature Availability](https://atlasgo.io/features#database-features).

To define a Go struct as a database `VIEW`, implement the `ViewDef` method as follow:

```go
// User is a regular gorm.Model stored in the "users" table.
type User struct {
  gorm.Model
  Name string
  Age  int
}

// WorkingAgedUsers is mapped to the VIEW definition below.
type WorkingAgedUsers struct {
  Name string
  Age  int
}

func (WorkingAgedUsers) ViewDef(dialect string) []gormschema.ViewOption {
  return []gormschema.ViewOption{
    gormschema.BuildStmt(func(db *gorm.DB) *gorm.DB {
      return db.Model(&User{}).Where("age BETWEEN 18 AND 65").Select("name, age")
    }),
  }
}
```

In order to pass a plain `CREATE VIEW` statement, use the `CreateStmt` as follows:

```go
type BotlTracker struct {
  ID   uint
  Name string
}

func (BotlTracker) ViewDef(dialect string) []gormschema.ViewOption {
  var stmt string
  switch dialect {
  case "mysql":
    stmt = "CREATE VIEW botl_trackers AS SELECT id, name FROM pets WHERE name LIKE 'botl%'"
  }
  return []gormschema.ViewOption{
    gormschema.CreateStmt(stmt),
  }
}
```

To include both VIEWs and TABLEs in the migration generation, pass all models to the `Load` function:

```go
stmts, err := gormschema.New("mysql").Load(
  &models.User{},			// Table-based model.
  &models.WorkingAgedUsers{},	// View-based model.
)
```

The view-based model works just like a regular models in GORM queries. However, make sure the view name is identical to the struct name, and in case they are differ, configure the name using the `TableName` method:

```go
func (WorkingAgedUsers) TableName() string {
  return "working_aged_users_custom_name" // View name is different than pluralized struct name.
}
```

#### Trigger

> Note: Trigger feature is only available for logged-in users, run `atlas login` if you haven't already. To learn more about logged-in features for Atlas, visit [Feature Availability](https://atlasgo.io/features#database-features).

To attach triggers to a table, use the `Triggers` method as follows:

```go
type Pet struct {
  gorm.Model
  Name string
}

func (Pet) Triggers(dialect string) []gormschema.Trigger {
  var stmt string
  switch dialect {
  case "mysql":
    stmt = "CREATE TRIGGER pet_insert BEFORE INSERT ON pets FOR EACH ROW SET NEW.name = UPPER(NEW.name)"
  }
  return []gormschema.Trigger{
    gormschema.NewTrigger(gormschema.CreateStmt(stmt)),
  }
}
```

### Additional Configuration

To supply custom `gorm.Config{}` object to the provider use the [Go Program Mode](#as-go-file) with
the `WithConfig` option. For example, to disable foreign keys:

```go
loader := New("sqlite", WithConfig(
    &gorm.Config{
        DisableForeignKeyConstraintWhenMigrating: true,
    },
))
```

For a full list of options, see the [GORM documentation](https://gorm.io/docs/gorm_config.html).

### Usage

Once you have the provider installed, you can use it to apply your GORM schema to the database:

#### Apply

You can use the `atlas schema apply` command to plan and apply a migration of your database to
your current GORM schema. This works by inspecting the target database and comparing it to the
GORM schema and creating a migration plan. Atlas will prompt you to confirm the migration plan
before applying it to the database.

```bash
atlas schema apply --env gorm -u "mysql://root:password@localhost:3306/mydb"
```
Where the `-u` flag accepts the [URL](https://atlasgo.io/concepts/url) to the
target database.

#### Diff

Atlas supports a [versioned migration](https://atlasgo.io/concepts/declarative-vs-versioned#versioned-migrations) 
workflow, where each change to the database is versioned and recorded in a migration file. You can use the
`atlas migrate diff` command to automatically generate a migration file that will migrate the database
from its latest revision to the current GORM schema.

```bash
atlas migrate diff --env gorm 
```

### Supported Databases

The provider supports the following databases:
* MySQL
* PostgreSQL
* SQLite
* SQL Server

### Frequently Asked Questions

* **Foreign key constraints not generated correctly** - When setting up your [Go Program](#as-go-file) and using [Customize JoinTable](https://gorm.io/docs/many_to_many.html#Customize-JoinTable), 
  you may encounter issues with foreign key constraints. To avoid these issues, ensure that all models, including the join tables, are passed to the `Load` function.

  For example if those are your models:
  ```go
  type Person struct {
    ID        int
    Name      string
    Addresses []Address `gorm:"many2many:person_addresses;"`
  }
  
  type Address struct {
    ID   int
    Name string
  }
  
  type PersonAddress struct {
    PersonID  int `gorm:"primaryKey"`
    AddressID int `gorm:"primaryKey"`
    CreatedAt time.Time
    DeletedAt gorm.DeletedAt
  }
  ```
  
  you should use the following code:
  ```go
  stmts, err := gormschema.New("mysql").Load(&Models.Address{}, &Models.Person{}, &Models.PersonAddress{})
  ```

* **How to handle enums and custom types?** -
    The recommended way to handle custom types that are not supported by GORM such as postgres enums is to use [composite schemas](https://atlasgo.io/atlas-schema/projects#data-source-composite_schema). 

    First you need to define your custom type inside state file, lets call it `schema.sql`:
    ```sql
    CREATE TYPE "status" AS ENUM ('active', 'inactive', 'deleted');
    ```
 
  Next, you need to add the custom type to your GORM model using [type field tag](https://gorm.io/docs/models.html#Fields-Tags): 
  ```diff filename ="models/player.go"
  package models
  
  import (
  	"gorm.io/gorm"
  )
  
  type Player struct {
  	gorm.Model
  	ID      int `gorm:"primaryKey"`
  +	Status  string `gorm:"type:status"`
  }
  
  ```
  Next, you need to use schema composed of your GORM schema and `schema.sql` file, you can do it by using the next `atlas.hcl` config file:
  
  ```hcl
  data "external_schema" "gorm" {
    program = [
      "go",
      "run",
      "-mod=mod",
      "ariga.io/atlas-provider-gorm",
      "load",
      "--path", "./models",
      "--dialect", "postgres",
    ]
  }
  
  data "composite_schema" "app" {
    schema "public" {
      url = "file://schema.sql"
    }
    schema "public" {
      url = data.external_schema.gorm.url
    }
  }
  
  env "composed" {
    src = data.composite_schema.app.url
    dev = "docker://postgres/15/dev?search_path=public"
    migration {
      dir = "file://migrations"
    }
    format {
      migrate {
        diff = "{{ sql . \"  \" }}"
      }
    }
  }
  
  ```
  Now, when running: `atlas migrate diff --env composed` new migration file should be created, containing the custom enum type:
  
  ```sql filename ="20240623142238.sql"
  -- Create enum type "status"
  CREATE TYPE "status" AS ENUM ('active', 'inactive', 'deleted');
  -- Modify "players" table
  ALTER TABLE "players" ADD COLUMN "status" "status" NULL;
  ```


### License

This project is licensed under the [Apache License 2.0](LICENSE).
