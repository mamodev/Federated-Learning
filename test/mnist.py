from model import Model, numpy_to_torch, torch_to_numpy

import torch
import torch.nn as nn
import torch.optim as optim
import torch.utils.data as data
import torchvision.transforms as transforms
import torchvision
import torch.nn.functional as F

transform = transforms.Compose([transforms.ToTensor()])
    
class SimpleCNN(nn.Module):
    def __init__(self, num_classes=10):
        super(SimpleCNN, self).__init__()
        self.conv1 = nn.Conv2d(1, 32, kernel_size=3, stride=1, padding=1)
        self.conv2 = nn.Conv2d(32, 64, kernel_size=3, stride=1, padding=1)
        self.pool = nn.MaxPool2d(kernel_size=2, stride=2, padding=0)
        self.fc1 = nn.Linear(64 * 7 * 7, 128)
        self.fc2 = nn.Linear(128, num_classes)
        self.dropout = nn.Dropout(0.5)

    def forward(self, x):
        x = self.pool(torch.relu(self.conv1(x)))
        x = self.pool(torch.relu(self.conv2(x)))
        x = x.view(-1, 64 * 7 * 7)
        x = torch.relu(self.fc1(x))
        x = self.dropout(x)
        x = self.fc2(x)
        return x


class MNISTModel(Model):
    def __init__(self, params):
        self.device = torch.device('cuda' if torch.cuda.is_available() else 'mps' if torch.backends.mps.is_available() else 'cpu')
        print(f"Using device {self.device}")
        self.params = params

    def train_from(self, weights):
        weights = numpy_to_torch(weights)
        self.__ensure_train_set()

        net = SimpleCNN(self.NUM_CLASSES)
        optimizer = optim.SGD(net.parameters(), lr=self.params['learning_rate'], momentum=self.params['momentum'])

        net.train()
        if weights is not None:
          net.load_state_dict(weights)
        
        net.to(self.device)
        loss_function = nn.CrossEntropyLoss().to(self.device)

        for i in range(self.params['epochs']):
          print("Epoch " + str(i))
          # for data, target in self.train_loader:
          for i, (data, target) in enumerate(self.train_loader):
              # print(f"Batch {i}, data shape: {data.shape}")
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

        # net = SimpleNN(self.INPUT_SIZE, 500, self.NUM_CLASSES)
        net = SimpleCNN(self.NUM_CLASSES)
        net.load_state_dict(weights)
        net.eval()

        correct = 0
        total = 0
        
        net.to(self.device)
        
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
        self.__ensure_test_set()
        net = SimpleCNN(self.NUM_CLASSES)
        state = net.state_dict()
        return torch_to_numpy(state)

    def preload_test_set(self):
       self.__ensure_train_set()

    def __ensure_train_set(self):
      train_dataset = torchvision.datasets.MNIST(root='data', train=True, transform=transforms.ToTensor(), download=True)

      self.NUM_CLASSES = len(train_dataset.classes)
      self.INPUT_SIZE = train_dataset[0][0].numel()

      class_subs = {}

      if 'indices' not in self.params:
        self.train_dataset = train_dataset
        self.train_loader = data.DataLoader(train_dataset, batch_size=self.params['batch_size'], shuffle=self.params['shuffle'])
        return

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

    def __ensure_test_set(self):
      self.test_dataset = torchvision.datasets.MNIST(root='data', train=False, transform=transform, download=True)

      self.NUM_CLASSES = len(self.test_dataset.classes)
      self.INPUT_SIZE = self.test_dataset[0][0].numel()

      self.test_loader = data.DataLoader(self.test_dataset, batch_size=self.params['batch_size'], shuffle=self.params['shuffle'])


if __name__ == "__main__":
  model = MNISTModel({
    'batch_size': 256,
    'shuffle': True,
    'learning_rate': 0.01,
    'momentum': 0.9,
    'epochs': 6,
    # 'indices': {
    #   # '0': [0, 1000],
    #   # '1': [0, 1000],
    #   # '2': [0, 1000],
    #   # '3': [0, 1000],
    #   # '4': [0, 1000],
    #   # '5': [0, 1000],
    #   # '6': [0, 1000],
    #   # '7': [0, 1000],
    #   # '8': [0, 1000],
    #   # '9': [0, 1000]
    # }
  })

  model.preload_test_set()
  print(model.INPUT_SIZE)
  print(model.NUM_CLASSES)

  weights = model.get_def_weights()
  print(model.evaluate(weights))

  for i in range(10):
    weights = model.train_from(weights)
    print(model.evaluate(weights))

  print("Done")
