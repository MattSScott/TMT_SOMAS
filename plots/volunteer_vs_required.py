import json
import matplotlib.pyplot as plt

log_dir = "JSONlogs/output.json"
volunteers = []
required = []
actually_eliminated = []
turn_numbers = []
rho = None
tau = None

with open(log_dir, "r") as file:
    GAME_DATA = json.load(file)
    config = GAME_DATA["Config"]
    rho = config["PopulationRho"]
    tau = config["ASPThreshold"]
    turn_number = 0
    for ITER in GAME_DATA["Iterations"]:
        acc_elim = len(ITER.get("EliminatedAgents", []))
        num_vols = ITER["NumVolunteers"]
        num_req = ITER["TotalRequiredEliminations"]
        actually_eliminated.append(acc_elim)
        volunteers.append(num_vols)
        required.append(num_req)
        turn_numbers.append(turn_number)
        turn_number += 1


total_turns = len(turn_numbers)
simplified_x_ticks = range(0, total_turns + 1, 5)

# Plot
plt.figure(figsize=(14, 5))
plt.xlim(0, total_turns)
plt.plot(turn_numbers, volunteers, label="Volunteers", color="green")
plt.plot(turn_numbers, required, linestyle="--", label="Required", color="red")
plt.plot(
    turn_numbers,
    actually_eliminated,
    label="Eliminated",
    color="blue",
)
plt.xticks(simplified_x_ticks, rotation=-45)
plt.xlabel("Iteration")
plt.ylabel("Number of Agents")
plt.title(rf"Volunteers vs Required Eliminations per Turn ($\rho$={rho}, $\tau$={tau})")
# plt.ylim(0, max(volunteers + required) + 1)  # Auto-scaled y-axis
plt.legend()
plt.grid(True)
plt.tight_layout()
plt.show()
