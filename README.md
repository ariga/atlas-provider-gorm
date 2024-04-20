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

If all of your GORM models exist in a single package, and either embed `gorm.Model` or contain `gorm` struct tags, 
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

* **Foreign key constraints not generated correctly** -
  If a [Customize JoinTable](https://gorm.io/docs/many_to_many.html#Customize-JoinTable) is defined in the schema, 
  you need to use the provider as a [Go Program](#as-go-file) and set it up using the `WithJoinTable` option.
  
  for example if those are your models:
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
  stmts, err := gormschema.New("mysql",
  	gormschema.WithJoinTable(
  		&Models.Person{}, "Addresses", &Models.PersonAddress{},
  	),
  ).Load(&Models.Address{}, &Models.Person{})
  ```

### License

This project is licensed under the [Apache License 2.0](LICENSE).
