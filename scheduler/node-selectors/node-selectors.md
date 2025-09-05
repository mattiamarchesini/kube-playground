## 1. Node Selector: The Simplest Filter

Node Selector (Attraction) 🧲: A Pod says, "I want to run on a node with this label."

A node selector is the most basic rule you can give the scheduler. It's a mandatory, black-and-white instruction.

How it works: The scheduler looks at the pod's nodeSelector field (e.g., disktype: ssd). It then filters the list of all
nodes, keeping only those that have the exact matching label(s). Any node without the label is discarded.

Role: A hard requirement. If no nodes match the selector, the pod remains unscheduled.

`kubectl label nodes <node-name> <key=value>`
