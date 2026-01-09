package segments

import (
	"encoding/json"
	"net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Todoist struct {
	Base

	TaskCount int
}

type TasksResponse struct {
	Results []Task `json:"results"`
}

type Task struct {
	ID string `json:"id"`
}

func (t *Todoist) Enabled() bool {
	err := t.GetData()
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func (t *Todoist) Template() string {
	return "{{ .TaskCount }} "
}

func (t *Todoist) GetData() error {
	apikey := t.options.Template(APIKey, ".", t)

	httpTimeout := t.options.Int(options.HTTPTimeout, options.DefaultHTTPTimeout)

	addHeaders := func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+apikey)
		req.Header.Set("Accept", "application/json")
	}

	body, err := t.env.HTTPRequest("https://api.todoist.com/api/v1/tasks/filter?query=due today", nil, httpTimeout, addHeaders)
	if err != nil {
		return err
	}

	var response TasksResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	t.TaskCount = len(response.Results)

	return nil
}
