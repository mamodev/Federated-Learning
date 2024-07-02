import subprocess
import struct
import socket
import json
import time
import os
import csv

import torch
import torch.optim as optim
import torchvision

from network import Net
from fed.py_torch import PyTorchModel


test_loader = None

server_bin_path = "../bin/server"
server_args = ["../model/original.npz"]

client_py_path = "client.py"

local_folder = "../python"

ssh_remote_folder = "/Users/marco/PyTest"
ssh_remote_host = "192.168.1.16"
ssh_username = "marco"
ssh_password = "1234"

save_folder = "save"
if not os.path.exists(save_folder):
    os.makedirs(save_folder)

def loader_fact():
    global test_loader
    if test_loader is None:
        test_dataset = torchvision.datasets.MNIST(
            './data/', 
            train=False, 
            download=True, 
            transform=torchvision.transforms.Compose([
                torchvision.transforms.ToTensor(),
                torchvision.transforms.Normalize((0.1307,), (0.3081,))
            ])
        )

        test_loader = torch.utils.data.DataLoader(test_dataset, batch_size=250, shuffle=False)

    return test_loader


def eval_model (model, round, csv_writer):
  print(f"== [Server] Evaluating model for round {round}")
  net = Net()
  optimizer = optim.SGD(net.parameters(), lr=0.01, momentum=0.9)
  net = PyTorchModel(net, optimizer, loader_fact, loader_fact)

  try: 
    eval = net.evaluate(model, {})
    csv_writer.writerow([round, eval["total"], eval["correct"], eval["loss"], eval["accuracy"]])
    print(f"== [Server] Evaluated model: {eval}")
  except Exception as e:
    print(f"Error evaluating model: {e}")
  pass


def start_server (params):
  return subprocess.Popen([server_bin_path] + server_args,
                          stdout=subprocess.PIPE,
                          stdin=subprocess.PIPE)

# Client Parameters:
# {
#   "client_idx": int,
#   "n_clients": int,
#   "data_round_split": int < MAX_ROUNDS,
#   "data_round_split_idx": int < data_round_split,
#   "knolodge_amount": 0.01,
#   "bias": 0.1,
#   "noise": 0.1,

#   "epochs": int,

#   "agg_host": "domain or ip",
#   "agg_port": int (port),
# }

#  Server Parameters:
# {
#   "min_for_agg": 10,
# }

# Coordinator Parameters:
# {
#   "test_name": "test1",
#   "max_rounds": 10,
#   "model_retrive_distribution": [0.1, 0.2, 0.3, 0.4], (10% 0 rounds, 20% 1 rounds, 30% 2 rounds, 40% 3 rounds) 
#   "model_commit_distribution": [0.1, 0.2, 0.3, 0.4], (10% 0 rounds, 20% 1 rounds, 30% 2 rounds, 40% 3 rounds) 
# }


def map_percentages_to_integers(percentages, total):
    mapped_list = []
    remaining_total = total

    for percentage in percentages:
        if remaining_total <= 0:
            mapped_list.append(0)  # if no more rounds can be added
        else:
            rounds = int(percentage * total)
            mapped_list.append(rounds)
            remaining_total -= rounds

    # Adjust if there's any remaining total that wasn't allocated due to rounding
    if remaining_total > 0:
        for i in range(len(mapped_list)):
            if remaining_total <= 0:
                break
            if mapped_list[i] < total:
                mapped_list[i] += 1
                remaining_total -= 1

    return mapped_list



def save_params_to_file(params):
  test_folder = f"{save_folder}/{params['test_name']}"
  if not os.path.exists(test_folder):
    os.makedirs(test_folder)

  with open(f"{save_folder}/{params['test_name']}/params.json", "w") as file:
    json.dump(params, file)

  pass

def start_simulation(params, server, clients): 

  save_params_to_file(params)

  map_prc_retrive = map_percentages_to_integers(params["model_retrive_distribution"], params["n_clients"])

  # print(map_prc_retrive)
  # print(params["model_commit_latency"])

  client_commit_latency = [0 for _ in range(params["n_clients"])]
  client_retrive_time = [0 for _ in range(params["n_clients"])]

  ret_idx = 0
  for i in range(params["n_clients"]):
    if map_prc_retrive[ret_idx] <= 0:
      ret_idx += 1

    if map_prc_retrive[ret_idx] > 0:
      client_retrive_time[i] = 0
      client_commit_latency[i] = params["model_commit_latency"][ret_idx]
      map_prc_retrive[ret_idx] -= 1

  

  # print(client_commit_latency)
  # print(client_retrive_time)

  client_cicle = [start + latency + 1 for latency, start in zip(client_commit_latency, client_retrive_time)]
  # print(client_cicle)
  round = 0
  real_round = 0

  client_data = zip(clients, client_cicle, client_retrive_time, client_commit_latency)
  client_data = list(client_data)

  csv_writer = csv.writer(open(f"{save_folder}/{params['test_name']}/data.csv", "w"))
  csv_writer.writerow(["Round", "Total", "Correct", "Loss", "Accuracy"])

  while real_round < params["max_rounds"]:
    print(f"Round {round} - Real Round {real_round}")
    start_client_set = [client for client, size, retrive_time, latency in client_data if retrive_time % size == round % size]
    stop_client_set = [client for client, size, retrive_time, latency in client_data if retrive_time + latency % size == round % size]

    debug_0 = [client.name for client, size, retrive_time, latency in client_data]
    debug_1 = [retrive_time % size for client, size, retrive_time, latency in client_data]
    debug_2 = [retrive_time + latency % size for client, size, retrive_time, latency in client_data]

    # print(debug_0, len(client_data))
    # print(debug_1)
    # print(debug_2)

    for idx, client in enumerate(start_client_set): 
      client.download_model(params, idx, real_round)

    for client in stop_client_set:
      client.upload_model()

    if len(stop_client_set) > 0:
      server.stdin.write("AGGREGATE".encode('latin-1'))
      server.stdin.flush()

      len_bytes = server.stdout.read(4)
      if len_bytes == b'':
          break

      length = struct.unpack('>I', len_bytes)[0]

      buff = b''
      while len(buff) < length:
        buff += server.stdout.read(length)

      eval_model(buff, real_round, csv_writer)
      real_round += 1

    round += 1




class Client:
  def __init__(self, name, socket, address):
    self.name = name
    self.socket = socket
    self.address = address

  def upload_model (self):
    print(f"== [{self.name}] Uploading model") 
    # write SEND str
    self.sendData("SEND".encode('latin-1'))
    # read OK
    ack = self.reciveData()
    if ack.decode('latin-1') != "OK":
      print(f"== [{self.name}] Error: Expected OK but got {ack.decode('latin-1')}")
      return
    pass

  def download_model (self, params, idx, round):
    params["data_round_split_idx"] = round
    params["client_idx"] = idx

    print(f"== [{self.name}] Downloading model")
    # write json parameters
    self.sendData(json.dumps(params).encode('latin-1'))
    pass

  def handshake(self):
    # Send client name
    name = self.reciveData().decode('latin-1')
    self.sendData(name.encode('latin-1'))

    self.name = name

    print(f"Connected to controller")

  def reciveAll(self, length):
    readed = 0
    data = b''
    while readed < length:
        rec = self.socket.recv(length - readed)
        readed += len(rec)
        data += rec

    return data

  def reciveData(self):
    # Receive data length
    length_bytes = self.reciveAll(4)
    length = struct.unpack('>I', length_bytes)[0]

    data = self.reciveAll(length)
    return data


  def sendData(self, data):
    length = len(data)
    self.socket.sendall(struct.pack('>I', length))
    self.socket.sendall(data)



if __name__ == "__main__":
    rounds = 30
    params = {
      # Test Parameters
      "test_name": "test3",
      "model_retrive_distribution": [0.5, 0.4],
      "model_commit_latency": [1, 2],
      "max_rounds": rounds,

      "n_clients": 1,
      "data_round_split": 1,
      "knolodge_amount": 1,
      
      "epochs": 6,
      "batch_size": 50,
      "noise": 0,
      "bias": 0,

      "agg_host": "192.168.1.47",
      "agg_port": 8080,
      "CAP": 1000,
    }

    # n_clients = [1, 2, 4, 6, 8, 10]
    # n_clients = [1, 2, 4, 6, 8, 10]
    model_commit_latency=[
      # [1, 1],
      # [1, 2],
      # [1, 3],
      [1, 6],
      [1, 10],
    ]

    n_clients = [4]
    max_clients = max(n_clients)
    epochs = [12]
    batch_sizes = [50]

    # clients = [Client(f"C-{i}") for i in range(params["n_clients"])]

    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    address = ("0.0.0.0", 6969)
    server_socket.bind(address)
    server_socket.listen(100)


    server = None
    thread = None
    
    try:
      clients = []
      print(f"Waiting for {max_clients} clients...")
      while len(clients) < max_clients: 
        client_socket, client_address = server_socket.accept()
        print(f"Client connected: {client_address}")
        client = Client(f"C-{len(clients)}", client_socket, client_address)
        client.handshake()
        clients.append(client)
        print(f"Client {client.name} handshake completed, remaining {max_clients - len(clients)}")  

     
      server = None
      for modlat in model_commit_latency:
        for epoch in epochs:
          for batch_size in batch_sizes:
              for n in n_clients:
                  server = start_server(params)
                  time.sleep(1)

                  params["model_commit_latency"] = modlat

                  params["lat_max"] = max(modlat)
                  c_distr = params["model_retrive_distribution"]
                  params["lat_avg"] = sum([c_distr[i] * modlat[i] for i in range(len(c_distr))])
                  params["lat_vare"] = sum([c_distr[i] * (modlat[i] - params["lat_avg"])**2 for i in range(len(c_distr))])/n

                  print(f"lat_arr: {modlat}")
                  print(f"lat_max: {params['lat_max']}, lat_avg: {params['lat_avg']}, lat_vare: {params['lat_vare']}")
                  
                  params["batch_size"] = batch_size
                  params["epochs"] = epoch
                  params["n_clients"] = n
                  params["test_name"] = f"test_ep{epoch}_bs{batch_size}_cl{n}_lat{modlat[0]}_{0}_{modlat[1]}"

                  # modulate knolodge amount (1/n)
                  # params["knolodge_amount"] = n/10
                  params["knolodge_amount"] = 1

                  start_simulation(params, server, clients)

                  print("Server terminated")
                  s1 = server
                  server = None
                  s1.terminate()
                  s1.wait()

    except KeyboardInterrupt:
      print("Ctrl+C detected, terminating...")
    except Exception as e:
      print(f"An error occurred: {e}")
    finally:
      server_socket.close()
      print("Server terminated")

