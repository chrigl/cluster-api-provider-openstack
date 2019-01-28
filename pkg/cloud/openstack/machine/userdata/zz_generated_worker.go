package userdata

/*
This file is auto-generated DO NOT TOUCH!
*/

const (
	workerCloudConfig = `#cloud-config

write_files:
  - path: /etc/kubernetes/kubeadm_config.yaml
    permissions: "0444"
    encoding: b64
    content: {{.KubeadmConfig}}

merge_how: "list(append)+dict(recurse_array)+str()"

`
)
