package main

import "fmt"

type APICONTEXT struct {
  PUBLICKEY   string
  APIKEY      string
  ENVIRONMENT string
  ssl         bool
  address     string
  port        int
  headers     map[string]string
  parameters  map[string]string
}

func (api *APICONTEXT) setDefault() {

  api.address = "openapi.m-pesa.com"
  api.ssl = true
  api.port = 443

  // Defualt Headers

  api.headers = make(map[string]string)
  api.parameters = make(map[string]string)

  bearer := fmt.Sprintf("Bearer %v", api.createBearerToken(api.APIKEY))
  api.addHeader("Authorization", bearer)

  api.addHeader("Host", api.address)

  api.addHeader("Origin", "*")
  api.addHeader("Content-Type", "application/json")

}

func (api *APICONTEXT) getURL(endpoint string) string {
  url := ""
  if api.ssl {
    url = fmt.Sprintf("https://%v:%v%v", api.address, api.port, api.getPath(endpoint))
  } else {
    url = fmt.Sprintf("http://%v:%v%v", api.address, api.port, api.getPath(endpoint))
  }

  return url
}

func (api *APICONTEXT) addHeader(key, value string) {
  api.headers[key] = value
}

func (api *APICONTEXT) getHeaders() map[string]string {
  return api.headers
}

func (api *APICONTEXT) addParameter(key, value string) {
  api.parameters[key] = value
}

func (api *APICONTEXT) getParameters() map[string]string {
  return api.parameters
}

func (api *APICONTEXT) getPath(url string) string {
  if api.ENVIRONMENT == "production" {
    return fmt.Sprintf("/openapi/ipg/v2/vodacomTZN/%v/", url)
  } else {
    return fmt.Sprintf("/sandbox/ipg/v2/vodacomTZN/%v/", url)
  }
}