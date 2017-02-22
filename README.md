# Argo

## Installation
`brew install favish/favish/argo`

## Example Commands and use cases:
### Starting from scratch to get a functional environment
`argo components install && argo components start`  
Install will acquire all required binaries locally, start will make sure your local services are up and running.

### Spin up project infrastructure locally
`argo project deploy`  
Creates all kubernetes services required to run the project in the cwd.

## Components
* [Docker](https://www.docker.com/what-docker) - Allows you to run a single process like nginx in an isolated environment with very little overhead.
* [Minikube](https://github.com/kubernetes/minikube) - You can't run docker containers on OSX natively yet so we have to virtualize an environment with access to a modern linux kernel with this tool. 
* [Kubernetes & Kubectl](https://github.com/kubernetes/kubernetes) - A system for managing docker containers which are grouped into _pods_, orchestrated via _deployments_ and exposed to the outside world via _services_. Here is a great explanation [video](https://www.youtube.com/watch?v=4ht22ReBjno)
* [Helm](https://github.com/kubernetes/helm) - Managing deployments by hand can be tedious so we use helm to provide localized variables and dependencies among our infrastructure.
* [VirtualBox]() - Virtualization platform that _minikube_ launches into.
* [Google-cloud-sdk](https://cloud.google.com/sdk/) - Set of tools for manipulating resources in the Google Cloud Platform.


## Tasks
### Effectively TODO, ideal final tree of commands.  Not entirely implemented.

* Install some or all of the components ()
  * `argo components install` prompts for each one.
  * `argo components install gcloud` executes and only solicits needed info.

* Uninstall Components
  * `argo components uninstall  `
  * `argo components uninstall gcloud`
   
project needs argo.rc that has helm chart path/release
    - argo.rc needs dev/production cluster-name/context-info
        - enough info to obtain credentials for these clusters and point to the right part of them
    - mysql location information, credential information

* Project commands
  * `argo project create PATH --webroot=[OPTIONAL WEB ROOT LOCATION] --sync`
    - path default to .
    - path can be repo
        - if is repo, clone
        
    - create a kubernetes context derived from PATH (either cwd, or repo name)
    - set context to active context
    
    - helm install HELM-CHART(from argo.rc)
    - helm will need to be informed which directory to use to mount the project.
        - default $PWD/webroot
    
    - Notify user infrastructure is complete and they need to run argo sync to update database and files
        - or sync after if flag is present
    
  * `argo project sync PATH`
    - Should grab database and files and insert into running argo infrastructure
        - warn and exit if not running
        
    - Use argo.rc in project to find dev cluster connection information
        - use kubectl port-forward to route dev/prod mysql to localhost and dump into running argo
        - same thing for nfs, temp-mount files and transfer over
    
    * `argo project destroy PATH`
      - Helm delete
      
## Argo.rc spec
- Project Name
- Helm Chart
- dev cluster connection details
