import io
import torch
from torch import nn
import torch.optim as optim
import torch.utils.data as data
import torchvision.transforms as transforms
import torchvision
import torch.nn.functional as F

import numpy as np

def numpy_to_torch(data):
  file_sream = io.BytesIO(data)
  npDict = np.load(file_sream)
  torchDict = {}
  for key in npDict:
    torchDict[key] = torch.from_numpy(npDict[key])
  return torchDict

def torch_to_numpy(data):
  npDict = {}
  for key in data:
    npDict[key] = data[key].cpu().numpy()
  file_stream = io.BytesIO()
  np.savez(file_stream, **npDict)
  return file_stream.getvalue()


def get_dataset(dataset):
    if dataset == 'MNIST':
      return {
          'dataset': torchvision.datasets.MNIST,
          'transform': transforms.Compose([ 
              transforms.ToTensor(),
              transforms.Normalize((0.1307,), (0.3081,))
          ]),
          'input_channels': 1,
          'input_height': 28,
          'input_width': 28,
          'num_classes': 10
      }
    
    if dataset == 'CIFAR10':  
        return {
            'dataset': torchvision.datasets.CIFAR10,
            'transform': transforms.Compose([
                transforms.ToTensor(),
                transforms.Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5))
            ]),
            'input_channels': 3,
            'input_height': 32,
            'input_width': 32,
            'num_classes': 10
        }
    
    if dataset == 'FashionMNIST':
        return {
            'dataset': torchvision.datasets.FashionMNIST,
            'transform': transforms.Compose([
                transforms.ToTensor(),
                transforms.Normalize((0.5,), (0.5,))
            ]),
            'input_channels': 1,
            'input_height': 28,
            'input_width': 28,
            'num_classes': 10
        }

    raise ValueError(f"Unknown dataset: {dataset}")

def get_model(params):
    network = params['network']
    dataset = params['name'] 


    ds = get_dataset(dataset)
    num_classes = ds['num_classes']

    if network == 'SimpleNN':
        return UniversalModel(ds['dataset'], ds['transform'], SimpleNN(ds['input_channels'], ds['input_height'], ds['input_width'], num_classes), params)
    
    if network == 'SimpleCNN':
        return UniversalModel(ds['dataset'], ds['transform'], SimpleCNN(ds['input_channels'], ds['input_height'], ds['input_width'], num_classes), params)

    raise ValueError(f"Unknown network: {network}")

class SimpleNN(nn.Module):
  def __init__(self, input_channels, input_height, input_width, num_classes):
      super(SimpleNN, self).__init__()
      input_size = input_channels * input_height * input_width
      self.layer1 = nn.Linear(input_size, 128)
      self.layer2 = nn.Linear(128, 64)
      self.layer3 = nn.Linear(64, num_classes)

  def forward(self, x):
      x = x.view(x.size(0), -1)  # Flatten the input tensor

      x = torch.relu(self.layer1(x))
      x = torch.relu(self.layer2(x))
      x = self.layer3(x)
      return x

class SimpleCNN(nn.Module):
    def __init__(self, input_channels, input_height, input_width, num_classes):
        super(SimpleCNN, self).__init__()
        self.conv1 = nn.Conv2d(input_channels, 32, kernel_size=3, stride=1, padding=1)
        self.conv2 = nn.Conv2d(32, 64, kernel_size=3, stride=1, padding=1)
        self.pool = nn.MaxPool2d(kernel_size=2, stride=2, padding=0)
        self.dropout = nn.Dropout(0.5)
        
        # Compute the size of the feature map after the convolution and pooling layers
        self.feature_size = self._get_conv_output_size(input_channels, input_height, input_width)
        
        self.fc1 = nn.Linear(self.feature_size, 128)
        self.fc2 = nn.Linear(128, num_classes)

    def _get_conv_output_size(self, input_channels, height, width):
        # Pass a dummy tensor through the conv and pool layers to get the output size
        dummy_input = torch.zeros(1, input_channels, height, width)
        output = self.pool(torch.relu(self.conv1(dummy_input)))
        output = self.pool(torch.relu(self.conv2(output)))
        output_size = output.view(-1).size(0)
        return output_size

    def forward(self, x):
        x = self.pool(torch.relu(self.conv1(x)))
        x = self.pool(torch.relu(self.conv2(x)))
        x = x.view(x.size(0), -1)  # Flatten the tensor
        x = torch.relu(self.fc1(x))
        x = self.dropout(x)
        x = self.fc2(x)
        return x
    

class UniversalModel:
    def __init__(self, dataset, transform, network, params):
        if 'use_accelerator' in params and params['use_accelerator']:
          self.device = torch.device('cuda' if torch.cuda.is_available() else 'mps' if torch.backends.mps.is_available() else 'cpu')
        else:
          self.device = torch.device('cpu')

        self.dataset = dataset
        self.network = network
        self.transform = transform
        self.def_weights = network.state_dict()
        self.params = params

    def train_from(self, weights):
        weights = numpy_to_torch(weights)
        self.__ensure_train_set()

        net = self.network
        optimizer = optim.SGD(net.parameters(), lr=self.params['learning_rate'], momentum=self.params['momentum'])
        # optimizer = optim.Adam(net.parameters(), lr=self.params['learning_rate'])


        net.train()
        if weights is not None:
          net.load_state_dict(weights)
        
        net.to(self.device)
        loss_function = nn.CrossEntropyLoss().to(self.device)

        for i in range(self.params['epochs']):
          # print("Epoch " + str(i))
          # for data, target in self.train_loader:
          for i, (data, target) in enumerate(self.train_loader):
              data = data.to(self.device)
              target = target.to(self.device)

              optimizer.zero_grad()
              output = net(data)
              loss = loss_function(output, target)
              loss.backward()
              optimizer.step()
        
        return torch_to_numpy(net.state_dict())
    
    def evaluate(self, weights):
        weights = numpy_to_torch(weights)
        self.__ensure_test_set()

        net = self.network
        net.load_state_dict(weights)
        net.to(self.device)
        net.eval()

        correct = 0
        total = 0
        with torch.no_grad():
            for data, target in self.test_loader:
                data = data.to(self.device)
                target = target.to(self.device)
                outputs = net(data)
                _, predicted = torch.max(outputs.data, 1)
                total += target.size(0)
                correct += (predicted == target).sum().item()

        return {
            'total': total,
            'correct': correct,
            'accuracy': correct / total
        }
    

    def get_def_weights(self):
        return torch_to_numpy(self.def_weights)
    
    def preload_test_set(self):
       self.__ensure_train_set()

    def __ensure_test_set(self):
        if hasattr(self, 'test_loader'):
          return
        
        test_dataset = self.dataset(root='data', train=False, transform=self.transform, download=True)
        self.test_loader = data.DataLoader(test_dataset, batch_size=self.params['batch_size'], shuffle=self.params['shuffle'])
        
    def __ensure_train_set(self):
        if hasattr(self, 'train_loader'):
          return
        
        train_dataset = self.dataset(root='data', train=True, transform=self.transform, download=True)

        if 'indices' not in self.params:
          self.train_dataset = train_dataset
          self.train_loader = data.DataLoader(train_dataset, batch_size=self.params['batch_size'], shuffle=self.params['shuffle'], num_workers=4, pin_memory=True)
          return
        
        class_subs = {}
        class_idxs = self.params['indices']

        class_indices = {int(c): [] for c in class_idxs.keys()}
        for i, (_, target) in enumerate(train_dataset):
            c_val = int(target)
            if c_val in class_indices and len(class_indices[c_val]) < class_idxs[str(c_val)][1]:
                class_indices[c_val].append(i)
            
            # Check if we've collected enough indices for all classes
            if all(len(class_indices[int(c)]) >= class_idxs[c][1] for c in class_idxs):
                break

        # Filter and create subsets based on the specified ranges
        for c in class_idxs.keys():
            c_val = int(c)
            range_start, range_end = class_idxs[c]
            filtered_indices = class_indices[c_val][range_start:range_end]
            class_subs[c] = data.Subset(train_dataset, filtered_indices)


        self.train_dataset = data.ConcatDataset([class_subs[c] for c in class_subs.keys()])
        self.train_loader = data.DataLoader(self.train_dataset, batch_size=self.params['batch_size'], shuffle=self.params['shuffle'])


if __name__ == '__main__':
  params = {
    'name': 'MNIST',
    'network': 'SimpleCNN',
    'learning_rate': 0.01,
    'momentum': 0.9,
    'batch_size': 256,
    'shuffle': True,
    'epochs': 1,
    'use_accelerator': True,
    'indices': {
        '0': [0, 1000],
        '1': [0, 1000],
        '2': [0, 1000],
        '3': [0, 1000],
        '4': [0, 1000],
        '5': [0, 1000],
        '6': [0, 1000],
        '7': [0, 1000],
        '8': [0, 1000],
        '9': [0, 1000]
    }
  }
  
  datasets = ['CIFAR10', 'FashionMNIST'] 

  for dataset in datasets:
    print(f"Dataset: {dataset}")
    params['name'] = dataset
    model = get_model(params) 
    model.preload_test_set()
    weights = model.get_def_weights()

    print(model.evaluate(weights))

    for i in range(10):
      weights = model.train_from(weights)
      print(model.evaluate(weights))


    # for dataset in datasets:
    #   print(f"Dataset: {dataset}")
    #   ds = get_dataset(dataset)
    #   train_dataset = ds['dataset'](root='data', train=True, transform=ds['transform'], download=True)

    #   class_counts = {i: 0 for i in range(ds['num_classes'])}
    #   for _, target in train_dataset:
    #     class_counts[int(target)] += 1

    #   print(f"Number of classes: {ds['num_classes']}")
    #   print(f"Class counts: {class_counts}")
    #   print(f"Total samples: {len(train_dataset)}")
    

