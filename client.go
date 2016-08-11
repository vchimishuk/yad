// Copyright 2016 Viacheslav Chimishuk <vchimishuk@yandex.ru>
//
// Yad is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Yad is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Yad. If not, see <http://www.gnu.org/licenses/>.

package yad

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	// Default User-Agent header to use.
	defaultUserAgent = "Yad/0.1.0"
	// Default Yandex server address.
	defaultURL = "https://cloud-api.yandex.net/v1/disk/"
)

// Client is a Yandex.Disk REST client.
type Client struct {
	// URL specifies REST API calls base URL.
	URL string
	// HTTP is an actual HTTP client implementation.
	HTTP http.Client
	// UserAgent specifies User-Agent request header value.
	UserAgent string
	// token is a OAuth token of the application using Yandex.Disk.
	token string
}

// NewClient returns a new default configured Client object.
// Later all public fields can be assigned to change default behaviour:
// server's address can be changed via URL property, HTTP client via HTTP
// property and User-Agent header via UserAgent one.
func NewClient(token string) *Client {
	return &Client{
		URL:       defaultURL,
		UserAgent: defaultUserAgent,
		token:     token,
	}
}

// Stats returns Disk statistics (free space, used space, etc.).
func (c *Client) Stats() (*Stats, error) {
	body, err := c.do(http.MethodGet, "", url.Values{})
	if err != nil {
		return nil, err
	}

	s := &Stats{}
	err = json.Unmarshal([]byte(body), s)

	return s, err
}

// List returns one page of directory contents sorted by name.
func (c *Client) List(path string, offset int, limit int) (*ResourceList, error) {
	vals := url.Values{}
	vals.Set("path", path)
	vals.Set("offset", strconv.Itoa(offset))
	vals.Set("limit", strconv.Itoa(limit))
	vals.Set("sort", "name")

	body, err := c.do(http.MethodGet, "resources", vals)
	if err != nil {
		return nil, err
	}

	res := &Resource{}
	err = json.Unmarshal([]byte(body), res)
	if err != nil {
		return nil, err
	}
	if res.Type != ResourceTypeDir {
		return nil, errors.New("not a directory")
	}

	return res.SubList, nil
}

// ListAll returns the whole directory contents sorted by name.
func (c *Client) ListAll(path string) (*ResourceList, error) {
	list := &ResourceList{}
	offset := 0

	for {
		l, err := c.List(path, offset, 100)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, l.Items...)
		if len(l.Items) == 0 || len(l.Items) < 100 {
			break
		}
		offset += 100
	}
	list.Limit = len(list.Items)

	return list, nil
}

// Download reads file from the Disk and writes its content into given
// io.Writer. If path parameter points to a directory ZIP archive is
// written. Returns number of bytes written and error if any.
func (c *Client) Download(path string, w io.Writer) (int64, error) {
	vals := url.Values{}
	vals.Set("path", path)

	link, err := c.requestLink(http.MethodGet, "resources/download", vals)
	if err != nil {
		return 0, err
	}
	resp, err := c.newRequest(link.Method, link.Href, url.Values{}, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return io.Copy(w, resp.Body)
}

// Upload uploads fine contents readed from the given io.Reader
// to the disk resource located by the given path.
func (c *Client) Upload(path string, r io.Reader) error {
	vals := url.Values{}
	vals.Set("path", path)
	vals.Set("overwrite", "true")

	link, err := c.requestLink(http.MethodGet, "resources/upload", vals)
	if err != nil {
		return err
	}
	req, err := c.newRequest(link.Method, link.Href, url.Values{}, r)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server responsed %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	return nil
}

// UploadInternet upload file from the Internet, pointed by url, to Disk.
// Returns Link to an Operation which can be used to check uploading status.
func (c *Client) UploadInternet(path string, uri string) (*Link, error) {
	vals := url.Values{}
	vals.Set("path", path)
	vals.Set("url", uri)

	link, err := c.requestLink(http.MethodPost, "resources/upload", vals)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// Copy copies remote resource described by `from` path to new `path` location.
// Copy returns a Link object which can be operation for non empty directory or
// Link describing new object location for file and empty directory.
func (c *Client) Copy(path, from string) (*Link, error) {
	vals := url.Values{}
	vals.Set("path", path)
	vals.Set("from", from)
	vals.Set("overwrite", "true")

	link, err := c.requestLink(http.MethodPost, "resources/copy", vals)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// Move moves resource at path `from` to new path `path`.
// Move returns a Link object which can be operation for non empty directory or
// Link describing new object location for file and empty directory.
func (c *Client) Move(path, from string) (*Link, error) {
	vals := url.Values{}
	vals.Set("path", path)
	vals.Set("from", from)
	vals.Set("overwrite", "true")

	link, err := c.requestLink(http.MethodPost, "resources/move", vals)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// Delete deletes resource at path `from`.
// Delete returns a Link object which can be operation for non empty
// directory or nil for file and empty directory.
func (c *Client) Delete(path string, permanent bool) (*Link, error) {
	vals := url.Values{}
	vals.Set("path", path)
	vals.Set("permanently", fmt.Sprintf("%t", permanent))

	return c.requestOptionalOp(http.MethodDelete, "resources", vals)
}

// MkDir creates new empty directory.
// Returns Link to the new resource on success.
func (c *Client) MkDir(path string) (*Link, error) {
	vals := url.Values{}
	vals.Set("path", path)

	link, err := c.requestLink(http.MethodPut, "resources", vals)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// TrashDelete removes one resource from trash folder.
func (c *Client) TrashDelete(path string) (*Link, error) {
	// With empty path all trash objects are removed.
	if path == "" {
		return nil, errors.New("path is empty")
	}

	vals := url.Values{}
	vals.Set("path", path)

	return c.requestOptionalOp(http.MethodDelete, "trash/resources", vals)
}

// TrashDeleteAll removes all resources from trash folder.
func (c *Client) TrashDeleteAll() (*Link, error) {
	return c.requestOptionalOp(http.MethodDelete, "trash/resources",
		url.Values{})
}

// TrashRestore restores resource from trash under path `newPath`.
// If `newPath` is empty resource restored under its original path
// before deletion.
func (c *Client) TrashRestore(path, newPath string) (*Link, error) {
	vals := url.Values{}
	vals.Set("overwrite", "false")
	vals.Set("path", path)
	if newPath != "" {
		vals.Set("name", newPath)
	}

	link, err := c.requestLink(http.MethodPut, "trash/resources/restore", vals)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (c *Client) OpStatus(op *Link) (Status, error) {
	if !op.IsOperation() {
		return 0, errors.New("not operation")
	}

	vals := url.Values{}
	vals.Set("id", op.Operation())
	fmt.Println("id:", op.Operation())

	req, err := c.newRequest(op.Method, "operations", vals, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, err
	}
	body, err := readBody(resp)
	if err != nil {
		return 0, err
	}

	st := &struct {
		Status string `json:"status"`
	}{}
	if err := json.Unmarshal(body, st); err != nil {
		return 0, err
	}

	var s Status
	switch st.Status {
	case "success":
		s = StatusSuccess
	case "failure":
		s = StatusFailure
	case "in-progress":
		s = StatusInProgress
	default:
		return 0, fmt.Errorf("invalid status %s", st.Status)
	}

	return s, nil
}

func (c *Client) do(method string, url string, vals url.Values) (string, error) {
	req, err := c.newRequest(method, url, vals, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	body, err := readBody(resp)
	if err != nil {
		return "", err
	}

	return string(body), err
}

func (c *Client) requestLink(method string, url string, vals url.Values) (*Link, error) {
	body, err := c.do(method, url, vals)
	if err != nil {
		return nil, err
	}

	link := &Link{}
	err = json.Unmarshal([]byte(body), &link)

	if link.Templated {
		return nil, errors.New("unsupported templated link")
	}

	return link, err
}

func (c *Client) requestOptionalOp(method string, url string, vals url.Values) (*Link, error) {
	req, err := c.newRequest(method, url, vals, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	} else if resp.StatusCode == http.StatusAccepted {
		body, err := readBody(resp)
		if err != nil {
			return nil, err
		}
		link := &Link{}
		err = json.Unmarshal(body, &link)
		if err != nil {
			return nil, err
		}
		if link.Templated {
			return nil, errors.New("unsupported templated link")
		}

		return link, nil
	}

	return nil, readErrBody(resp)
}

func (c *Client) newRequest(method string, url string, vals url.Values,
	body io.Reader) (*http.Request, error) {

	req, err := http.NewRequest(method, c.URL+url, body)
	if err == nil {
		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Authorization", "OAuth "+c.token)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
	}
	req.URL.RawQuery = vals.Encode()

	return req, err
}

func readBody(r *http.Response) ([]byte, error) {
	defer r.Body.Close()

	if r.StatusCode < 200 || r.StatusCode >= 300 {
		return nil, readErrBody(r)
	}

	return ioutil.ReadAll(r.Body)
}

func readErrBody(r *http.Response) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	e := &Error{StatusCode: r.StatusCode}
	err = json.Unmarshal(b, e)
	if err != nil {
		return err
	}

	return e
}
