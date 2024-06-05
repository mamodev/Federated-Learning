import torch
import torch.utils.data
import torchvision
from network import Net

from fed.model import dict_to_params
from fed.http import HttpFedClient
from fed.py_torch import PyTorchModel, torch_to_numpy, numpy_to_torch
from fed.store import MemoryStore 

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

test_loader = torch.utils.data.DataLoader(test_dataset, batch_size=250, shuffle=False)

def test_loader_fact():
    return test_loader

params = dict_to_params({"type": "oracle"})
print("Oracle client params: ", params)



model = PyTorchModel(net, optimizer, test_loader_fact, test_loader_fact)


# rawModel = torch_to_numpy(net.state_dict())
# open("../model/original.npz", "wb").write(rawModel)

# content = open("../model/original.npz", "rb").read()
# net.load_state_dict(numpy_to_torch(content))
# rawModel = torch_to_numpy(net.state_dict())
# model.evaluate(rawModel, {})
# model.evaluate(rawModel, {})
# model.evaluate(rawModel, {})
# model.evaluate(rawModel, {})

client = HttpFedClient("http://localhost:8080", store, model, params)

print("Subscribing...")
client.subscribe()
client.unsubscribe()
print("Unsubscribed")
