package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/andygrunwald/go-jira"
)

var (
	appVer  = "1.0"
	appHash = "unset"
)

// AccountInfo is a type to hold the url, username, and password
// these values will be taken from the environment
type AccountInfo struct {
	jiraURL  string
	jiraUser jira.User
}

func main() {

	jiraUser := jira.User{
		Name:     os.Getenv("jira_username"),
		Password: os.Getenv("jira_password"),
	}

	account := AccountInfo{
		jiraURL:  os.Getenv("jira_url"),
		jiraUser: jiraUser,
	}

	var showVer bool
	var issueKey, timeToAdd, timeComment string

	flag.BoolVar(&showVer, "v", false, "Shows version")
	flag.StringVar(&issueKey, "ticket", "", "The ticket to add time to - please use the format of PROJ-ticketnumber")
	flag.StringVar(&timeToAdd, "time", "", "Please user Jira's time format of 1h 30m to log time")
	flag.StringVar(&timeComment, "worklog", "", "The worklog comment")
	flag.Parse()

	if showVer {
		stat, _ := os.Stat(os.Args[0])
		buildTime := stat.ModTime().Format("01/02/2006 03:04:05 PM MST")
		fmt.Printf("Version Number: %s\nGit Hash: %s\nUTC Build Time: %s", appVer, appHash, buildTime)
		os.Exit(0)
	}

	if issueKey == "" || timeToAdd == "" || timeComment == "" {
		fmt.Println("Invalid entry, please use -h for usage.")
		os.Exit(0)
	}
	if os.Getenv("jira_url") == "" || os.Getenv("jira_username") == "" || os.Getenv("jira_password") == "" {
		fmt.Println("This program requires environment variables be set for jira_username, jira_password with your account information, and jira_url for the base url for the Jira instance you want to log your time against")
		os.Exit(0)
	}

	CheckAccess(issueKey)

	worklogentry := jira.WorklogRecord{
		IssueID:   issueKey,
		Comment:   timeComment,
		TimeSpent: timeToAdd,
	}

	AddWorklog(account, worklogentry)
}

//CheckAccess is a function to make sure you can hit the issue
func CheckAccess(issueKey string) {

	jiraUser := jira.User{
		Name:     os.Getenv("jira_username"),
		Password: os.Getenv("jira_password"),
	}

	account := AccountInfo{
		jiraURL:  os.Getenv("jira_url"),
		jiraUser: jiraUser,
	}

	jiraClient, err := jira.NewClient(nil, account.jiraURL)
	if err != nil {
		panic(err)
	}
	jiraClient.Authentication.SetBasicAuth(account.jiraUser.Name, account.jiraUser.Password)

	issue, _, err := jiraClient.Issue.Get(issueKey, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Adding time to %s: %+v\n", issue.Key, issue.Fields.Summary)

}

// AddWorklog is a function to turn the data into the
func AddWorklog(account AccountInfo, entry jira.WorklogRecord) error {
	uri := fmt.Sprintf("%srest/api/2/issue/%s/worklog", account.jiraURL, entry.IssueID)

	worklogData := map[string]interface{}{
		"comment":   entry.Comment,
		"timeSpent": entry.TimeSpent,
		"started":   time.Now().UTC().Format("2006-01-02T15:04:05.000-0700"),
	}

	b, err := json.Marshal(worklogData)
	jsonStr := string(b)

	resp, err := makeRequestWithContent("POST", uri, jsonStr, account)
	if err != nil {
		return err
	}

	if resp.StatusCode == 201 {
		fmt.Println("Time added successfully")
		return nil
	}
	err = fmt.Errorf("Unexpected Response From POST")
	return err
}

func makeRequestWithContent(method string, uri string, content string, account AccountInfo) (resp *http.Response, err error) {
	buffer := bytes.NewBufferString(content)
	req, _ := http.NewRequest(method, uri, buffer)

	if resp, err = makeRequest(req, account); err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 {
		fmt.Println("401 response: %v", resp)
		req, _ = http.NewRequest(method, uri, bytes.NewBufferString(content))
		return makeRequest(req, account)
	}

	return resp, err
}

func makeRequest(req *http.Request, account AccountInfo) (resp *http.Response, err error) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	user := account.jiraUser.Name
	password := account.jiraUser.Password
	if password == "" {
		return nil, nil
	}

	req.SetBasicAuth(user, password)

	client := &http.Client{}
	if resp, err = client.Do(req); err != nil {
		fmt.Println("Failed to %s %s: %s", req.Method, req.URL.String(), err)
		return nil, err
	}

	runtime.SetFinalizer(resp, func(r *http.Response) {
		r.Body.Close()
	})

	if resp.StatusCode != 201 {
		fmt.Println("Return code from post request: ", resp.StatusCode)
		b, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Error Body: ", string(b))
	}

	return resp, nil
}
