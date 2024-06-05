import sys
import torch
import torch.utils.data
import torchvision
from network import Net

from fed.http import HttpFedClient
from fed.py_torch import PyTorchModel
from fed.store import MemoryStore 
from fed.model import dict_to_params

if len(sys.argv) < 3:
    print("Usage: python main.py <client_name> <number_in_dataset> [...<number_in_dataset>]")
    sys.exit(1)

clientName = sys.argv[1]

numbers = []
for i in range(2, len(sys.argv)):
    numbers.append(int(sys.argv[i]))
numbers = list(dict.fromkeys(numbers))

print(f"Bias: {numbers}...")

store = MemoryStore()
store.store("group_token", "GROUP_TOKEN")

net = Net()
optimizer = torch.optim.SGD(net.parameters(), lr=0.01, momentum=0.9)

test_dataset = torchvision.datasets.MNIST(
    './data/', 
    train=False, 
    download=True, 
    transform=torchvision.transforms.Compose([
        torchvision.transforms.ToTensor(),
        torchvision.transforms.Normalize((0.1307,), (0.3081,))
    ])
)

train_dataset = torchvision.datasets.MNIST(
    './data/', 
    train=True, 
    download=True, 
    transform=torchvision.transforms.Compose([
        torchvision.transforms.ToTensor(),
        torchvision.transforms.Normalize((0.1307,), (0.3081,))
    ])
)

train_idx = []
test_idx = []

mask = [False] * 10
for i in numbers:
    mask[i] = True
for i in range(min(len(train_dataset), 500)):
    if mask[train_dataset[i][1]]:
        train_idx.append(i)

for i in range(min(len(test_dataset), 1200)):
    if mask[test_dataset[i][1]]:
        test_idx.append(i)

test_dataset = torch.utils.data.Subset(test_dataset, test_idx)
train_dataset = torch.utils.data.Subset(train_dataset, train_idx)

train_loader = torch.utils.data.DataLoader(train_dataset, batch_size=250, shuffle=False)
test_loader = torch.utils.data.DataLoader(test_dataset, batch_size=250, shuffle=False)

def test_loader_fact():
    return test_loader

def train_loader_fact():
    return train_loader


params = dict_to_params({"type": "worker"})

model = PyTorchModel(net, optimizer, test_loader_fact, train_loader_fact)

client = HttpFedClient("http://localhost:8080", store, model, params)

print("Subscribing...")
client.subscribe()
client.unsubscribe()
print("Unsubscribed")
