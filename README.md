# Argo

## Installation
`brew install favish/favish/argo`

### Starting from scratch to get a functional environment
`argo components install && argo components start`  
Install will acquire all required binaries locally, start will make sure your local services are up and running.

#### Components installed:
* [Docker](https://www.docker.com/what-docker) - Allows you to run a single process like nginx in an isolated environment with very little overhead.
* [Minikube](https://github.com/kubernetes/minikube) - You can't run docker containers on OSX natively yet so we have to virtualize an environment with access to a modern linux kernel with this tool. 
* [Kubernetes & Kubectl](https://github.com/kubernetes/kubernetes) - A system for managing docker containers which are grouped into _pods_, orchestrated via _deployments_ and exposed to the outside world via _services_. Here is a great explanation [video](https://www.youtube.com/watch?v=4ht22ReBjno)
* [Helm](https://github.com/kubernetes/helm) - Managing deployments by hand can be tedious so we use helm to provide localized variables and dependencies among our infrastructure.
* [VirtualBox]() - Virtualization platform that _minikube_ launches into.
* [Google-cloud-sdk](https://cloud.google.com/sdk/) - Set of tools for manipulating resources in the Google Cloud Platform.

### Spin up project infrastructure locally
`argo project deploy`  
Creates all kubernetes services required to run the project in the cwd.  Optionally add `--debug` to see all commands run,
and `-y` to skip the prompt.

### Operating on dev/production

#### Deploying
Deployments should generally be left to CI to handle, but CI is only building the project dependencies via composer and then
running argo on the resulting artifact.  You can actually use the favish/build-deps image to do this without installing anything
locally.

If you need to bypass CI or are testing changes, you can run `argo project deploy -e=$DEPLOY_TARGET`
where deploy target is either `dev` or `prod`.  If no `CIRCLE_SHA1` variable is present in your shell when you run this command,
argo will use the `latest` tag for application code, you will most likely want to set this to match the git commit sha1 hash
that is currently on the infrastructure you're operating against (generally the last commit CI ran against).

