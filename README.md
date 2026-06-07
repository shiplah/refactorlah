# RefactorLah — deterministic refactoring for agents

A conservative refactoring CLI for humans and AI agents. It handles the common case where moving code is more than a filesystem operation, but still should not require a chain of separate tool calls. Instead of manually combining `git mv`, namespace edits, import clean-up, reference updates, and follow-up validation, run one command and get a deterministic result.

`refactorlah` does **not** try to be a universal refactor engine. It rewrites only references it can prove from project configuration, and warns on anything uncertain.

> [!WARNING]
> RefactorLah is experimental and still needs dogfooding to get better. As far as we know it is usable for careful trials, but you should review its output and keep your project under version control.

## Install

### Prebuilt Binaries

Download the archive for your platform from [GitHub Releases](https://github.com/NickSdot/refactorlah/releases), extract it, and put `refactorlah` on your `PATH`.

Current release targets:

- macOS Apple Silicon
- Linux ARM64
- Windows ARM64

### From Source

Source installs require Go and a working C toolchain.

```bash
git clone git@github.com:NickSdot/refactorlah.git
cd refactorlah
bin/install.sh # installs to ~/.local/bin by default; bin/install.sh ~/foo for different locations
```

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

See [.docs/usage.md](.docs/usage.md) for batch files, wildcard rules, path resolution, scan scope, JSON output, and validation flags.

## What It Does

- moves files and directories through Git or the filesystem
- rewrites deterministic references for [supported languages and frameworks](.docs/language-support.md)
- validates replacement ranges before writing
- reports uncertain or dynamic references instead of guessing
- supports text and JSON output for humans and agents

Current language and framework support includes PHP, Python, Go, Symfony/Twig, and static asset imports.

## Configuration

Projects may add `.refactorlah.json` at the command's working directory or up to three directory levels below it to exclude generated, fixture, or stub files from refactoring:

```json
{
  "exclude": [
    "local/phpstan/tests/fixtures/**"
  ],
  "include": [],
  "checks": [
    ["composer", "stan"]
  ],
  "tests": [
    ["composer", "test"]
  ]
}
```

`include` entries override `exclude` entries. Excluded paths are not moved, semantically rewritten, or reported as semantic warnings. Configured `checks` run after apply. Configured `tests` run only with `--run-tests`.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for tests, builds, release workflow, and adapter structure.

## License

RefactorLah is released under the [MIT License](LICENSE).
