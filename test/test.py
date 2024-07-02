
from mnist import MNISTModel

model = MNISTModel({
  'lr': 0.01,
  'momentum': 0.5,
  'epochs': 3,
  'batch_size': 64,
  'shuffle': False,
})

weights = model.train_from(None)

print(model.evaluate(weights))