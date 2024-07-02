import sys
import socket
import json
from channel import Channel
from aggregator import PersistentModelsBuffer, Aggregator, AvgStrategy
from nn import get_model
import time
import traceback

import subprocess 

#usage: python3 coordinator.py  <port>

MAX_PENDING_CONNECTIONS = 10

def custom_print(*args):
  print("[COORDINATOR]: ", *args)

def print_progress_bar(iteration, total, length=50):
    percent = ("{0:.1f}").format(100 * (iteration / float(total)))
    filled_length = int(length * iteration // total)
    bar = '█' * filled_length + '-' * (length - filled_length)
    sys.stdout.write(f'\r|{bar}| {percent}% Complete')
    sys.stdout.flush()
    
def __run_simulation(c, channels, agg, initial_model):
  tt = len(c["timeline"])
  # print_progress_bar(0, tt)
  indices = c["dataset"]["indices"]

  for idx, ch in enumerate(channels):
    c['dataset']['indices'] = indices[str(idx)]
    ch.send(c["dataset"])

  for ch in channels:
    ch.receive_ack()


  for i, t in enumerate(c["timeline"]):
    custom_print(f"Actions at time {i + 1}/{tt}: {len(t)}")
    for action in t:
      type = action[0]

      custom_print(f"Action: {type}, clients: {action[1:]}")

      if type == "AGG":
          initial_model = agg.aggregate()
          continue

      clients = action[1:]
      for c_idx in clients:
        if type == "COMM":
          channels[c_idx - 1].send_command("COMM")
          channels[c_idx - 1].receive_ack()
          weigths = channels[c_idx - 1].receive()
          agg.push(weigths)
        elif type == "RETR":
          channels[c_idx - 1].send_command("RETR")
          channels[c_idx - 1].receive_ack()
          channels[c_idx - 1].send(initial_model)

    # print_progress_bar(i + 1, tt)

  print()



def start_simulation(c, PORT, HOST):

  server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
  server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
  server_socket.bind((HOST, PORT))

  server_socket.listen(MAX_PENDING_CONNECTIONS)

  channels = []

  sub_processes = []
  for i in range(c["n_clients"]):
    # unlink stdout from the coordinator
    sub_processes.append(subprocess.Popen(["python3", "client.py", "localhost", str(PORT)]))
    # sub_processes.append(subprocess.Popen(["python3", "client.py", "localhost", str(PORT)], stdout=subprocess.PIPE, stderr=subprocess.PIPE))

  while len(channels) < c["n_clients"]:
    client_socket, _ = server_socket.accept()
    channels.append(Channel(client_socket))

  # model = MNISTModel({
  #   'lr': c['dataset']['learning_rate'],
  #   'momentum': c['dataset']['momentum'],
  #   'epochs': c['dataset']['epochs'],
  #   'batch_size': c['dataset']['batch_size'],
  #   'shuffle': False,
  # })
  
  c['dataset']['shuffle'] = True
  c['dataset']['network'] = 'SimpleCNN'
  c['dataset']['use_accelerator'] = True
  model = get_model(c['dataset'])

  intial_model = model.get_def_weights()

  MODELS_FOLDER = f"models/{c['name']}-{c['dataset']['name']}-out"


  agg = Aggregator(AvgStrategy(), PersistentModelsBuffer(MODELS_FOLDER))

  agg.set_default(intial_model)
  with open(f"{MODELS_FOLDER}/config.json", "w") as f:
    json.dump(c, f)

  try:
    __run_simulation(c, channels, agg, intial_model)
  except KeyboardInterrupt:
    custom_print("Simulation interrupted")
  except Exception as e:
    # custom_print("Unexpected error:", e.with_traceback())
    custom_print("Unexpected error:", traceback.format_exc())

  finally:
    for ch in channels:
      ch.close()

    server_socket.close()
    time.sleep(1)

    for p in sub_processes:
      p.send_signal(subprocess.signal.SIGINT)

    time.sleep(1)

    for p in sub_processes:
      if p.poll() is None:
        p.kill()

    for p in sub_processes:
      p.wait()

if __name__ == "__main__":  
  # load config from stdin input
  if len(sys.argv) < 2:
    custom_print("Usage: python3 coordinator.py <port>")
    sys.exit(1)
  
  PORT = int(sys.argv[1])

  custom_print("Starting coordinator, waiting for configuration")

  c = json.load(sys.stdin)  

  if isinstance(c, list):
    for i, config in enumerate(c):
      custom_print(f"Running simulation {i + 1}/{len(c)}")
      start_simulation(config, PORT, "localhost")

  else:
    start_simulation(c, PORT, "localhost")