<div align="center">

# gcpas (GCP Account Switcher)

[![stars](https://img.shields.io/github/stars/michalswi/gcp-account-switcher?style=for-the-badge&color=353535)](https://github.com/michalswi/gcp-account-switcher)
[![forks](https://img.shields.io/github/forks/michalswi/gcp-account-switcher?style=for-the-badge&color=353535)](https://github.com/michalswi/gcp-account-switcher/fork)
[![releases](https://img.shields.io/github/v/release/michalswi/gcp-account-switcher?style=for-the-badge&color=353535)](https://github.com/michalswi/gcp-account-switcher/releases)

manage multiple GCP accounts and projects without logging in every time.
Credentials are cached by `gcloud` — switching is instant after the first login.

</div>



## \# setup (one time per account)

```bash
gcpas add --name work    --account you@work.com   --project work-proj-id   --login
gcpas add --name personal --account you@gmail.com --project personal-proj  --login
```

`--login` opens a browser once to cache the token. You won't need it again unless the token expires.

Any fields not provided as flags are prompted interactively. To skip the prompts for region, zone, domain, and description:

```bash
gcpas add --name work --account you@work.com --project work-proj-id --skip
```

### all `add` flags

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Profile alias |
| `--account` | `-a` | GCP account email |
| `--project` | `-p` | GCP project ID |
| `--region` | `-r` | Default compute region |
| `--zone` | `-z` | Default compute zone |
| `--domain` | `-d` | Domain/org label (for reference) |
| `--desc` | | Description |
| `--login` | `-l` | Trigger `gcloud auth login` after adding |
| `--skip` | `-s` | Skip prompts for region, zone, domain, and description |

You can also snapshot your current active `gcloud` state into a profile:

```bash
gcpas init work
```

## \# switching accounts

```bash
gcpas use work         # switch instantly, no browser
gcpas use personal
gcpas use              # interactive numbered picker
gcpas use work --login # force re-authentication
```

## \# other commands

```bash
gcpas list             # show all profiles
gcpas current          # show active profile + gcloud state + authorized accounts
gcpas delete personal  # remove a profile (does not revoke gcloud auth)
```

## \# profiles are stored in

```
~/.gcp-switcher/profiles.json
```

Each profile holds: name, account email, project ID, region, zone, domain, description.
