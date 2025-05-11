import json
import matplotlib.pyplot as plt
import subprocess
import numpy as np
from collections import defaultdict, Counter
import pandas as pd


def classify(value):
    if value < 0.5:
        return "Extinct"
    elif value <= 1.5:
        return "Stable"
    else:
        return "Expanding"


log_dir = "JSONlogs/output.json"
iters = 30
attach_types = ["Dismissive", "Fearful", "Preoccupied", "Secure", "Equal Split"]
props = np.identity(4)
props = np.vstack([props, [0.25, 0.25, 0.25, 0.25]])
data = defaultdict(list)

for row, type in zip(props, attach_types):
    print(type)
    D, F, P, S = row
    for _ in range(iters):
        subprocess.run(
            [
                "./tmtSimulator",
                "-numAgents=40",
                "-iters=200",
                "-rho=0.3",
                "-tau=0.35",
                "-mu=0",
                f"-dismissive={D}",
                f"-fearful={F}",
                f"-preoccupied={P}",
                f"-secure={S}",
            ]
        )

        with open(log_dir, "r") as file:
            GAME_DATA = json.load(file)
            CONFIG = GAME_DATA["Config"]
            init_agents = CONFIG["NumAgents"]
            FINAL_ITER = GAME_DATA["Iterations"][-1]
            final_agents = FINAL_ITER["NumberOfAgents"]
        pop_change = final_agents / init_agents
        data[type].append(pop_change)


records = {}

for agent, ratios in data.items():
    categories = [classify(r) for r in ratios]
    records[agent] = Counter(categories)


# ensure missing values are filled
for record in records.values():
    for category in ["Extinct", "Stable", "Expanding"]:
        record[category] = record.get(category, 0)


df = pd.DataFrame(records)
print(df)


fig, axes = plt.subplots(2, 3, figsize=(12, 12))
axes = axes.flatten()

for idx, agent_type in enumerate(attach_types):
    ax = axes[idx]
    series = df[agent_type].reindex(["Extinct", "Stable", "Expanding"])
    series = series / series.sum()
    ax.bar(series.index, series.values, color=["skyblue", "lightgreen", "salmon"])
    ax.set_title(agent_type)
    # ax.set_xticklabels(series.index)
    ax.set_ylim(0, 1)
    ax.set_ylabel("Proportion")

# Count the number of each outcome per agent
plt.show()
