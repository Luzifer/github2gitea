[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/github2gitea)](https://goreportcard.com/report/github.com/Luzifer/github2gitea)
![](https://badges.fyi/github/license/Luzifer/github2gitea)
![](https://badges.fyi/github/downloads/Luzifer/github2gitea)
![](https://badges.fyi/github/latest-release/Luzifer/github2gitea)

# Luzifer / github2gitea

`github2gitea` is a small tool to automatically create migrations inside a Gitea instance to mirror Github repositories.

For those wanting to mirror their Github repos to an own Gitea instance in case something happens to their Github user or organization you just need to execute this tool with the required parameters and for every mached repo (see below) a migration will be created to keep your Gitea repo up to date with the Github version.

When creating the migration

- the description will be set automatically
- the repo name will match the name of the repo on Github
- for private repos a Github token will be added to the sync URL
- for private repos on Github a private repo will be created in Gitea

## Usage

```console
# github2gitea --help
Usage of github2gitea:
  -n, --dry-run                    Only report actions to be done, don't execute them
      --gitea-token string         Token to interact with Gitea instance
      --gitea-url string           URL of the Gitea instance
      --github-token string        Github access token
      --log-level string           Log level (debug, info, warn, error, fatal) (default "info")
      --migrate-archived           Create migrations for archived repos
      --migrate-forks              Create migrations for forked repos
      --migrate-private            Migrate private repos (the given Github Token will be entered as sync credential!) (default true)
      --source-expression string   Regular expression to match the full name of the source repo (i.e. '^Luzifer/.*$')
      --target-user int            ID of the User / Organization in Gitea to assign the repo to
      --target-user-name string    Username of the given ID (to check whether repo already exists)
      --version                    Prints current version and exits
```

You can see there is a lot of options you need to set so here is a little walk-through:

| Option | Required | Description |
| ---- | :---: | ---- |
| `dry-run` | | You should enable it at first to have a look what github2gitea will do |
| `gitea-token` | X | Go to "Settings" in your profile menu, create a new access token |
| `gitea-url` | X | The URL your Gitea is available at. For example: `https://try.gitea.io/` |
| `github-token` | X | Fetch it in your user-settings under "Developer Settings" and assign `repo` permissions |
| `log-level` | | The levels used are `debug`, `info`, `warn`, `error` - Most users should use `info` |
| `migrate-archived` | | Set to `true` to also create migrations for archived repos |
| `migrate-forks` | | Set to `true` to also create migrations for forked repos |
| `migrate-private` | | Set to `false` not to create migrations for private repos |
| `source-expression` | X | Regular expression to match the *full name* of the repo: `^Luzifer/` will for example match `Luzifer/github2gitea` |
| `target-user` | X | ID of your Gitea user or organization (ask your instance admin to look it up in the site admin) |
| `target-user-name` | X | Name of the user the ID belongs to |

During the `dry-run` github2gitea will print warnings for any action not executed due to the dry run. Keep an eye on the logs.
