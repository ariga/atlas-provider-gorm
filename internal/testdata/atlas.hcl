variable "dialect" {
  type = string
}

locals {
  dev_url = {
    mysql = "docker://mysql/8/dev"
    postgres = "docker://postgres/15"
    sqlserver = "docker://sqlserver/2022-latest"
    sqlite = "sqlite://file::memory:?cache=shared"
    spanner = "docker://spanner/latest"
  }[var.dialect]
}

data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./models",
    "--dialect", var.dialect,
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = local.dev_url
  migration {
    dir = "file://migrations/${var.dialect}"
  }
  diff {
    skip {
      rename_constraint = true
    }
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
