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
  * `argo install` prompts for each one.
  * `argo install gcloud` executes and only solicits needed info.

* Uninstall Components
  * `argo uninstall`
  * `argo uninstall gcloud`
   
* Launch project
  * `argo launch HELM_NAME`
  

Contexts?
  
  
