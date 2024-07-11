// a loading function is a function that takes the current round and the total number of rounds

import { Dataset, DSClientsIndices } from "./datasets";

// and returns a number between 0 and 1 representing the percentage of the training dataset that can be loaded
type LoadingFunction = (round: number, n_rounds: number) => number;

const linearLoading: LoadingFunction = (round, n_rounds) =>
  (round + 1) / n_rounds;
const quadraticLoading: LoadingFunction = (round, n_rounds) =>
  ((round + 1) / n_rounds) ** 2;
const cubicLoading: LoadingFunction = (round, n_rounds) =>
  ((round + 1) / n_rounds) ** 3;
const logaritmicLoading: LoadingFunction = (round, n_rounds) =>
  Math.log(round + 1) / Math.log(n_rounds);

const constantLoading: LoadingFunction = () => 1;

const randomLoading: LoadingFunction = () => Math.random();

const addBase = (fn: LoadingFunction, base: number) => {
  return (round: number, n_rounds: number) => {
    return (fn(round, n_rounds) + base) / (1 + base);
  };
};

export const LoadingFunctions = {
  linear: addBase(linearLoading, 0.1),
  quadratic: addBase(quadraticLoading, 0.1),
  cubic: addBase(cubicLoading, 0.1),
  logaritmic: addBase(logaritmicLoading, 0.1),
  constant: addBase(constantLoading, 0.1),
  random: addBase(randomLoading, 0.1),
};

export type LoadingFunctionName = keyof typeof LoadingFunctions;

export function getLoadingSerie(
  n_rounds: number,
  fn_name: LoadingFunctionName
) {
  const fn = LoadingFunctions[fn_name];
  return Array.from({ length: n_rounds }, (_, i) => fn(i, n_rounds));
}

//DISTRIBUTION

export type DistributionFunction = (
  classes: number,
  bias: number,
  base: number,
  i: number
) => number;

function normalDistr(classes: number, bias: number, base: number, i: number) {
  const p = Math.exp(-((i - classes / 2) ** 2) * 3 * bias);
  return (p + base) / (1 + base);
}

function linearDistr(classes: number, bias: number, base: number, i: number) {
  const p = Math.max(
    1 - i / (1 - (bias == 1 ? 0.9999 : bias)) / (classes - 1),
    0
  );
  return (p + base) / (1 + base);
}

export const DistributionFunctions = {
  normal: normalDistr,
  linear: linearDistr,
};

export type DistributionFunctionName = keyof typeof DistributionFunctions;

export function normalizeDistribution(
  DS: Dataset,
  fn_name: DistributionFunctionName,
  bias: number,
  base: number,
  offset = 0
) {
  const fn = DistributionFunctions[fn_name];
  const classes = DS.getNumClasses();
  const raw_distr = Array.from({ length: classes }, (_, i) => {
    return fn(classes, bias, base, (i + offset) % classes);
  });

  const sum = raw_distr.reduce((acc, v) => acc + v, 0);

  return raw_distr.map((v) => v / sum);
}

export function computeIndices(
  ds: Dataset,
  n_clients: number,
  distr: DistributionFunctionName,
  bias: number,
  base: number,
  knolwdgeAmount: number = 1 / n_clients
): DSClientsIndices {
  const indices: DSClientsIndices = {};
  const classes = ds.getNumClasses();

  const clientsDistr = Array.from({ length: n_clients }, (_, i) => {
    return normalizeDistribution(ds, distr, bias, base, i);
  });

  const total_pc = Array.from(
    { length: classes },
    (_, c) => {
      return clientsDistr.reduce((acc, v) => acc + v[c] * knolwdgeAmount, 0);
    },
    0
  );

  const computedDistr = clientsDistr.map((v) => {
    return v.map((v, c) => ds.classes[c][1] * knolwdgeAmount * classes * v);
  });

  console.log("total_pc", [...clientsDistr]);
  console.log("total_pc", [...total_pc]);

  const ooo = true;
  while (ooo) {
    const overflows = total_pc.findIndex((v) => v > 1);
    if (overflows === -1) {
      break;
    }

    const reducePc = 1 / total_pc[overflows];

    for (let i = 0; i < n_clients; i++) {
      computedDistr[i][overflows] *= reducePc;
    }

    total_pc[overflows] = 1;
  }

  for (let j = 0; j < ds.getNumClasses(); j++) {
    let start_index = 0;

    for (let i = 0; i < n_clients; i++) {
      indices[i] = indices[i] || {};
      indices[i][j] = indices[i][j] || [];

      const c_size = Math.floor(computedDistr[i][j]);
      const end_index = start_index + c_size;

      indices[i][j] = [start_index, end_index];

      start_index = end_index;
    }
  }

  return indices;
}
