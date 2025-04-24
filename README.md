# GitHub Enterprise Cloud Migrator CLI

Building a CLI application for performing migrations between GitHub Enterprise Cloud organizations under the same EMU/Enterprise Cloud and GHES to GHEC migrations.

## Prerequisites

- You should have the [gh cli](https://cli.github.com/) installed.
- Be an administrator of your GitHub Enterprise Cloud or EMU instance. Ensure your token has the [proper permissions](https://docs.github.com/en/migrations/using-github-enterprise-importer/migrating-between-github-products/managing-access-for-a-migration-between-github-products#required-scopes-for-personal-access-tokens) - it is recommended to have `admin:org` for the validations to pass.

## Usage

Download the extension with the `gh cli`

```bash
gh extension install ryaugusta/gh-ec-migrator
```

Start the extension

```bash
gh ec-migrator
```

---
> [!NOTE]
> Set the `GH_TOKEN` environment variable with your GitHub PAT before running the program.
>
> The token should be authorized in both the target & source organizations
---
> [!IMPORTANT]
> This utility is currently in `beta`

## Contributing

Contributions are welcomed from the community.

**Want a new feature?**  Open an issue with your feature request.

**Want to contribute?** Fork this repository and open a pull request!
