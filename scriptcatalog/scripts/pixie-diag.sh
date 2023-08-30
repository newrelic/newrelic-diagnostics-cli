#!/bin/bash

# check for namespace
if [ -z "$1" ]
  then
    echo "No namespace passed"
    echo "usage: pixie-diag.sh <namespace>"
    exit 0
fi

# Timestamp
timestamp=$(date +"%Y%m%d%H%M%S")

# namespace variable
namespace=$1

# Create a log file
exec > >(tee -a "$PWD/pixie_diag_$timestamp.log") 2>&1

echo -e "\n*****************************************************\n"
echo -e "Checking agent status\n"
echo -e "*****************************************************\n"

# Check for px
if ! [ -x "$(command -v px)" ]; then
  echo 'Error: px is not installed.' >&2
else
  echo "Get agent status from Pixie"
  px run px/agent_status
  # Skip if unable to get agent status
  if [ $? -eq 0 ]; then
    echo ""
    echo "Collect logs from Pixie"
    px collect-logs
  fi
fi

# Check HELM releases
echo -e "\n*****************************************************\n"
echo -e "Checking HELM releases\n"
echo -e "*****************************************************\n"

helm list -A -n $namespace

# Check System Info
echo -e "\n*****************************************************\n"
echo -e "Key Information\n"
echo -e "*****************************************************\n"

nodes=$(kubectl get nodes | awk '{print $1}' | tail -n +2)

# check node count
nodecount=$(kubectl get nodes --selector=kubernetes.io/hostname!=node_host_name | tail -n +2 | wc -l)
echo "Cluster has "$nodecount" nodes"

if [ $nodecount -gt 100 ]
  then
    echo "Node limit is greater than 100"
fi

# check node memory capacity
MEMORY=$(kubectl get nodes -o jsonpath='{.items[0].status.capacity.memory}' | sed 's/Ki$//')
echo "MEMORY=$MEMORY"
if [[ "$MEMORY" -lt 7950912 ]]; then
echo "Pixie requires nodes with 8 Gb of memory or more, got ${MEMORY}."
fi

# pods not running
podsnr=$(kubectl get pods -n newrelic | grep -v Running | tail -n +2 | awk '{print $1}')

# count of pods not running
podsnrc=$(printf '%s\n' $podsnr | wc -l)

if [ $podsnrc -gt 0 ]
  then
    echo "There are $podsnrc pods not running!"
    echo "These pods are not running"
    printf '%s\n' $podsnr
fi

echo -e "\n*****************************************************\n"
echo -e "Node Information\n"
echo -e "*****************************************************\n"

for node_name in $nodes
  do
    # Get K8s version and Kernel from nodes
    echo ""
    echo "System Info from $node_name"
    kubectl describe node $node_name | grep -i 'Kernel Version\|OS Image\|Operating System\|Architecture\|Container Runtime Version\|Kubelet Version'
    done

# Check Allocated resources Available/Consumed
echo -e "\n*****************************************************\n"
echo -e "Checking Allocated resources Available/Consumed\n"
echo -e "*****************************************************\n"

for node_name in $nodes
  do
    # Get Allocated resources from nodes
    echo ""
    echo "Node Allocated resources info from $node_name"
    kubectl describe node $node_name | grep "Allocated resources" -A 9
  done

# Get kubectl describe node output for 3 nodes
echo -e "\n*****************************************************\n"
echo -e "Collecting Node Detail (limited to 3 nodes)\n"
echo -e "*****************************************************\n"

nodedetailcounter=0
for node_name in $nodes
  do
    if [ $nodedetailcounter -lt 3 ]
    then
      # Get node detail from a sampling of nodes
      echo -e "\nCollecting node detail from $node_name"
      kubectl describe node $node_name
      let "nodedetailcounter+=1"
    else
      break
    fi
  done

#Get all Kubernetes resources in namespace

echo -e "\n*****************************************************\n"
echo -e "Check all Kubernetes resources in namespace\n"
echo -e "*****************************************************\n"

# Get all api-resources in namespace
for i in $(kubectl api-resources --verbs=list -o name | grep -v "events.events.k8s.io" | grep -v "events" | sort | uniq);
do
  echo -e "\nResource:" $i;
  # An array of important namespace resources
  array=("configmaps" "rolebindings.rbac.authorization.k8s.io" "endpoints" "secrets" "networkpolicies" "serviceaccounts" "pods" "endpointslices" "deployments.apps" "horizontalpodautoscalers" "ingresses" "networkpolicies")
  str=$'resources\n=================='
  if [[ "${array[*]}" =~ "$i" ]]; then
    echo -e "\n px-operator $str"
    kubectl -n px-operator get --ignore-not-found ${i};
    echo -e "\n $namespace $str"
    kubectl -n $namespace get --ignore-not-found ${i};
  else 
    echo -e "\n $namespace $str"
    kubectl -n $namespace get --ignore-not-found ${i};
  fi
done

nr_deployments=$(kubectl get deployments -n $namespace | awk '{print $1}' | tail -n +2)
olm_deployments=$(kubectl get deployments -n olm | awk '{print $1}' | tail -n +2)
px_deployments=$(kubectl get deployments -n px-operator | awk '{print $1}' | tail -n +2)

for deployment_name in $nr_deployments $olm_deployments $px_deployments
  do
    # Get logs from deployments
    if [[ $deployment_name =~ ^newrelic-bundle-nri-kube-events.*$ ]];
    then
      echo -e "\n*****************************************************\n"
      echo -e "Logs from $deployment_name container: kube-events\n"
      echo -e "*****************************************************\n"
      kubectl logs --tail=50 deployments/$deployment_name -c kube-events -n $namespace
      echo -e "\n*****************************************************\n"
      echo -e "Logs from $deployment_name container: forwarder\n"
      echo -e "*****************************************************\n"
      kubectl logs --tail=50 deployments/$deployment_name -c forwarder -n $namespace
    else
      if [[ $deployment_name == "vizier-operator" ]]; then
        ns="px-operator"
      elif [[ $deployment_name == "catalog-operator" || $deployment_name == "olm-operator" ]]; then
        ns="olm"
      else
        ns=$namespace
      fi

      echo -e "\n*****************************************************\n"
      echo -e "Logs from $deployment_name\n"
      echo -e "*****************************************************\n"
      kubectl logs --tail=50 deployments/$deployment_name -n $ns
    fi
  done

echo -e "\n*****************************************************\n"
echo -e "Checking pod events\n"
echo -e "*****************************************************\n"

pods=$(kubectl get pods -n $namespace | awk '{print $1}' | tail -n +2)

for pod_name in $pods
  do
    # Get events from pods in New Relic namespace
    echo ""
    echo "Events from pod name $pod_name"
    kubectl get events --all-namespaces --sort-by='.lastTimestamp'  | grep -i $pod_name
    done

gzip -9 -c pixie_diag_$timestamp.log > pixie_diag_$timestamp.log.gzip

echo -e "\n*****************************************************\n"
echo -e "File created = pixie_diag_<timestamp>.log\n"
echo -e "File created = pixie_diag_<timestamp>.log.gzip\n"
echo -e "*****************************************************\n"
echo "End pixie-diag"
