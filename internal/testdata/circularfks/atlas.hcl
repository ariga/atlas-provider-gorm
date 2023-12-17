variable "dialect" {
  type = string
}

locals {
  dev_url = {
    mysql = "docker://mysql/8/dev"
    postgres = "docker://postgres/15"
    sqlite = "sqlite://file::memory:?cache=shared"
  }[var.dialect]
}

data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", ".",
    "--dialect", var.dialect,
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = local.dev_url
  migration {
    dir = "file://migrations/${var.dialect}"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
