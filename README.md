# gcps — GCP Account Switcher

Manage multiple GCP accounts and projects without logging in every time.
Credentials are cached by `gcloud` — switching is instant after the first login.

## Build & Install

```bash
go build -o gcps .
sudo mv gcps /usr/local/bin/
```

## Setup (one time per account)

```bash
gcps add --name work    --account you@work.com   --project work-proj-id   --login
gcps add --name personal --account you@gmail.com --project personal-proj  --login
```

`--login` opens a browser once to cache the token. You won't need it again unless the token expires.

You can also snapshot your current active `gcloud` state into a profile:

```bash
gcps init work
```

## Switching accounts

```bash
gcps use work         # switch instantly, no browser
gcps use personal
gcps use              # interactive numbered picker
gcps use work --login # force re-authentication
```

## Other commands

```bash
gcps list             # show all profiles
gcps current          # show active profile + gcloud state + authorized accounts
gcps delete personal  # remove a profile (does not revoke gcloud auth)
```

## Profiles are stored in

```
~/.gcp-switcher/profiles.json
```

Each profile holds: name, account email, project ID, region, zone, domain, description.
