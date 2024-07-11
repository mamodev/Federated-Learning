import { DatasetConfig } from "./datasets";

export type REvent = "COMM" | "RETR" | "AGG";

export interface RoundEvent {
  type: REvent;
  client: number;
}

export type Timeline = {
  [round: number]: RoundEvent[];
};

export interface TestData {
  name: string;
  dataset: DatasetConfig;
  n_clients: number;
  rounds: number;

  timelines: {
    name: string;
    timeline: Timeline;
    rounds: number;
  }[];
}
