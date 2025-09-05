# 3. Taints and Tolerations: The Reverse Filter

Taints & Tolerations (Repulsion) 🚫: A Node says, "No pods can run on me unless they have a specific permission
(toleration)."

kubectl taint nodes <node-name> <key=value>:[NoSchedule|PreferNoSchedule|NoExecute]

To remove a taint add a - at the end
kubectl taint node controlplane node-role.kubernetes.io/control-plane:NoSchedule-

Taints work in the opposite direction. Instead of the pod choosing a node, the node repels pods.

How it works: A taint on a node (e.g., gpu=true:NoSchedule) marks it as off-limits. During its filtering phase, the
scheduler will automatically discard any tainted node.

The Exception (Tolerations): If a pod has a toleration that matches the node's taint, the scheduler ignores that taint
for that specific pod. This means the node is not filtered out and remains a candidate for placement.
