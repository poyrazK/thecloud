package sdk

import "fmt"

type Deployment struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	Name         string `json:"name"`
	Image        string `json:"image"`
	Replicas     int    `json:"replicas"`
	CurrentCount int    `json:"current_count"`
	Ports        string `json:"ports"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func (c *Client) CreateDeployment(name, image string, replicas int, ports string) (*Deployment, error) {
	req := struct {
		Name     string `json:"name"`
		Image    string `json:"image"`
		Replicas int    `json:"replicas"`
		Ports    string `json:"ports"`
	}{
		Name:     name,
		Image:    image,
		Replicas: replicas,
		Ports:    ports,
	}

	var dep Deployment
	err := c.post("/containers/deployments", req, &dep)
	return &dep, err
}

func (c *Client) ListDeployments() ([]Deployment, error) {
	var deps []Deployment
	err := c.get("/containers/deployments", &deps)
	return deps, err
}

func (c *Client) GetDeployment(id string) (*Deployment, error) {
	var dep Deployment
	err := c.get(fmt.Sprintf("/containers/deployments/%s", id), &dep)
	return &dep, err
}

func (c *Client) ScaleDeployment(id string, replicas int) error {
	req := struct {
		Replicas int `json:"replicas"`
	}{Replicas: replicas}
	return c.post(fmt.Sprintf("/containers/deployments/%s/scale", id), req, nil)
}

func (c *Client) DeleteDeployment(id string) error {
	return c.delete(fmt.Sprintf("/containers/deployments/%s", id), nil)
}
