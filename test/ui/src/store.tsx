import { useEffect, useState } from "react";
import { TestData } from "./types";
import { Datasets } from "./datasets";
import { PatternGenerator } from "./patterns";

class Store {
  private listeners: { [key: string]: (() => void)[] } = {};
  private tests: { [key: string]: TestData } = {};

  constructor() {}

  getTests() {
    return Object.keys(this.tests);
  }

  addTest(test: TestData) {
    this.tests[test.name] = test;
    this.notify("tests");
  }

  removeTest(key: string) {
    delete this.tests[key];
    this.notify("tests");
  }

  cloneTest(key: string) {
    const test = this.tests[key];
    this.addTest({ ...test, name: `${test.name} (copy)` });
  }

  getTest(key: string) {
    return this.tests[key];
  }

  changeTest<F extends keyof TestData>(
    key: string,
    field: F,
    value: TestData[F]
  ) {
    this.tests[key][field] = value;

    if (field === "n_clients") {
      const ds_key = this.tests[key].dataset.name as keyof typeof Datasets;
      this.changeDatasetField(
        key,
        "indices",
        Datasets[ds_key].getDefaultIndices(this.tests[key].n_clients)
      );
    }

    this.notify(`test.${key}`);
    this.notify(`test.${key}.${field}`);
  }

  changeDatasetField<F extends keyof TestData["dataset"]>(
    key: string,
    field: F,
    value: TestData["dataset"][F]
  ) {
    this.tests[key].dataset[field] = value;

    if (field === "name") {
      const ds_key = value as keyof typeof Datasets;
      this.changeDatasetField(
        key,
        "indices",
        Datasets[ds_key].getDefaultIndices(this.tests[key].n_clients)
      );
    }

    this.notify(`test.${key}`);
    this.notify(`test.${key}.dataset`);
    this.notify(`test.${key}.dataset.${field}`);
  }

  getTimelines(key: string) {
    return this.tests[key].timelines.map((_, i) => i);
  }

  getTimeline(key: string, timeline: number) {
    return this.tests[key].timelines[timeline];
  }

  addTimeline(key: string) {
    this.tests[key].timelines.push({
      name: "New Pattern",
      timeline: {},
      rounds: 10,
    });

    this.notify(`test.${key}`);
    this.notify(`test.${key}.timelines`);
  }

  applyPatternToTimeline(
    key: string,
    timeline: number,
    pattern: PatternGenerator
  ) {
    const n_clients = this.tests[key].n_clients;

    const newTimeline = pattern(n_clients);

    this.tests[key].timelines[timeline] = {
      ...this.tests[key].timelines[timeline],
      ...newTimeline,
    };

    this.notify(`test.${key}.timelines.${timeline}`);
    this.notify(`test.${key}.timelines.${timeline}.timeline`);
    this.notify(`test.${key}.timelines.${timeline}.rounds`);
  }

  removeTimeline(key: string, timeline: number) {
    this.tests[key].timelines.splice(timeline, 1);

    this.notify(`test.${key}`);
    this.notify(`test.${key}.timelines`);
  }

  changeTimelineField<F extends keyof TestData["timelines"][number]>(
    key: string,
    timeline: number,
    field: F,
    value: TestData["timelines"][number][F]
  ) {
    this.tests[key].timelines[timeline][field] = value;

    this.notify(`test.${key}.timelines.${timeline}`);
    this.notify(`test.${key}.timelines.${timeline}.${field}`);
  }

  notify(key: string) {
    this.listeners[key]?.forEach((l) => l());
    this.listeners["*"]?.forEach((l) => l());
  }

  persist() {
    window.localStorage.setItem("tests", JSON.stringify(this.tests));
  }

  load() {
    const data = window.localStorage.getItem("tests");
    if (!data) return;
    this.tests = JSON.parse(data);
    this.notify("tests");
  }

  addEventListener(key: string, listener: () => void) {
    if (!this.listeners[key]) {
      this.listeners[key] = [];
    }
    this.listeners[key].push(listener);

    return () => {
      this.removeEventListener(key, listener);
    };
  }

  removeEventListener(key: string, listener: () => void) {
    if (!this.listeners[key]) {
      return;
    }
    this.listeners[key] = this.listeners[key].filter((l) => l !== listener);
  }
}

export const store = new Store();

export function useTestList() {
  const [tests, setTests] = useState<string[]>(store.getTests());

  useEffect(() => {
    const listener = store.addEventListener("tests", () => {
      setTests(store.getTests());
    });

    return () => {
      listener();
    };
  }, []);

  return tests;
}

export function useTestData(key: string) {
  const [data, setData] = useState({ ref: store.getTest(key) });

  useEffect(() => {
    const listener = store.addEventListener(`test.${key}`, () => {
      setData({ ref: store.getTest(key) });
    });

    return () => {
      listener();
    };
  }, [key]);

  return data.ref;
}

export function useTestField<F extends keyof TestData>(
  key: string,
  field: F
): TestData[F] {
  const [data, setData] = useState({ ref: store.getTest(key)[field] });

  useEffect(() => {
    const listener = store.addEventListener(`test.${key}.${field}`, () => {
      // setData(store.getTest(key)[field]);
      setData({ ref: store.getTest(key)[field] });
    });

    return () => {
      listener();
    };
  }, [key, field]);

  return data.ref;
}

export function useTestTimelines(key: string) {
  const [timelines, setTimelines] = useState(store.getTimelines(key));

  useEffect(() => {
    const listener = store.addEventListener(`test.${key}`, () => {
      setTimelines(store.getTimelines(key));
    });

    return () => {
      listener();
    };
  }, [key]);

  return timelines;
}

export function useTimeline(test: string, timeline: number) {
  const [data, setData] = useState({
    ref: store.getTimeline(test, timeline),
  });

  useEffect(() => {
    const listener = store.addEventListener(
      `test.${test}.timelines.${timeline}`,
      () => {
        setData({ ref: store.getTimeline(test, timeline) });
      }
    );

    return () => {
      listener();
    };
  }, [test, timeline]);

  return data.ref;
}
