# Extension points

The Kubernetes scheduler can be customized using scheduler plugins, which are compiled into the scheduler and activated
through configuration. These plugins register themselves at specific extension points in the scheduling cycle to add or
modify scheduling logic:

- scheduling queue
  - PrioritySort
- filtering
  - NodeResourcesFit
  - NodeName
  - NodeUnschedulable
  - TaintToleration
  - NodePorts
  - NodeAffinity
- scoring
  - NodeResourcesFit
  - ImageLocality
  - TaintToleration
  - NodeAffinity
- binding
  - DefaultBinder

## Scheduling Cycle

This is the main phase for selecting the best node for a pod.

**QueueSort**: Determines the order in which pods are taken from the scheduling queue. Only one QueueSort plugin can be
active. If you want to schedule high-priority pods before low-priority ones, this is the point to hook into.

**PreFilter**: Used for pre-processing information about a pod or doing initial checks before filtering nodes. For
example, it can check if the pod's requirements are even possible to meet in the cluster.

**Filter**: This is a critical extension point. Filter plugins are used to eliminate nodes that cannot run the pod. A
pod must pass the filter for all active plugins at this point. Examples include checking for sufficient CPU/memory,
matching node selectors, or respecting taints.

**PostFilter**: Called if no nodes are found for the pod after the filtering phase. It can be used for logging or
preemption logic (deciding if the pod can kick out other, lower-priority pods).

**PreScore**: A point for performing pre-computation before the scoring phase begins. This can be used to generate
shared data that all Score plugins can use, improving efficiency.

**Score**: After filtering, the scheduler scores each remaining viable node. Score plugins assign an integer score to
each node, with higher scores being better. The scheduler then combines the scores from all active Score plugins to rank
the nodes. For example, a plugin might give a higher score to a node with fewer running pods.

**Reserve**: This is the first point in the binding cycle. Once a node is selected, the Reserve plugin makes a note of
the resource reservation in the scheduler's internal state. This prevents other pods from being scheduled on that node
with outdated resource information while the current pod is being bound.

## Binding Cycle

This phase involves applying the decision to the cluster.

**Permit**: This point allows for a final check or delay before a pod is bound. A Permit plugin can approve, deny, or
delay the binding. This is useful for implementing things like scheduling quotas or external validation checks.

**PreBind**: Called right before the binding occurs. It can be used to prepare anything the node might need, such as
provisioning a network volume.

**Bind**: The core of the binding cycle. The Bind plugin is responsible for updating the Kubernetes API to assign the
pod to the selected node.

**PostBind**: A final hook called after the pod has been successfully bound. It's primarily used for cleanup or logging
purposes.

## References

- https://github.com/kubernetes/community/blob/master/contributors/devel/sig-scheduling/scheduling_code_hierarchy_overview.md
- https://kubernetes.io/blog/2017/03/advanced-scheduling-in-kubernetes/
- https://jvns.ca/blog/2017/07/27/how-does-the-kubernetes-scheduler-work/
- https://stackoverflow.com/questions/28857993/how-does-kubernetes-scheduler-work
