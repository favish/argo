package schemas

import (
	"github.com/spf13/viper"
	"bytes"
	"os"
	"github.com/fatih/color"
)

// True/false for required values
var drupal = []byte(`
  gcp:
    project: false
    cluster: false
    region: false
    compute_zone: false
  network:
    hostname: false
    container_cidr: false
    redirect_www: false
    use_ingress: false
  applications:
    basic_auth:
      enabled: false
      b64_passwd: false
      node_port: false
    cron:
      key: false
      host: false
    drupal:
      autoscale:
        enabled: false
        min: false
        max: false
        targetCpu: false
      env: false
      image: false
      local:
        project_root: false
        theme_dir: false
      resources:
        requests:
          cpu: false
          memory: false
        limits:
          cpu: false
          memory: false
    mysql:
      cloudsql_instance: false
      db: true
      user: true
      pass: true
      host: false
    nfs:
      enabled: true
      service_ip: false
      requested_storage: false
    redis:
      resources:
        requests:
          cpu: false
          memory: false
    varnish:
      node_port: false
      secret: false
      secondary_host: false
    xdebug:
      host_ip: false
`)

var DrupalSchema = viper.New()

func init() {
	DrupalSchema.SetConfigType("yaml")
	err := DrupalSchema.ReadConfig(bytes.NewBuffer(drupal))
	if (err != nil) {
		color.Red("Error in schema for drupal: %s", err)
		os.Exit(1)
	}
}