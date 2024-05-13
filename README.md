# tfsetup

`tfsetup` uses the go templating language to generate a setup.tf file for terraform projects. Its intended use to help with any policy-based or repetitive setup.

It requires the following:
* `tfsetup-context.json` in the current directory
* `tfsetup-config/` as a directory somewhere at or above the current directory
* `tfsetup-config/context.json`
* `tfsetup-config/setup.tmpl`

`tfsetup` Can be run with `--generate` or `--check`. In generate mode, it creates a file called `setup.tf` in the current directory containing the results of evaluating `tfsetup-config/setup.tmpl` with input containing the following fields:
* `Config` -- the data from `tfsetup-config/setup.tmpl`
* `Project` -- the data from `tfsetup-config.json`
* `Path` -- the relative path from the directory containing `tfsetup-config` to the local directory

If `tofu` or `terraform` is found in the path, `tfsetup` will run `tofu fmt` or `terraform fmt` on the resulting setup file. This is the only thing that makes this tool specific to terraform. Otherwise, it is just running the go templating system.

# Example

`tfsetup-config/context.json`:
```json
{
  "accounts": {
    "acct1": "123456789012",
    "acct2": "210987654321"
  }
}
```

`tfsetup-context.json`:
```json
{
  "account": "acct1"
}
```

`tfsetup-config/setup.tmpl`:
```
# Generated by tfsetup

data "external" "tfgen_check_generated" {
  program = ["tfsetup", "--check"]
}

locals {
  region         = "us-east-1"
  aws_account_id = "{{index .Config.accounts .Project.account}}"
}

terraform {
  backend "s3" {
    bucket         = "terraform-state-{{index .Config.accounts .Project.account}}"
    dynamodb_table = "terraform_locks"
    encrypt        = true
    key            = "examples/{{.Path}}/terraform.tfstate"
    region         = "us-east-1"
  }
}

provider "aws" {
  region              = local.region
  allowed_account_ids = [local.aws_account_id]
}
```

# Release Reminder

Update `Version` in main.go, then push tag.
