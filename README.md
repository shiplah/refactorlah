# RefactorLah — deterministic refactoring for agents

A conservative refactoring CLI for humans and AI agents. It handles the common case where moving code is more than a filesystem operation, but still should not require a chain of separate tool calls. Instead of manually combining `git mv`, namespace edits, import clean-up, reference updates, and follow-up validation, run one command and get a deterministic result.

`refactorlah` does **not** try to be a universal refactor engine. It rewrites only references it can prove from project configuration, and warns on anything uncertain.

> [!WARNING]
> RefactorLah is experimental and still needs dogfooding to get better. As far as we know it is usable for careful trials, but you should review its output and keep your project under version control.

## Install

Until release archives are published, install from the repository:

```bash
git clone git@github.com:NickSdot/refactorlah.git
cd refactorlah
bin/install.sh
```

The installed command is source-checkout-independent. PHP and Python refactors do not require PHP, Composer, or Python on the target machine.

To use a different install directory:

```bash
bin/install.sh ~/bin
```

Source installs require Go with cgo support so the native language parsers can be compiled into the CLI.

## Command Usage

```bash
refactorlah move app/Services/Billing app/Domain/Billing
refactorlah move app/Services/Billing app/Domain/Billing --dry
refactorlah move app/Services/Billing/InvoiceService.php app/Domain/Billing/InvoiceService.php
refactorlah move src/app/services/billing.py src/app/domain/billing.py
refactorlah move --use-list app/Foo.php,app/Bar.php tests/A.php,tests/B.php
refactorlah move --use-file moves.txt
refactorlah move 'src/Old/*Worker.php' 'src/New/*Rule.php'
refactorlah move app/Services/Billing app/Domain/Billing --format=json
```

See [.docs/usage.md](.docs/usage.md) for batch files, wildcard rules, JSON output, and validation flags.

## What It Does

- moves files and directories through Git or the filesystem
- rewrites deterministic references for [supported languages and frameworks](.docs/language-support.md)
- validates replacement ranges before writing
- reports uncertain or dynamic references instead of guessing
- supports text and JSON output for humans and agents

Current language and framework support includes PHP, Python, Go, Symfony/Twig, and static asset imports.

## Configuration

Projects may add `.refactorlah.json` at the command's working directory or up to three directory levels below it to exclude semantic scans for generated, fixture, or stub files:

```json
{
  "exclude": [
    "local/phpstan/tests/fixtures/**"
  ],
  "include": []
}
```

`include` entries override `exclude` entries. The core still plans requested moves; this config only limits semantic rewrites and warnings.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for tests, builds, release workflow, and adapter structure.

## License

RefactorLah is released under the [MIT License](LICENSE).
