export class Dataset {
  name: string;
  classes: [number, number][];

  constructor(name: string, classes: [number, number][]) {
    this.name = name;
    this.classes = classes;
  }

  getDefaultIndices(n_clients: number): DSClientsIndices {
    const indices: DSClientsIndices = {};
    const n_classes = this.classes.length;

    for (let j = 0; j < n_classes; j++) {
      const c_size = Math.floor(this.classes[j][1] / n_clients);
      let start_index = 0;

      for (let i = 0; i < n_clients; i++) {
        const end_index = start_index + c_size;

        indices[i] = indices[i] || {};
        indices[i][j] = indices[i][j] || [];
        indices[i][j] = [start_index, end_index];

        start_index = end_index;
      }
    }

    return indices;
  }
}

const MNIST = new Dataset("MNIST", [
  [0, 5923],
  [1, 6742],
  [2, 5958],
  [3, 6131],
  [4, 5842],
  [5, 5421],
  [6, 5918],
  [7, 6265],
  [8, 5851],
  [9, 5949],
]);

const CIFAR10 = new Dataset("CIFAR10", [
  [0, 5000],
  [1, 5000],
  [2, 5000],
  [3, 5000],
  [4, 5000],
  [5, 5000],
  [6, 5000],
  [7, 5000],
  [8, 5000],
  [9, 5000],
]);

const FashionMNIST = new Dataset("FashionMNIST", [
  [0, 6000],
  [1, 6000],
  [2, 6000],
  [3, 6000],
  [4, 6000],
  [5, 6000],
  [6, 6000],
  [7, 6000],
  [8, 6000],
  [9, 6000],
]);

export const Datasets = {
  MNIST,
  CIFAR10,
  FashionMNIST,
};

// Client => Class => [start, end]
type DsIndex = {
  [key: number]: [number, number];
};

type DSClientsIndices = {
  [key: number]: DsIndex;
};

export const Networks = ["SimpleCNN", "SimpleNN"];

export interface DatasetConfig {
  name: string;
  epochs: number;
  batch_size: number;
  learning_rate: number;
  momentum: number;
  shuffle: boolean;
  indices: DSClientsIndices;
  network: string;
}
