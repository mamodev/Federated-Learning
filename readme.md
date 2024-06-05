# Federated Learning Made Easy

This is a project wich aims to make federated learning easy & fast to develop

## Demo Requirements

To launch demo program you will need tmux, go-lang, python3.

_Tmux is needed to split and orgnanize terminal window to observe the system simulation.If you don't have tmux or your os does not support tmux, you can manually start the application_

Steps:

1. Enter demo folder: cd demo
2. Make bin dir: mkdir bin
3. Build server: go build -o ./bin ./cmd/server/
4. Run simulation: ./simulation.sh

---

![](demo.gif)

The server (panel 0:0) is responsable for coordination of N clients. It spawns 2 Task:
- 1) EVAL_TASK for only the oracle (machine wich have a complete dataset wich can give a "real" evautation.
- 2) TRAIN_TASK foreach client substcribed to registry.

Training process: 
Loop:
  1. Send Task to N clients (With latest model)
  2. Recive & Aggregate 
  3. Evaluate result
