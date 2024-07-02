import { TestData } from "./types";

type Config = Omit<TestData, "timeline"> & {
  timeline: [string, number][][];
};

export function toConfig(data: TestData): Config {
  const sorted_entries = Object.entries(data.timeline).sort(
    ([a], [b]) => parseInt(a) - parseInt(b)
  );

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const timeline = sorted_entries.map(([_, events]) => {
    const evt = events.map(
      ({ type, client }) => [type, client] as [string, number]
    );

    // event sort order RETR, COMM, AGG
    evt.sort(([a], [b]) => {
      if (a === "RETR") return -1;
      if (b === "RETR") return 1;
      if (a === "COMM") return -1;
      if (b === "COMM") return 1;
      return 0;
    });

    return evt;
  });

  return {
    ...data,
    timeline,
  };
}
