# PEQUOD

## Example Commands and use cases:
* New developer getting started with nothing installed.
* Existing developer with a hodgepodge of things installed in different ways (curl, brew, pkg)
* Current user upgrading tools to stay current

## Components
#### [Oh My ZSH](https://github.com/robbyrussell/oh-my-zsh)
A collection of opensource tools for a better terminal experience.

#### [Docker](https://www.docker.com/what-docker)
Allows you to run a single process like nginx in an isolated environment with very little overhead.

#### [Minikube](https://github.com/kubernetes/minikube)
You can't run docker containers on OSX natively yet so we have to virtualize an environment with access to a modern linux kernel with this tool. 

#### [Kubernetes & Kubectl](https://github.com/kubernetes/kubernetes)
A system for managing docker containers which are grouped into _pods_, orchestrated via _deployments_ and exposed to the outside world via _services_. Here is a great explanation: https://www.youtube.com/watch?v=4ht22ReBjno

#### [Helm](https://github.com/kubernetes/helm)
Managing deployments by hand can be tedious so we use helm to provide localized variables and dependencies among our infrastructure.

#### [VirtualBox]()
Virtualization platform that _minikube_ launches into.

#### [Google-cloud-sdk](https://cloud.google.com/sdk/)
Set of tools for manipulating resources in the Google Cloud Platform.


## Tasks
 
* Install some or all of the components ()
  * `argo install components ` prompts for each one.
  * `argo install components gcloud` executes and only solicits needed info.

* Uninstall Components
  * `argo uninstall components `
  * `argo uninstall components gcloud`
   
project needs argo.rc that has helm chart path/release
    - argo.rc needs dev/production cluster-name/context-info
        - enough info to obtain credentials for these clusters and point to the right part of them
    - mysql location information, credential information

* Project commands
  * `argo create PATH --webroot=[OPTIONAL WEB ROOT LOCATION] --sync`
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
    
  * `argo sync PATH`
    - Should grab database and files and insert into running argo infrastructure
        - warn and exit if not running
        
    - Use argo.rc in project to find dev cluster connection information
        - use kubectl port-forward to route dev/prod mysql to localhost and dump into running argo
        - same thing for nfs, temp-mount files and transfer over
    
    * `argo destroy PATH`
      - Helm delete