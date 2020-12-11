package rainforest

// EnvironmentParams are the parameters used to create a new Environment
type EnvironmentParams struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Environment represents an environment in Rainforest
type Environment struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

// CreateTemporaryEnvironment creates a new temporary environment and returns the
// Environment.
func (c *Client) CreateTemporaryEnvironment(urlString string) (*Environment, error) {
	body := EnvironmentParams{
		Name: "temporary-env-for-custom-url-via-CLI",
		URL:  urlString,
	}
	req, err := c.NewRequest("POST", "environments", &body)
	if err != nil {
		return nil, err
	}

	var env Environment
	_, err = c.Do(req, &env)
	if err != nil {
		return nil, err
	}

	return &env, nil
}

func (c *Client) IsEnvironmentDefault(id int) (bool, error) {

}

func (c *Client) SetEnvironmentDefault(id int, makeDefault bool) (bool, error) {

}
