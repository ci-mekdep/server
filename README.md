# eMekdep: the high performance, composable  API.


## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Database Migration Versioning Style](#database-migration-versioning-style)
- [Git](#git)
- [API Versioning](#api-versioning)
- [Contributors](#contributors)


## Prerequisites

* Go language: https://golang.org/
* PostgreSQL: https://www.postgresql.org/


## Installation

Note: 
The `main` branch is the not development version of eMekdeo. Use the `dev` branch for installation and contributing.

The following steps need to be performed inside a terminal window.
```bash
git clone https://github.com/mekdep/server.git -b dev
```
use the git clone param `--depth 1` if the cURL 18 error occurred
```
cd server
```
then create a fresh Postgres database,
and create the .env file via
```
cp .env.example .env
vim .env
```
and setup the db settings and so on.

apply database migrations, eg:
```
sudo -u postgres psql
\c the_name_of_your_new_created_database
\conninfo
\i database/up.sql
\q
```
finally, run
```
go run .
```


## Database Migration Versioning Style

Versioning is crucial for tracking changes made to your database schema and data over time.

We use **Version Numbers:** Assign sequential version numbers to each migration script. This provides a clear linear history.


## Git

Please note we have a code of conduct, please follow it in all your interactions with the project.

### Conventional Commits

A specification for adding human and machine readable meaning to commit messages

[https://www.conventionalcommits.org/en/v1.0.0/](https://www.conventionalcommits.org/en/v1.0.0/)

### Contributing

When you are working on an issue, create another branch and then push it and create a PR. Eg:
```
git checkout -b author-category-issue-name
```
after code changes
```
git add .
git commit -m "feat: add new feature"
git push origin author-category-issue-name
```
after successful PR merge, delete the issue branch from local and remote
```
git checkout dev
git push -d origin author-category-issue-name
git branch -D author-category-issue-name
```


## API Versioning

The versioning scheme we use is [SemVer](https://semver.org/).


## Contributors

Thanks goes to all wonderful contributors ✨



*eMekdep* &copy; 2023-now, eMekdep Inc .
