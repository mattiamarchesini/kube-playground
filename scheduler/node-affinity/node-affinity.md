# 2. Node Affinity: The Advanced Filter & Scorer

Node affinity is a more expressive and flexible version of the node selector. It allows for more complex logic.
The point of **node affinity** is to give you much more powerful, flexible, and expressive control over where your pods
are scheduled compared to the simpler `nodeSelector`.

It allows you to define both strict requirements and weighted preferences:

- requiredDuringScheduling... (Hard Rule): This works just like a node selector but with more powerful logic (e.g., In,
  NotIn, Exists). The scheduler uses this as a filter to discard nodes that don't meet the criteria. A pod will not be
  scheduled if this rule isn't met.

- preferredDuringScheduling... (Soft Rule): This is where the scheduler's scoring phase comes in. The scheduler doesn't
  filter out nodes based on this rule. Instead, for each node that passed the filtering stage, the scheduler checks if
  it meets the preferred affinity rules. If it does, the node gets a higher score, making it a more desirable choice.
  The scheduler ultimately places the pod on the node with the highest total score.

Think of `nodeSelector` as a blunt instrument: "This pod _must_ run on a node with `label: value`."

`nodeAffinity` is like a full set of surgical tools. It lets you say things like:

- "This pod **must** run on a node in `us-east-1` that is **not** a GPU instance." (Complex "hard" rules).
- "This pod would **prefer** to run on a node with an SSD, but if none are available, a regular disk is fine." ("Soft"
  preferences).

---

## The Two Types of Node Affinity

Node affinity rules are defined within the `affinity` field of a Pod's specification. There are two main types:

### 1 `requiredDuringSchedulingIgnoredDuringExecution` (Hard Rule ✅)

This is a **strict requirement**. The scheduler **will not** schedule the pod unless a node that meets the specified
criteria is found. It's the more powerful successor to `nodeSelector`.

- **"Required During Scheduling"**: The rule must be met for the pod to be scheduled.
- **"Ignored During Execution"**: If the labels on the node change _after_ the pod is scheduled, the pod will not be
  evicted.

**The point of this is to:**

- **Use Complex Logic:** Go beyond the simple `AND` logic of `nodeSelector`. You can use operators like `In`, `NotIn`,
  `Exists`, and `DoesNotExist`.
- **Ensure Compliance/Security:** Guarantee that a pod handling sensitive data only runs on nodes within a specific
  security zone.
- **Guarantee Hardware:** Ensure a computationally intensive pod lands only on a node with a specific CPU model.

### Example: `OR` Logic

This pod **must** run in either `us-east-1` or `us-west-2`. This is impossible with `nodeSelector`.

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: topology.kubernetes.io/zone
              operator: In
              values:
                - us-east-1
                - us-west-2
```

### 2 `preferredDuringSchedulingIgnoredDuringExecution` (Soft Rule ⚖️)

This is a **preference**, not a requirement. The scheduler will try to find a node that meets the criteria and give it a
higher score, making it more likely to be chosen. If no such node is available, the pod will be scheduled on any other
available node.

**The point of this is to:**

- **Optimize Performance:** Try to place a latency-sensitive application on a node with a fast SSD (`disk=ssd`) but
  still allow it to run elsewhere if needed.
- **Improve Availability:** Prefer to spread pods across different zones to minimize the impact of an outage.
- **Lower Costs:** Prefer to schedule pods on cheaper Spot Instances, but allow them to run on more expensive On-Demand
  instances if no Spot Instances are available.

#### Example: Weighted Preference

This pod would _strongly prefer_ to run on a node with a high-performance SSD.

```yaml
affinity:
  nodeAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100 # A score from 1-100
        preference:
          matchExpressions:
            - key: disk-type
              operator: In
              values:
                - ssd-nvme
```

---

## Summary: `nodeSelector` vs. Node Affinity

| Feature         | `nodeSelector`                   | Node Affinity                                      |
| :-------------- | :------------------------------- | :------------------------------------------------- |
| **Rule Type**   | **Hard Rule Only**               | **Hard (`required`) and Soft (`preferred`) Rules** |
| **Logic**       | Simple `AND` of key-value pairs. | Rich operators (`In`, `NotIn`, `Exists`, etc.).    |
| **Flexibility** | Very limited.                    | Highly flexible for complex scenarios.             |
| **Analogy**     | A simple lock and key.           | A sophisticated combination lock.                  |

So, the point of **node affinity** is to move beyond basic pod placement and give you granular, flexible, and
intelligent control to optimize for performance, cost, and reliability in your cluster.
