import { Timeline } from "./types";

export type PatternGenerator = (clients: number) => {
  rounds: number;
  timeline: Timeline;
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type PatternFactory = (...args: any[]) => PatternGenerator;

export type Pattern = {
  name: string;
  factory: PatternFactory;
  args: {
    name: string;
    type: "number";
  }[];
};

function evenLatencyPattern(prc: number, latency: number): PatternGenerator {
  return (clients: number) => {
    const c_late = Math.floor(clients * prc);

    // lat 0 => period 2
    // lat 1 => period 3
    // lat 2 => period 5
    // lat 3 => period 7
    // lat x => period 2x + 1

    const period = 2 * latency + 1;

    const timeline: Timeline = {};

    for (let i = 0; i < period; i++) {
      timeline[i] = [];
      for (let j = 0; j < clients; j++) {
        if (j < c_late) {
          if (i == 0) timeline[i].push({ type: "RETR", client: j });
          if (i == period - 1) timeline[i].push({ type: "COMM", client: j });
        } else {
          timeline[i].push({
            type: (i + 1) % 2 === 0 ? "COMM" : "RETR",
            client: j,
          });
        }
      }

      if ((i + 1) % 2 === 0) timeline[i].push({ type: "AGG", client: -1 });
    }

    return { rounds: period, timeline };
  };
}

export const Patterns: Pattern[] = [
  {
    name: "Even Latency",
    factory: evenLatencyPattern,
    args: [
      { name: "Percentage (0-1)", type: "number" },
      { name: "Latency (agg Rounds)", type: "number" },
    ],
  },
];
