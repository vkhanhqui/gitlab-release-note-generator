# Gitlab Release Note Generator

A Gitlab release note generator that generates release note on latest tag

Golang version of [gitlab-release-note-generator](https://github.com/jk1z/gitlab-release-note-generator)

## Feature
-  Generate release note on the latest tag based on merge requests and issues
-  Distinguished title with issues or merge requests that have the following labels: **enhancement**, **breaking change**, **feature** and **bug**

   *(Note. if an issue or merge request that has 2 or more labels, that issue or merge request will be displayed again under the corresponding title)*

-  Can be integrated as a CD service. Tutorial below


## How it works
1. Find the latest tag
2. Find the previous tag that is on the same branch as the latest tag.
3. Locate the date range between the latest and the previous tag. If there is only a tag in the project, then the `from` date will be the project creation date and the `to` date will be that tag's creation date.
4. Find all **Merged** merge requests and **Closed** issues within that time range
5. Generate a release note/changelog based on the findings above.

## How to run this app

1. Create an .env with these vars
```
GITLAB_API_ENDPOINT=''
GITLAB_PERSONAL_TOKEN=''
GITLAB_PROJECT_ID=''
TARGET_BRANCH='main'
TARGET_TAG_REGEX='^release-.*$'
TZ='Asia/Saigon'
ISSUE_CLOSED_SECONDS=0
```

2. Run
```
go run main.go
```


## Options

These can be specified using environment variables

* ```GITLAB_API_ENDPOINT```: Your gitlab instance's endpoint, eg: ```https://gitlab.com/api/v4```
* ```GITLAB_PERSONAL_TOKEN```: A gitlab personal access token with `api` permission. [How to Tutorial](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html)
* ```GITLAB_PROJECT_ID```: Your project id that is located under ```settings > general```
* ```TARGET_BRANCH```: The branch to look for release tags, eg: ```main```
* ```TARGET_TAG_REGEX```:  Regular expression of the release tags to search, eg: ```^release-.*$```
* ```TZ```: The timezone for your release notes, eg: ```Asia/Saigon```
* ```ISSUE_CLOSED_SECONDS```: The amount of seconds to search after the last commit,  useful for Merge Requests that close their tickets a second after the commit, eg: ```0```


## Credits
Also, thanks to [github-changelog-generator](https://github.com/github-changelog-generator/github-changelog-generator)
