package scriptrunner

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	"gopkg.in/yaml.v3"
)

const GitEndpoint = "https://api.github.com/repos/newrelic/newrelic-diagnostics-cli/contents/scriptcatalog"

type CatalogItem struct {
	Name        string   `yaml:"name"`
	Filename    string   `yaml:"filename"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`
	OS          string   `yaml:"os"`
	OutputFiles []string `yaml:"outputFiles"`
}

type GitHubResponse struct {
	Name             string `json:"name"`
	Path             string `json:"path"`
	Sha              string `json:"sha"`
	Size             int    `json:"size"`
	URL              string `json:"url"`
	HTMLURL          string `json:"html_url"`
	GitURL           string `json:"git_url"`
	DownloadURL      string `json:"download_url"`
	Type             string `json:"type"`
	Content          string `json:"content"`
	Encoding         string `json:"encoding"`
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	Links            struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		HTML string `json:"html"`
	} `json:"_links"`
}

type ICatalogDependencies interface {
	MakeRequest(name ...string) (*http.Response, error)
}

type Catalog struct {
	Deps ICatalogDependencies
}

type CatalogDependencies struct{}

func (c *Catalog) GetCatalog() ([]CatalogItem, error) {
	res, err := c.Deps.MakeRequest()
	if err != nil {
		return nil, err
	}
	b, err := c.checkAndReadResponse(res)
	if err != nil {
		return nil, err
	}
	var ghResp []GitHubResponse
	err = json.Unmarshal(b, &ghResp)
	if err != nil {
		return nil, err
	}
	list, err := c.process(ghResp)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (c *Catalog) GetScript(s CatalogItem) ([]byte, error) {
	return c.downloadFile("scripts/" + s.Filename)
}

func (c *Catalog) downloadFile(name string) ([]byte, error) {
	res, err := c.Deps.MakeRequest(name)
	if err != nil {
		return nil, err
	}
	b, err := c.checkAndReadResponse(res)
	if err != nil {
		return nil, err
	}
	var ghResp GitHubResponse
	err = json.Unmarshal(b, &ghResp)
	if err != nil {
		return nil, err
	}

	return b64.StdEncoding.DecodeString(ghResp.Content)
}

func (c *Catalog) checkAndReadResponse(res *http.Response) ([]byte, error) {
	if res.StatusCode == 403 {
		limitReset, err := strconv.ParseInt(res.Header.Get("X-Ratelimit-Reset"), 10, 64)
		if err != nil {
			return nil, errors.New("github rate limit exceeded")
		}
		currentTime := time.Now().Unix()
		remainingTime := limitReset - currentTime
		if remainingTime < 0 {
			return nil, errors.New("github rate limit exceeded")
		}
		return nil, fmt.Errorf("github rate limit exceeded. Rate limit resets in %s", time.Duration(remainingTime*int64(time.Second)))
	}
	return io.ReadAll(res.Body)
}

func (c *Catalog) process(ghRes []GitHubResponse) ([]CatalogItem, error) {
	if len(ghRes) < 1 {
		return []CatalogItem{}, nil
	}
	items := []CatalogItem{}
	for _, r := range ghRes {
		if strings.HasSuffix(r.Name, ".yml") {
			b, err := c.downloadFile(r.Name)
			if err != nil {
				return nil, err
			}
			ci := CatalogItem{}
			err = yaml.Unmarshal(b, &ci)
			if err != nil {
				return nil, err
			}
			items = append(items, ci)
		}
	}
	return items, nil
}

func (c *CatalogDependencies) MakeRequest(name ...string) (*http.Response, error) {
	wrapper := httpHelper.NewHTTPRequestWrapper()
	wrapper.Method = http.MethodGet
	wrapper.URL = GitEndpoint
	if name != nil && name[0] != "" {
		wrapper.URL += "/" + name[0]
	}
	wrapper.Headers["Accept"] = "application/vnd.github+json"
	wrapper.Headers["X-GitHub-Api-Version"] = "2022-11-28"
	return httpHelper.MakeHTTPRequest(wrapper)
}
