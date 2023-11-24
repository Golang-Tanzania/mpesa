package mpesa

import (
	"net/http"

	"github.com/spf13/viper"
)

func (c *Client) LoadKeys(path, filename, filetype, env string) {

	viper.AddConfigPath(path)
	viper.SetConfigName(filename)
	viper.SetConfigType(filetype)
	viper.ReadInConfig()

	if err := viper.ReadInConfig(); err != nil {
		panic("Error reading config file, " + err.Error())
	}

	c.Keys = &Keys{
		PublicKey: viper.GetString("public_key"),
		ApiKey:    viper.GetString("api_key"),
	}

	c.Environment = env
}

func (c *Client) SetHttpClient(client *http.Client) {
	if client == nil {
		c.Client = http.DefaultClient
		return
	}
	c.Client = client
}
