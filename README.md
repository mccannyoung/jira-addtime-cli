# jira-addtime-cli
A simple go program to add time to a JIRA issue via commandline.

This program is meant to be compiled and distributed to others. Because of this, it relies on certain environmental variables to be set. jira_username, jira_password and jira_url (for the base url for the jira -- for on-demand it would look like https://my-company.atlassian.net/) are required to be set on the machine running the executable generated from this code.

This app uses basic authentication, which isn't the safest option around.

uses:
```
\main.exe:
  -ticket string
        The ticket to add time to - please use the format of PROJ-ticketnumber
  -time string
        Please user Jira's time format of 1h 30m to log time
  -v    Shows version
  -worklog string
        The worklog comment
exit status 2
```
