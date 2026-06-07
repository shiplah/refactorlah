# RefactorLah -- deterministic refactoring for agents

A conservative refactoring CLI for humans and AI agents. It handles the common case where moving code is more than a filesystem operation, but still should not require a chain of separate tool calls.

Instead of manually combining `git mv`, namespace edits, import clean-up, reference updates, and follow-up validation, run one command and get a deterministic result.

`refactorlah` does **not** try to be a universal refactor engine. It rewrites only references it can prove from project configuration, and warns on anything uncertain.

> [!WARNING]
> This is currently a hacky pre-alpha experiment. It is useful for dogfooding and careful trials, but you should review its output and keep your project under version control before relying on it.

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

`move` applies by default. Use `--dry` to preview without writing.

See [.docs/usage.md](.docs/usage.md) for batch files, wildcard rules, JSON output, validation flags, and project configuration.

## What It Does

- moves files and directories through Git or the filesystem
- rewrites deterministic references for supported languages and frameworks
- validates replacement ranges before writing
- reports uncertain or dynamic references instead of guessing
- supports text and JSON output for humans and agents

Native support currently covers PHP, Python, Go, Symfony/Twig, and static asset imports. See [.docs/language-support.md](.docs/language-support.md) for the detailed support matrix and known gaps.

Project-level scan configuration is available through `.refactorlah.json`; see [.docs/usage.md](.docs/usage.md#configuration).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for tests, builds, release workflow, and adapter structure.

## Status

This is a safe working foundation, not a complete universal refactoring engine. It is designed to reduce fragile manual refactor workflows, especially for agents, while staying conservative about what it rewrites.
